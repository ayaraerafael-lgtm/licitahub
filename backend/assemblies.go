package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type createAssemblyRequest struct {
	MatchID    string `json:"matchId"`
	InterestID string `json:"interestId"`
	Title      string `json:"title"`
	DueDate    string `json:"dueDate"`
}

type updateAssemblyTaskRequest struct {
	Title                string `json:"title"`
	Description          string `json:"description"`
	Status               string `json:"status"`
	Priority             string `json:"priority"`
	ResponsibleCompanyID string `json:"responsibleCompanyId"`
	ResponsibleUserID    string `json:"responsibleUserId"`
	DueDate              string `json:"dueDate"`
}

type updateAssemblyTaskStatusRequest struct {
	Status string `json:"status"`
}

type createAssemblyStageRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type createAssemblyTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type createAssemblyCommentRequest struct {
	Content string `json:"content"`
}

type createAssemblyEvidenceRequest struct {
	EvidenceType string `json:"evidenceType"`
	Title        string `json:"title"`
	ExternalURL  string `json:"externalUrl"`
	Note         string `json:"note"`
	FileDataURL  string `json:"fileDataUrl"`
	FileName     string `json:"fileName"`
	MimeType     string `json:"mimeType"`
}

func (s sessionUser) canCoordinateAssembly() bool {
	return s.CompanyID != "" && (s.RoleKey == "company_admin" || s.RoleKey == "commercial")
}

func (s sessionUser) canWorkOnAssembly() bool {
	return s.CompanyID != "" && s.RoleKey != "reader"
}

func (a *app) ensureAssemblyMigrations(ctx context.Context) error {
	_, err := a.runPSQL(ctx, assemblySchemaSQL)
	if err != nil {
		return fmt.Errorf("nao foi possivel preparar a Central de Montagem: %w", err)
	}
	_, err = a.runPSQL(ctx, defaultAssemblyTemplateSQL)
	if err != nil {
		return fmt.Errorf("nao foi possivel preparar o modelo padrao da Central de Montagem: %w", err)
	}
	return nil
}

func (a *app) handleAssemblies(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		matchID := strings.TrimSpace(r.URL.Query().Get("matchId"))
		interestID := strings.TrimSpace(r.URL.Query().Get("interestId"))
		if matchID == "" && interestID == "" {
			writeError(w, http.StatusBadRequest, "consorcio ou participacao individual obrigatorio")
			return
		}
		if matchID != "" {
			a.getAssemblyByMatch(w, r, matchID, session)
			return
		}
		a.getAssemblyByInterest(w, r, interestID, session)
	case http.MethodPost:
		a.createAssembly(w, r, session)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleMyAssemblyTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if err := a.refreshOccurredTenders(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar a situacao dos editais")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."dueAt" NULLS LAST, item."updatedAt" DESC), '[]'::json)
		FROM (
			SELECT
				task.id::text AS id,
				task.title,
				COALESCE(task.description, '') AS description,
				task.status,
				task.priority,
				task.due_at AS "dueAt",
				task.updated_at AS "updatedAt",
				stage.title AS "stageTitle",
				ba.id::text AS "assemblyId",
				COALESCE(ci.match_id::text, '') AS "matchId",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				COALESCE((SELECT count(*) FROM bid_assembly_task_comments comment WHERE comment.task_id = task.id AND comment.deleted_at IS NULL), 0)::int AS "commentCount",
				COALESCE((SELECT count(*) FROM bid_assembly_task_evidences evidence WHERE evidence.task_id = task.id), 0)::int AS "evidenceCount"
			FROM bid_assembly_tasks task
			JOIN bid_assembly_stages stage ON stage.id = task.stage_id
			JOIN bid_assemblies ba ON ba.id = stage.assembly_id AND ba.status NOT IN ('cancelled', 'paused')
			LEFT JOIN consortium_intentions ci ON ci.id = ba.consortium_intention_id
			JOIN tenders t ON t.id = ba.tender_id
			JOIN bid_assembly_participants participant ON participant.assembly_id = ba.id
				AND participant.company_id = %s::uuid
				AND participant.user_id IS NULL
				AND participant.status = 'active'
			WHERE task.responsible_user_id = %s::uuid
			  AND task.status <> 'not_applicable'
			  AND t.status = 'published'
		) item;
	`, sqlQuote(session.CompanyID), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar suas tarefas")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleAssemblyCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}
	if err := a.refreshOccurredTenders(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar a situacao dos editais")
		return
	}
	if err := a.ensureCompanyIndividualAssemblies(r.Context(), session); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel preparar as montagens individuais")
		return
	}
	if err := a.ensureCompanyConsortiumAssemblies(r.Context(), session); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel preparar as montagens dos consorcios")
		return
	}
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	if _, err := time.Parse("2006-01", month); err != nil {
		writeError(w, http.StatusBadRequest, "mes invalido")
		return
	}
	monthStart := month + "-01"
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."openingDate", item."tenderNumber"), '[]'::json)
		FROM (
			SELECT
				ba.id::text AS id,
				ba.assembly_type AS "assemblyType",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				t.opening_date AS "openingDate",
				COALESCE(lead.trade_name, '') AS "leadCompanyName",
				COALESCE((
					SELECT round(
						100 * sum(CASE WHEN task.status = 'completed' THEN task.weight ELSE 0 END) /
						NULLIF(sum(CASE WHEN task.status <> 'not_applicable' THEN task.weight ELSE 0 END), 0)
					)::int
					FROM bid_assembly_tasks task
					JOIN bid_assembly_stages stage ON stage.id = task.stage_id
					WHERE stage.assembly_id = ba.id
				), 0) AS progress,
				COALESCE((
					SELECT count(*)::int
					FROM bid_assembly_tasks task
					JOIN bid_assembly_stages stage ON stage.id = task.stage_id
					WHERE stage.assembly_id = ba.id AND task.status <> 'not_applicable'
				), 0) AS "taskCount"
			FROM bid_assemblies ba
			JOIN bid_assembly_participants participant ON participant.assembly_id = ba.id
				AND participant.company_id = %s::uuid
				AND participant.user_id IS NULL
				AND participant.status = 'active'
			JOIN tenders t ON t.id = ba.tender_id
			LEFT JOIN companies lead ON lead.id = ba.lead_company_id
			WHERE ba.status NOT IN ('cancelled', 'paused')
			  AND t.deleted_at IS NULL
			  AND t.status = 'published'
			  AND t.opening_date >= %s::date
			  AND t.opening_date < (%s::date + interval '1 month')
		) item;
	`, sqlQuote(session.CompanyID), sqlQuote(monthStart), sqlQuote(monthStart)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o calendario de montagens")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleAssemblyList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}
	if err := a.refreshOccurredTenders(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar a situacao dos editais")
		return
	}
	if err := a.ensureCompanyIndividualAssemblies(r.Context(), session); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel preparar as montagens individuais")
		return
	}
	if err := a.ensureCompanyConsortiumAssemblies(r.Context(), session); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel preparar as montagens dos consorcios")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."openingDate" NULLS LAST, item."updatedAt" DESC), '[]'::json)
		FROM (
			SELECT
				ba.id::text AS id,
				ba.assembly_type AS "assemblyType",
				ba.status,
				ba.updated_at AS "updatedAt",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				t.opening_date AS "openingDate",
				COALESCE(lead.trade_name, '') AS "leadCompanyName",
				COALESCE((
					SELECT round(
						100 * sum(CASE WHEN task.status = 'completed' THEN task.weight ELSE 0 END) /
						NULLIF(sum(CASE WHEN task.status <> 'not_applicable' THEN task.weight ELSE 0 END), 0)
					)::int
					FROM bid_assembly_tasks task
					JOIN bid_assembly_stages stage ON stage.id = task.stage_id
					WHERE stage.assembly_id = ba.id
				), 0) AS progress,
				COALESCE((
					SELECT count(*)::int
					FROM bid_assembly_tasks task
					JOIN bid_assembly_stages stage ON stage.id = task.stage_id
					WHERE stage.assembly_id = ba.id AND task.status NOT IN ('completed', 'not_applicable')
				), 0) AS "openTaskCount",
				COALESCE((
					SELECT json_agg(member_item ORDER BY member_item->>'companyName')
					FROM (
						SELECT json_build_object('companyName', company.trade_name) AS member_item
						FROM bid_assembly_participants member
						JOIN companies company ON company.id = member.company_id
						WHERE member.assembly_id = ba.id AND member.user_id IS NULL AND member.status = 'active'
					) listed_members
				), '[]'::json) AS members
			FROM bid_assemblies ba
			JOIN bid_assembly_participants participant ON participant.assembly_id = ba.id
				AND participant.company_id = %s::uuid
				AND participant.user_id IS NULL
				AND participant.status = 'active'
			JOIN tenders t ON t.id = ba.tender_id
			LEFT JOIN companies lead ON lead.id = ba.lead_company_id
			WHERE ba.status NOT IN ('cancelled', 'paused') AND t.deleted_at IS NULL AND t.status = 'published'
		) item;
	`, sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar as montagens")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleAssemblyByPath(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if err := a.refreshOccurredTenders(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar a situacao dos editais")
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/assemblies/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "montagem nao encontrada")
		return
	}
	assemblyID := parts[0]
	available, err := a.assemblyTenderIsPublished(r.Context(), assemblyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel verificar a situacao da montagem")
		return
	}
	if !available {
		writeError(w, http.StatusConflict, "esta montagem esta pausada porque o edital nao esta publicado")
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		a.getAssemblyByID(w, r, assemblyID, session)
		return
	}
	if len(parts) == 2 && parts[1] == "stages" && r.Method == http.MethodPost {
		a.createAssemblyStage(w, r, assemblyID, session)
		return
	}
	if len(parts) == 4 && parts[1] == "stages" && parts[3] == "tasks" && r.Method == http.MethodPost {
		a.createAssemblyTask(w, r, assemblyID, parts[2], session)
		return
	}
	if len(parts) >= 3 && parts[1] == "tasks" {
		taskID := parts[2]
		if len(parts) == 3 && r.Method == http.MethodPut {
			a.updateAssemblyTask(w, r, assemblyID, taskID, session)
			return
		}
		if len(parts) == 3 && r.Method == http.MethodDelete {
			a.deleteAssemblyTask(w, r, assemblyID, taskID, session)
			return
		}
		if len(parts) == 4 && parts[3] == "status" && r.Method == http.MethodPatch {
			a.updateAssemblyTaskStatus(w, r, assemblyID, taskID, session)
			return
		}
		if len(parts) == 4 && parts[3] == "comments" && r.Method == http.MethodPost {
			a.createAssemblyComment(w, r, assemblyID, taskID, session)
			return
		}
		if len(parts) == 4 && parts[3] == "evidences" && r.Method == http.MethodPost {
			a.createAssemblyEvidence(w, r, assemblyID, taskID, session)
			return
		}
	}
	writeError(w, http.StatusNotFound, "recurso da montagem nao encontrado")
}

func (a *app) assemblyTenderIsPublished(ctx context.Context, assemblyID string) (bool, error) {
	payload, err := a.queryJSON(ctx, fmt.Sprintf(`
		SELECT json_build_object('available', EXISTS (
			SELECT 1
			FROM bid_assemblies assembly
			JOIN tenders tender ON tender.id = assembly.tender_id
			WHERE assembly.id = %s::uuid
			  AND assembly.status NOT IN ('cancelled', 'paused')
			  AND tender.deleted_at IS NULL
			  AND tender.status = 'published'
		));
	`, sqlQuote(assemblyID)))
	if err != nil {
		return false, err
	}
	var result struct {
		Available bool `json:"available"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return false, err
	}
	return result.Available, nil
}

func (a *app) getAssemblyByMatch(w http.ResponseWriter, r *http.Request, matchID string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				m.id::text AS "matchId",
				ci.id::text AS "consortiumIntentionId",
				t.id::text AS "tenderId",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				t.opening_date AS "openingDate",
				ci.lead_company_id::text AS "leadCompanyId",
				lead.trade_name AS "leadCompanyName",
				COALESCE(ba.id::text, '') AS "assemblyId",
				(ba.id IS NOT NULL) AS "exists",
				(ci.lead_company_id = %s::uuid AND %s) AS "canCreate"
			FROM matches m
			JOIN consortium_intentions ci ON ci.match_id = m.id
			JOIN tenders t ON t.id = ci.tender_id
			JOIN companies lead ON lead.id = ci.lead_company_id
			JOIN consortium_members cm ON cm.consortium_intention_id = ci.id
				AND cm.company_id = %s::uuid AND cm.status = 'active'
			LEFT JOIN bid_assemblies ba ON ba.consortium_intention_id = ci.id AND ba.status <> 'cancelled'
			WHERE m.id = %s::uuid AND m.status = 'active'
			LIMIT 1
		) item;
	`, sqlQuote(session.CompanyID), sqlBool(session.canCoordinateAssembly()), sqlQuote(session.CompanyID), sqlQuote(matchID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel localizar a montagem")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "sua empresa nao participa deste consorcio ativo")
		return
	}
	var info struct {
		AssemblyID    string `json:"assemblyId"`
		LeadCompanyID string `json:"leadCompanyId"`
		Exists        bool   `json:"exists"`
	}
	if json.Unmarshal(payload, &info) == nil && info.Exists && info.AssemblyID != "" {
		a.getAssemblyByID(w, r, info.AssemblyID, session)
		return
	}
	if info.LeadCompanyID != "" {
		if err := a.ensureConsortiumAssembly(r.Context(), matchID, session); err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel preparar a montagem do consorcio")
			return
		}
		a.getAssemblyByMatch(w, r, matchID, session)
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) getAssemblyByInterest(w http.ResponseWriter, r *http.Request, interestID string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				''::text AS "matchId",
				''::text AS "consortiumIntentionId",
				ti.id::text AS "interestId",
				t.id::text AS "tenderId",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				t.opening_date AS "openingDate",
				%s::uuid::text AS "leadCompanyId",
				c.trade_name AS "leadCompanyName",
				'individual' AS "assemblyType",
				COALESCE(ba.id::text, '') AS "assemblyId",
				(ba.id IS NOT NULL) AS "exists",
				%s AS "canCreate"
			FROM tender_interests ti
			JOIN tenders t ON t.id = ti.tender_id
			JOIN companies c ON c.id = ti.company_id
			LEFT JOIN bid_assemblies ba ON ba.tender_id = ti.tender_id
				AND ba.owner_company_id = ti.company_id
				AND ba.assembly_type = 'individual'
				AND ba.status <> 'cancelled'
			WHERE ti.id = %s::uuid
			  AND ti.company_id = %s::uuid
			  AND ti.participation_mode IN ('individual', 'seeking_partners')
			  AND ti.status = 'published'
			  AND ti.deleted_at IS NULL
			  AND t.deleted_at IS NULL
			  AND t.status = 'published'
			LIMIT 1
		) item;
	`, sqlQuote(session.CompanyID), sqlBool(session.canCoordinateAssembly()), sqlQuote(interestID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel localizar a montagem individual")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "esta participacao individual nao esta disponivel para montagem")
		return
	}
	var info struct {
		AssemblyID string `json:"assemblyId"`
		Exists     bool   `json:"exists"`
	}
	if json.Unmarshal(payload, &info) == nil && info.Exists && info.AssemblyID != "" {
		a.getAssemblyByID(w, r, info.AssemblyID, session)
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createAssembly(w http.ResponseWriter, r *http.Request, session sessionUser) {
	if !session.canCoordinateAssembly() {
		writeError(w, http.StatusForbidden, "somente o administrador ou comercial da empresa lider pode iniciar a montagem")
		return
	}
	var req createAssemblyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.MatchID = strings.TrimSpace(req.MatchID)
	req.InterestID = strings.TrimSpace(req.InterestID)
	req.Title = strings.TrimSpace(req.Title)
	req.DueDate = strings.TrimSpace(req.DueDate)
	if req.MatchID == "" && req.InterestID == "" {
		writeError(w, http.StatusBadRequest, "consorcio ou participacao individual obrigatorio")
		return
	}
	if req.MatchID != "" && req.InterestID != "" {
		writeError(w, http.StatusBadRequest, "informe apenas um tipo de montagem")
		return
	}
	if req.InterestID != "" {
		a.createIndividualAssembly(w, r, session, req)
		return
	}
	dueDateSQL := "NULL"
	if req.DueDate != "" {
		openingDate, err := a.tenderOpeningDateForMatch(r.Context(), req.MatchID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel consultar a data do edital")
			return
		}
		if err := validateAssemblyDueDate(req.DueDate, openingDate, "prazo geral"); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		dueDateSQL = sqlQuote(req.DueDate) + "::date"
	}
	payload, err := a.queryJSON(r.Context(), buildCreateAssemblySQL(req, session, dueDateSQL, true))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel iniciar a montagem: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "defina a empresa lider antes de iniciar a montagem")
		return
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &created); err != nil || created.ID == "" {
		writeError(w, http.StatusInternalServerError, "montagem criada sem identificador")
		return
	}
	a.getAssemblyByID(w, r, created.ID, session)
}

func (a *app) createIndividualAssembly(w http.ResponseWriter, r *http.Request, session sessionUser, req createAssemblyRequest) {
	dueDateSQL := "selected.opening_date"
	if req.DueDate != "" {
		openingDate, err := a.tenderOpeningDateForInterest(r.Context(), req.InterestID, session.CompanyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel consultar a data do edital")
			return
		}
		if err := validateAssemblyDueDate(req.DueDate, openingDate, "prazo geral"); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		dueDateSQL = sqlQuote(req.DueDate) + "::date"
	}
	payload, err := a.queryJSON(r.Context(), buildCreateIndividualAssemblySQL(req, session, dueDateSQL))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel iniciar a montagem individual: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "a participacao individual nao esta disponivel para iniciar a montagem")
		return
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &created); err != nil || created.ID == "" {
		writeError(w, http.StatusInternalServerError, "montagem criada sem identificador")
		return
	}
	a.getAssemblyByID(w, r, created.ID, session)
}

func (a *app) ensureIndividualAssembly(ctx context.Context, interestID string, session sessionUser) error {
	payload, err := a.queryJSON(ctx, buildCreateIndividualAssemblySQL(createAssemblyRequest{InterestID: interestID}, session, "selected.opening_date"))
	if err != nil {
		return err
	}
	var created struct {
		ID string `json:"id"`
	}
	if strings.TrimSpace(string(payload)) == "null" || json.Unmarshal(payload, &created) != nil || created.ID == "" {
		return errors.New("a participacao individual nao esta disponivel para iniciar a montagem")
	}
	return nil
}

func (a *app) ensureConsortiumAssembly(ctx context.Context, matchID string, session sessionUser) error {
	payload, err := a.queryJSON(ctx, buildCreateAssemblySQL(createAssemblyRequest{MatchID: matchID}, session, "selected.opening_date", false))
	if err != nil {
		return err
	}
	var created struct {
		ID string `json:"id"`
	}
	if strings.TrimSpace(string(payload)) == "null" || json.Unmarshal(payload, &created) != nil || created.ID == "" {
		return errors.New("o consorcio nao esta disponivel para iniciar a montagem")
	}
	return nil
}

func (a *app) ensureCompanyIndividualAssemblies(ctx context.Context, session sessionUser) error {
	payload, err := a.queryJSON(ctx, fmt.Sprintf(`
		SELECT COALESCE(json_agg(item.id::text), '[]'::json)
		FROM (
			SELECT ti.id
			FROM tender_interests ti
			JOIN tenders t ON t.id = ti.tender_id
			WHERE ti.company_id = %s::uuid
			  AND ti.participation_mode = 'individual'
			  AND ti.status = 'published'
			  AND ti.deleted_at IS NULL
			  AND t.deleted_at IS NULL
			  AND NOT EXISTS (
				SELECT 1 FROM bid_assemblies ba
				WHERE ba.tender_id = ti.tender_id
				  AND ba.owner_company_id = ti.company_id
				  AND ba.assembly_type = 'individual'
				  AND ba.status <> 'cancelled'
			)
		) item;
	`, sqlQuote(session.CompanyID)))
	if err != nil {
		return err
	}
	var interestIDs []string
	if err := json.Unmarshal(payload, &interestIDs); err != nil {
		return err
	}
	for _, interestID := range interestIDs {
		if err := a.ensureIndividualAssembly(ctx, interestID, session); err != nil {
			return err
		}
	}
	return nil
}

func (a *app) ensureCompanyConsortiumAssemblies(ctx context.Context, session sessionUser) error {
	payload, err := a.queryJSON(ctx, fmt.Sprintf(`
		SELECT COALESCE(json_agg(item.match_id::text), '[]'::json)
		FROM (
			SELECT ci.match_id
			FROM consortium_intentions ci
			JOIN matches m ON m.id = ci.match_id AND m.status = 'active'
			JOIN tenders t ON t.id = ci.tender_id AND t.deleted_at IS NULL AND t.status = 'published'
			JOIN consortium_members member ON member.consortium_intention_id = ci.id
			WHERE member.company_id = %s::uuid
			  AND member.status = 'active'
			  AND ci.lead_company_id IS NOT NULL
			  AND NOT EXISTS (
				SELECT 1 FROM bid_assemblies ba
				WHERE ba.consortium_intention_id = ci.id AND ba.status <> 'cancelled'
			)
		) item;
	`, sqlQuote(session.CompanyID)))
	if err != nil {
		return err
	}
	var matchIDs []string
	if err := json.Unmarshal(payload, &matchIDs); err != nil {
		return err
	}
	for _, matchID := range matchIDs {
		if err := a.ensureConsortiumAssembly(ctx, matchID, session); err != nil {
			return err
		}
	}
	return nil
}

func (a *app) getAssemblyByID(w http.ResponseWriter, r *http.Request, assemblyID string, session sessionUser) {
	if err := a.syncAssemblyParticipants(r.Context(), assemblyID); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel sincronizar participantes da montagem")
		return
	}
	_ = a.createAssemblyDeadlineNotifications(r.Context(), assemblyID)
	payload, err := a.queryJSON(r.Context(), buildAssemblyDetailSQL(assemblyID, session))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar a montagem: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "montagem nao encontrada ou acesso nao permitido")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createAssemblyStage(w http.ResponseWriter, r *http.Request, assemblyID string, session sessionUser) {
	if !session.canCoordinateAssembly() {
		writeError(w, http.StatusForbidden, "seu perfil nao pode criar fases")
		return
	}
	var req createAssemblyStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "nome da fase obrigatorio")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH allowed AS (
			SELECT ba.id
			FROM bid_assemblies ba
			WHERE ba.id = %s::uuid AND ba.lead_company_id = %s::uuid AND ba.status <> 'cancelled'
		), inserted AS (
			INSERT INTO bid_assembly_stages (assembly_id, title, description, position, is_custom, created_by_user_id)
			SELECT allowed.id, %s, NULLIF(%s, ''),
				COALESCE((SELECT max(position) + 1 FROM bid_assembly_stages WHERE assembly_id = allowed.id), 1),
				true, %s::uuid
			FROM allowed
			RETURNING *
		)
		SELECT row_to_json(inserted) FROM inserted;
	`, sqlQuote(assemblyID), sqlQuote(session.CompanyID), sqlQuote(req.Title), sqlQuote(req.Description), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel criar a fase")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "apenas a empresa lider pode criar fases")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) createAssemblyTask(w http.ResponseWriter, r *http.Request, assemblyID, stageID string, session sessionUser) {
	if !session.canCoordinateAssembly() {
		writeError(w, http.StatusForbidden, "seu perfil nao pode criar tarefas")
		return
	}
	var req createAssemblyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "titulo da tarefa obrigatorio")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH allowed AS (
			SELECT bas.id
			FROM bid_assembly_stages bas
			JOIN bid_assemblies ba ON ba.id = bas.assembly_id
			WHERE ba.id = %s::uuid AND bas.id = %s::uuid
			  AND ba.lead_company_id = %s::uuid AND ba.status <> 'cancelled'
		), inserted AS (
			INSERT INTO bid_assembly_tasks (stage_id, title, description, position, is_custom, created_by_user_id)
			SELECT allowed.id, %s, NULLIF(%s, ''),
				COALESCE((SELECT max(position) + 1 FROM bid_assembly_tasks WHERE stage_id = allowed.id), 1),
				true, %s::uuid
			FROM allowed
			RETURNING *
		)
		SELECT row_to_json(inserted) FROM inserted;
	`, sqlQuote(assemblyID), sqlQuote(stageID), sqlQuote(session.CompanyID), sqlQuote(req.Title), sqlQuote(req.Description), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel criar a tarefa")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "apenas a empresa lider pode criar tarefas")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) updateAssemblyTask(w http.ResponseWriter, r *http.Request, assemblyID, taskID string, session sessionUser) {
	if !session.canWorkOnAssembly() {
		writeError(w, http.StatusForbidden, "seu perfil possui acesso somente para leitura")
		return
	}
	var req updateAssemblyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Status = normalizeAssemblyTaskStatus(req.Status)
	req.Priority = normalizeAssemblyTaskPriority(req.Priority)
	req.ResponsibleCompanyID = strings.TrimSpace(req.ResponsibleCompanyID)
	req.ResponsibleUserID = strings.TrimSpace(req.ResponsibleUserID)
	req.DueDate = strings.TrimSpace(req.DueDate)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "titulo da tarefa obrigatorio")
		return
	}
	dueAtSQL := "NULL"
	if req.DueDate != "" {
		openingDate, err := a.tenderOpeningDateForAssembly(r.Context(), assemblyID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel consultar a data do edital")
			return
		}
		if err := validateAssemblyDueDate(req.DueDate, openingDate, "prazo da tarefa"); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		dueAtSQL = sqlQuote(req.DueDate) + "::date"
	}
	payload, err := a.queryJSON(r.Context(), buildUpdateAssemblyTaskSQL(assemblyID, taskID, req, session, dueAtSQL))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar a tarefa: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "tarefa nao atribuida ao seu usuario ou empresa")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) deleteAssemblyTask(w http.ResponseWriter, r *http.Request, assemblyID, taskID string, session sessionUser) {
	if !session.canCoordinateAssembly() {
		writeError(w, http.StatusForbidden, "somente o administrador ou comercial da empresa lider pode excluir tarefas")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH allowed AS (
			SELECT task.id
			FROM bid_assembly_tasks task
			JOIN bid_assembly_stages stage ON stage.id = task.stage_id
			JOIN bid_assemblies assembly ON assembly.id = stage.assembly_id
			WHERE assembly.id = %s::uuid
			  AND task.id = %s::uuid
			  AND assembly.lead_company_id = %s::uuid
			  AND assembly.status <> 'cancelled'
		), deleted AS (
			DELETE FROM bid_assembly_tasks task
			USING allowed
			WHERE task.id = allowed.id
			RETURNING task.id::text AS id
		)
		SELECT row_to_json(deleted) FROM deleted;
	`, sqlQuote(assemblyID), sqlQuote(taskID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel excluir a tarefa")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "apenas a empresa lider pode excluir esta tarefa")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) updateAssemblyTaskStatus(w http.ResponseWriter, r *http.Request, assemblyID, taskID string, session sessionUser) {
	if !session.canWorkOnAssembly() {
		writeError(w, http.StatusForbidden, "seu perfil possui acesso somente para leitura")
		return
	}
	var req updateAssemblyTaskStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Status = normalizeAssemblyTaskStatus(req.Status)
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected AS (
			SELECT task.id, assembly.lead_company_id
			FROM bid_assembly_tasks task
			JOIN bid_assembly_stages stage ON stage.id = task.stage_id
			JOIN bid_assemblies assembly ON assembly.id = stage.assembly_id
			JOIN bid_assembly_participants access ON access.assembly_id = assembly.id
				AND access.company_id = %s::uuid AND access.user_id IS NULL AND access.status = 'active'
			WHERE assembly.id = %s::uuid AND task.id = %s::uuid AND assembly.status <> 'cancelled'
			  AND (assembly.lead_company_id = %s::uuid OR task.responsible_company_id = %s::uuid OR task.responsible_user_id = %s::uuid)
			LIMIT 1
		), updated AS (
			UPDATE bid_assembly_tasks task
			SET status = CASE WHEN selected.lead_company_id <> %s::uuid AND %s = 'completed' THEN 'under_review' ELSE %s END,
				submitted_at = CASE WHEN (CASE WHEN selected.lead_company_id <> %s::uuid AND %s = 'completed' THEN 'under_review' ELSE %s END) = 'under_review' THEN now() ELSE task.submitted_at END,
				completed_at = CASE WHEN selected.lead_company_id = %s::uuid AND %s = 'completed' THEN now() WHEN %s <> 'completed' THEN NULL ELSE task.completed_at END,
				updated_at = now()
			FROM selected
			WHERE task.id = selected.id
			RETURNING task.id::text AS id, task.status
		)
		SELECT row_to_json(updated) FROM updated;
	`, sqlQuote(session.CompanyID), sqlQuote(assemblyID), sqlQuote(taskID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.UserID),
		sqlQuote(session.CompanyID), sqlQuote(req.Status), sqlQuote(req.Status), sqlQuote(session.CompanyID), sqlQuote(req.Status), sqlQuote(req.Status), sqlQuote(session.CompanyID), sqlQuote(req.Status), sqlQuote(req.Status)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o andamento da tarefa")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "tarefa nao atribuida ao seu usuario ou empresa")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createAssemblyComment(w http.ResponseWriter, r *http.Request, assemblyID, taskID string, session sessionUser) {
	if !session.canWorkOnAssembly() {
		writeError(w, http.StatusForbidden, "seu perfil possui acesso somente para leitura")
		return
	}
	var req createAssemblyCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "comentario vazio")
		return
	}
	payload, err := a.queryJSON(r.Context(), buildCreateAssemblyCommentSQL(assemblyID, taskID, req.Content, session))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel comentar na tarefa")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "sua empresa nao participa desta montagem")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) createAssemblyEvidence(w http.ResponseWriter, r *http.Request, assemblyID, taskID string, session sessionUser) {
	if !session.canWorkOnAssembly() {
		writeError(w, http.StatusForbidden, "seu perfil possui acesso somente para leitura")
		return
	}
	var req createAssemblyEvidenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.EvidenceType = strings.ToLower(strings.TrimSpace(req.EvidenceType))
	req.Title = strings.TrimSpace(req.Title)
	req.ExternalURL = strings.TrimSpace(req.ExternalURL)
	req.Note = strings.TrimSpace(req.Note)
	req.FileDataURL = strings.TrimSpace(req.FileDataURL)
	req.FileName = strings.TrimSpace(req.FileName)
	req.MimeType = strings.TrimSpace(req.MimeType)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "titulo da evidencia obrigatorio")
		return
	}
	if req.EvidenceType != "file" && req.EvidenceType != "link" && req.EvidenceType != "note" {
		writeError(w, http.StatusBadRequest, "tipo de evidencia invalido")
		return
	}
	fileURL := ""
	fileSize := int64(0)
	if req.EvidenceType == "file" {
		if req.FileDataURL == "" || req.FileName == "" {
			writeError(w, http.StatusBadRequest, "selecione um arquivo")
			return
		}
		var err error
		fileURL, fileSize, err = saveAssemblyFile(req.FileDataURL, req.FileName)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if req.EvidenceType == "link" && req.ExternalURL == "" {
		writeError(w, http.StatusBadRequest, "informe o link da evidencia")
		return
	}
	if req.EvidenceType == "note" && req.Note == "" {
		writeError(w, http.StatusBadRequest, "informe a anotacao da evidencia")
		return
	}
	payload, err := a.queryJSON(r.Context(), buildCreateAssemblyEvidenceSQL(assemblyID, taskID, req, fileURL, fileSize, session))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel incluir a evidencia: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "sua empresa nao participa desta montagem")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) syncAssemblyParticipants(ctx context.Context, assemblyID string) error {
	_, err := a.runPSQL(ctx, fmt.Sprintf(`
		INSERT INTO bid_assembly_participants (assembly_id, company_id, role, status)
		SELECT ba.id, CASE WHEN ba.assembly_type = 'individual' THEN ba.owner_company_id ELSE cm.company_id END,
			CASE WHEN ba.assembly_type = 'individual' OR cm.company_id = ba.lead_company_id THEN 'coordinator' ELSE 'collaborator' END,
			'active'
		FROM bid_assemblies ba
		LEFT JOIN consortium_members cm ON cm.consortium_intention_id = ba.consortium_intention_id AND cm.status = 'active'
		WHERE ba.id = %s::uuid
		  AND (ba.assembly_type = 'individual' OR cm.id IS NOT NULL)
		ON CONFLICT DO NOTHING;
		UPDATE bid_assembly_participants bap
		SET role = CASE WHEN ba.assembly_type = 'individual' OR bap.company_id = ba.lead_company_id THEN 'coordinator' ELSE 'collaborator' END
		FROM bid_assemblies ba
		WHERE bap.assembly_id = ba.id
		  AND ba.id = %s::uuid
		  AND bap.user_id IS NULL
		  AND bap.status = 'active';
		UPDATE bid_assembly_participants bap
		SET status = 'removed', removed_at = now()
		FROM bid_assemblies ba
		WHERE bap.assembly_id = ba.id AND ba.id = %s::uuid AND ba.assembly_type = 'consortium' AND bap.user_id IS NULL
		  AND bap.status = 'active'
		  AND NOT EXISTS (
			SELECT 1 FROM consortium_members cm
			WHERE cm.consortium_intention_id = ba.consortium_intention_id
			  AND cm.company_id = bap.company_id AND cm.status = 'active'
		  );
	`, sqlQuote(assemblyID), sqlQuote(assemblyID), sqlQuote(assemblyID)))
	return err
}

func (a *app) createAssemblyDeadlineNotifications(ctx context.Context, assemblyID string) error {
	_, err := a.runPSQL(ctx, fmt.Sprintf(`
		WITH candidates AS (
			SELECT bat.id AS task_id, bat.responsible_user_id,
				CASE
					WHEN bat.due_at::date < CURRENT_DATE THEN 'overdue'
					WHEN bat.due_at::date = CURRENT_DATE THEN 'due_today'
					ELSE 'due_soon'
				END AS alert_type,
				bat.title, basm.id AS assembly_id
			FROM bid_assembly_tasks bat
			JOIN bid_assembly_stages stage ON stage.id = bat.stage_id
			JOIN bid_assemblies basm ON basm.id = stage.assembly_id
			JOIN tenders tender ON tender.id = basm.tender_id
			WHERE basm.id = %s::uuid
			  AND basm.status NOT IN ('cancelled', 'paused')
			  AND tender.status = 'published'
			  AND bat.responsible_user_id IS NOT NULL
			  AND bat.due_at IS NOT NULL
			  AND bat.status NOT IN ('completed', 'not_applicable')
			  AND bat.due_at::date <= CURRENT_DATE + 3
		), inserted_alerts AS (
			INSERT INTO bid_assembly_deadline_alerts (task_id, recipient_user_id, alert_type, alert_date)
			SELECT task_id, responsible_user_id, alert_type, CURRENT_DATE FROM candidates
			ON CONFLICT DO NOTHING
			RETURNING task_id, recipient_user_id, alert_type
		)
		INSERT INTO notifications (
			recipient_user_id, recipient_company_id, type, title, message,
			destination_screen, related_entity_type, related_entity_id
		)
		SELECT ia.recipient_user_id, u.company_id, 'system',
			CASE ia.alert_type WHEN 'overdue' THEN 'Tarefa em atraso' WHEN 'due_today' THEN 'Tarefa vence hoje' ELSE 'Prazo de tarefa proximo' END,
			c.title, 'assembly-board', 'bid_assembly', c.assembly_id
		FROM inserted_alerts ia
		JOIN candidates c ON c.task_id = ia.task_id AND c.responsible_user_id = ia.recipient_user_id
		JOIN users u ON u.id = ia.recipient_user_id;
	`, sqlQuote(assemblyID)))
	return err
}

func (a *app) tenderOpeningDateForMatch(ctx context.Context, matchID string) (string, error) {
	payload, err := a.queryJSON(ctx, fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT t.opening_date::date::text AS "openingDate"
			FROM matches m
			JOIN consortium_intentions ci ON ci.match_id = m.id
			JOIN tenders t ON t.id = ci.tender_id
			WHERE m.id = %s::uuid
			LIMIT 1
		) item;
	`, sqlQuote(matchID)))
	if err != nil {
		return "", err
	}
	return decodeAssemblyOpeningDate(payload)
}

func (a *app) tenderOpeningDateForInterest(ctx context.Context, interestID, companyID string) (string, error) {
	payload, err := a.queryJSON(ctx, fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT t.opening_date::date::text AS "openingDate"
			FROM tender_interests ti
			JOIN tenders t ON t.id = ti.tender_id
			WHERE ti.id = %s::uuid AND ti.company_id = %s::uuid
			LIMIT 1
		) item;
	`, sqlQuote(interestID), sqlQuote(companyID)))
	if err != nil {
		return "", err
	}
	return decodeAssemblyOpeningDate(payload)
}

func (a *app) tenderOpeningDateForAssembly(ctx context.Context, assemblyID string) (string, error) {
	payload, err := a.queryJSON(ctx, fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT t.opening_date::date::text AS "openingDate"
			FROM bid_assemblies ba
			JOIN tenders t ON t.id = ba.tender_id
			WHERE ba.id = %s::uuid
			LIMIT 1
		) item;
	`, sqlQuote(assemblyID)))
	if err != nil {
		return "", err
	}
	return decodeAssemblyOpeningDate(payload)
}

func decodeAssemblyOpeningDate(payload []byte) (string, error) {
	if strings.TrimSpace(string(payload)) == "null" {
		return "", errors.New("edital nao encontrado")
	}
	var result struct {
		OpeningDate string `json:"openingDate"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return "", err
	}
	return strings.TrimSpace(result.OpeningDate), nil
}

func validateAssemblyDueDate(value, openingDate, fieldName string) error {
	dueDate, err := time.Parse("2006-01-02", value)
	if err != nil {
		return fmt.Errorf("%s invalido", fieldName)
	}
	if strings.TrimSpace(openingDate) == "" {
		return errors.New("informe a data de abertura do edital antes de definir prazos da montagem")
	}
	opening, err := time.Parse("2006-01-02", openingDate)
	if err != nil {
		return errors.New("data de abertura do edital invalida")
	}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if dueDate.Before(today) {
		return fmt.Errorf("%s nao pode ser anterior a data atual", fieldName)
	}
	if dueDate.After(opening) {
		return fmt.Errorf("%s nao pode ser posterior a data de abertura do edital", fieldName)
	}
	return nil
}

func buildCreateAssemblySQL(req createAssemblyRequest, session sessionUser, dueDateSQL string, requireLeader bool) string {
	titleSQL := "selected.number || ' - ' || selected.agency"
	if req.Title != "" {
		titleSQL = sqlQuote(req.Title)
	}
	accessCondition := fmt.Sprintf(`
			  AND EXISTS (
				SELECT 1 FROM consortium_members cm
				WHERE cm.consortium_intention_id = ci.id AND cm.company_id = %s::uuid AND cm.status = 'active'
			  )`, sqlQuote(session.CompanyID))
	if requireLeader {
		accessCondition = fmt.Sprintf(`
			  AND ci.lead_company_id = %s::uuid
			  AND EXISTS (
				SELECT 1 FROM consortium_members cm
				WHERE cm.consortium_intention_id = ci.id AND cm.company_id = %s::uuid AND cm.status = 'active'
			  )`, sqlQuote(session.CompanyID), sqlQuote(session.CompanyID))
	}
	return fmt.Sprintf(`
		WITH selected AS (
			SELECT ci.*, t.number, t.agency, t.object, t.opening_date
			FROM matches m
			JOIN consortium_intentions ci ON ci.match_id = m.id
			JOIN tenders t ON t.id = ci.tender_id
			WHERE m.id = %s::uuid AND m.status = 'active'
			  AND t.deleted_at IS NULL
			  AND t.status = 'published'%s
			LIMIT 1
		), template AS (
			SELECT id FROM bid_assembly_templates WHERE is_system = true AND name = 'Modelo LicitaHub' AND is_active = true LIMIT 1
		), inserted_assembly AS (
			INSERT INTO bid_assemblies (
				consortium_intention_id, tender_id, template_id, lead_company_id,
				title, status, due_date, created_by_user_id
			)
			SELECT selected.id, selected.tender_id, template.id, selected.lead_company_id,
				%s, 'in_progress', %s, %s::uuid
			FROM selected CROSS JOIN template
			ON CONFLICT (consortium_intention_id) DO NOTHING
			RETURNING *
		), assembly AS (
			SELECT * FROM inserted_assembly
			UNION ALL
			SELECT existing.* FROM bid_assemblies existing JOIN selected ON selected.id = existing.consortium_intention_id
			LIMIT 1
		), inserted_participants AS (
			INSERT INTO bid_assembly_participants (assembly_id, company_id, role, status)
			SELECT assembly.id, cm.company_id,
				CASE WHEN cm.company_id = assembly.lead_company_id THEN 'coordinator' ELSE 'collaborator' END,
				'active'
			FROM assembly
			JOIN consortium_members cm ON cm.consortium_intention_id = assembly.consortium_intention_id AND cm.status = 'active'
			ON CONFLICT DO NOTHING
			RETURNING id
		), inserted_stages AS (
			INSERT INTO bid_assembly_stages (
				assembly_id, source_template_stage_id, title, description, position, created_by_user_id
			)
			SELECT assembly.id, template_stage.id, template_stage.title, template_stage.description,
				template_stage.position, %s::uuid
			FROM assembly
			JOIN bid_assembly_template_stages template_stage ON template_stage.template_id = assembly.template_id
			WHERE NOT EXISTS (SELECT 1 FROM bid_assembly_stages stage WHERE stage.assembly_id = assembly.id)
			RETURNING id, source_template_stage_id
		), inserted_tasks AS (
			INSERT INTO bid_assembly_tasks (
				stage_id, source_template_task_id, title, description, position, weight, created_by_user_id
			)
			SELECT stage.id, template_task.id, template_task.title, template_task.description,
				template_task.position, template_task.default_weight, %s::uuid
			FROM assembly
			JOIN inserted_stages stage ON true
			JOIN bid_assembly_template_tasks template_task ON template_task.template_stage_id = stage.source_template_stage_id
			WHERE NOT EXISTS (SELECT 1 FROM bid_assembly_tasks task WHERE task.stage_id = stage.id)
			RETURNING id
		), activity AS (
			INSERT INTO bid_assembly_activity_logs (assembly_id, actor_user_id, actor_company_id, action, description)
			SELECT assembly.id, %s::uuid, %s::uuid, 'assembly_created', 'Central de Montagem iniciada com o Modelo LicitaHub.'
			FROM assembly WHERE EXISTS (SELECT 1 FROM inserted_assembly)
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (SELECT assembly.id::text AS id FROM assembly) item;
	`, sqlQuote(req.MatchID), accessCondition, titleSQL, dueDateSQL,
		sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.CompanyID))
}

func buildCreateIndividualAssemblySQL(req createAssemblyRequest, session sessionUser, dueDateSQL string) string {
	titleSQL := "selected.number || ' - ' || selected.agency || ' | Participacao individual'"
	if req.Title != "" {
		titleSQL = sqlQuote(req.Title)
	}
	return fmt.Sprintf(`
		WITH selected AS (
			SELECT ti.tender_id, t.number, t.agency, t.opening_date
			FROM tender_interests ti
			JOIN tenders t ON t.id = ti.tender_id
			WHERE ti.id = %s::uuid AND ti.company_id = %s::uuid
			  AND ti.participation_mode IN ('individual', 'seeking_partners') AND ti.status = 'published'
			  AND ti.deleted_at IS NULL AND t.deleted_at IS NULL AND t.status = 'published'
			LIMIT 1
		), template AS (
			SELECT id FROM bid_assembly_templates WHERE is_system = true AND name = 'Modelo LicitaHub' AND is_active = true LIMIT 1
		), inserted_assembly AS (
			INSERT INTO bid_assemblies (
				consortium_intention_id, tender_id, template_id, assembly_type, owner_company_id,
				lead_company_id, title, status, due_date, created_by_user_id
			)
			SELECT NULL, selected.tender_id, template.id, 'individual', %s::uuid,
				%s::uuid, %s, 'in_progress', %s, %s::uuid
			FROM selected CROSS JOIN template
			ON CONFLICT DO NOTHING
			RETURNING *
		), assembly AS (
			SELECT * FROM inserted_assembly
			UNION ALL
			SELECT existing.* FROM bid_assemblies existing
			JOIN selected ON selected.tender_id = existing.tender_id
			WHERE existing.owner_company_id = %s::uuid AND existing.assembly_type = 'individual' AND existing.status <> 'cancelled'
			LIMIT 1
		), inserted_participant AS (
			INSERT INTO bid_assembly_participants (assembly_id, company_id, role, status)
			SELECT assembly.id, %s::uuid, 'coordinator', 'active' FROM assembly
			ON CONFLICT DO NOTHING
			RETURNING id
		), inserted_stages AS (
			INSERT INTO bid_assembly_stages (assembly_id, source_template_stage_id, title, description, position, created_by_user_id)
			SELECT assembly.id, template_stage.id, template_stage.title, template_stage.description, template_stage.position, %s::uuid
			FROM assembly
			JOIN bid_assembly_template_stages template_stage ON template_stage.template_id = assembly.template_id
			WHERE NOT EXISTS (SELECT 1 FROM bid_assembly_stages stage WHERE stage.assembly_id = assembly.id)
			RETURNING id, source_template_stage_id
		), inserted_tasks AS (
			INSERT INTO bid_assembly_tasks (stage_id, source_template_task_id, title, description, position, weight, created_by_user_id)
			SELECT stage.id, template_task.id, template_task.title, template_task.description, template_task.position, template_task.default_weight, %s::uuid
			FROM inserted_stages stage
			JOIN bid_assembly_template_tasks template_task ON template_task.template_stage_id = stage.source_template_stage_id
			RETURNING id
		), activity AS (
			INSERT INTO bid_assembly_activity_logs (assembly_id, actor_user_id, actor_company_id, action, description)
			SELECT assembly.id, %s::uuid, %s::uuid, 'individual_assembly_created', 'Montagem individual iniciada com o Modelo LicitaHub.'
			FROM assembly WHERE EXISTS (SELECT 1 FROM inserted_assembly)
			RETURNING id
		)
		SELECT row_to_json(item) FROM (SELECT assembly.id::text AS id FROM assembly) item;
	`, sqlQuote(req.InterestID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), titleSQL, dueDateSQL,
		sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.CompanyID))
}

func buildAssemblyDetailSQL(assemblyID string, session sessionUser) string {
	return fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				ba.id::text AS id,
				COALESCE(ba.consortium_intention_id::text, '') AS "consortiumIntentionId",
				COALESCE(ci.match_id::text, '') AS "matchId",
				ba.assembly_type AS "assemblyType",
				COALESCE(ba.owner_company_id::text, '') AS "ownerCompanyId",
				ba.tender_id::text AS "tenderId",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				t.opening_date AS "openingDate",
				ba.title,
				ba.status,
				ba.start_date AS "startDate",
				ba.due_date AS "dueDate",
				ba.lead_company_id::text AS "leadCompanyId",
				lead.trade_name AS "leadCompanyName",
				(ba.lead_company_id = %s::uuid AND %s) AS "canManage",
				%s AS "canWork",
				COALESCE((
					SELECT json_agg(row_to_json(member) ORDER BY member."companyName")
					FROM (
						SELECT bap.company_id::text AS "companyId", c.trade_name AS "companyName", bap.role
						FROM bid_assembly_participants bap
						JOIN companies c ON c.id = bap.company_id
						WHERE bap.assembly_id = ba.id AND bap.user_id IS NULL AND bap.status = 'active'
					) member
				), '[]'::json) AS members,
				COALESCE((
					SELECT json_agg(row_to_json(professional) ORDER BY professional."companyName", professional."fullName")
					FROM (
						SELECT u.id::text AS id, u.company_id::text AS "companyId", c.trade_name AS "companyName",
							u.full_name AS "fullName", COALESCE(u.job_title, '') AS "jobTitle",
							COALESCE(photo.file_url, '') AS "profilePhotoUrl"
						FROM users u
						JOIN companies c ON c.id = u.company_id
						JOIN bid_assembly_participants bap ON bap.assembly_id = ba.id
							AND bap.company_id = u.company_id AND bap.user_id IS NULL AND bap.status = 'active'
						LEFT JOIN media_files photo ON photo.id = u.profile_photo_media_id
						WHERE u.status = 'active' AND u.deleted_at IS NULL
					) professional
				), '[]'::json) AS professionals,
				COALESCE((
					SELECT json_agg(row_to_json(stage_item) ORDER BY stage_item.position)
					FROM (
						SELECT stage.id::text AS id, stage.title, COALESCE(stage.description, '') AS description,
							stage.position, stage.is_custom AS "isCustom",
							COALESCE((SELECT count(*) FROM bid_assembly_tasks task WHERE task.stage_id = stage.id), 0)::int AS "taskCount",
							COALESCE((SELECT count(*) FROM bid_assembly_tasks task WHERE task.stage_id = stage.id AND task.status = 'completed'), 0)::int AS "completedCount",
							COALESCE((SELECT count(*) FROM bid_assembly_tasks task WHERE task.stage_id = stage.id AND task.status = 'not_applicable'), 0)::int AS "notApplicableCount",
							COALESCE((
								SELECT round(100 * sum(CASE WHEN task.status = 'completed' THEN task.weight ELSE 0 END) /
									NULLIF(sum(CASE WHEN task.status <> 'not_applicable' THEN task.weight ELSE 0 END), 0))::int
								FROM bid_assembly_tasks task WHERE task.stage_id = stage.id
							), CASE WHEN EXISTS (SELECT 1 FROM bid_assembly_tasks task WHERE task.stage_id = stage.id) THEN 100 ELSE 0 END) AS progress,
							COALESCE((
								SELECT json_agg(row_to_json(task_item) ORDER BY task_item.position)
								FROM (
									SELECT task.id::text AS id, task.title, COALESCE(task.description, '') AS description,
										task.position, task.status, task.priority, task.weight,
										COALESCE(task.responsible_company_id::text, '') AS "responsibleCompanyId",
										COALESCE(responsible_company.trade_name, '') AS "responsibleCompanyName",
										COALESCE(task.responsible_user_id::text, '') AS "responsibleUserId",
										COALESCE(responsible_user.full_name, '') AS "responsibleUserName",
										task.due_at AS "dueAt", task.submitted_at AS "submittedAt", task.completed_at AS "completedAt",
										task.is_custom AS "isCustom",
										COALESCE((
											SELECT json_agg(row_to_json(comment_item) ORDER BY comment_item."createdAt")
											FROM (
												SELECT comment.id::text AS id, comment.content,
													u.full_name AS "userName", c.trade_name AS "companyName", comment.created_at AS "createdAt"
												FROM bid_assembly_task_comments comment
												LEFT JOIN users u ON u.id = comment.user_id
												LEFT JOIN companies c ON c.id = comment.company_id
												WHERE comment.task_id = task.id AND comment.deleted_at IS NULL
											) comment_item
										), '[]'::json) AS comments,
										COALESCE((
											SELECT json_agg(row_to_json(evidence_item) ORDER BY evidence_item."createdAt" DESC)
											FROM (
												SELECT evidence.id::text AS id, evidence.evidence_type AS "evidenceType", evidence.title,
													COALESCE(media.file_url, evidence.external_url, '') AS url,
													COALESCE(evidence.note, '') AS note, evidence.version_number AS "versionNumber",
													evidence.status, COALESCE(u.full_name, '') AS "userName",
													COALESCE(c.trade_name, '') AS "companyName", evidence.created_at AS "createdAt"
												FROM bid_assembly_task_evidences evidence
												LEFT JOIN media_files media ON media.id = evidence.media_file_id
												LEFT JOIN users u ON u.id = evidence.uploaded_by_user_id
												LEFT JOIN companies c ON c.id = evidence.company_id
												WHERE evidence.task_id = task.id
											) evidence_item
										), '[]'::json) AS evidences
									FROM bid_assembly_tasks task
									LEFT JOIN companies responsible_company ON responsible_company.id = task.responsible_company_id
									LEFT JOIN users responsible_user ON responsible_user.id = task.responsible_user_id
									WHERE task.stage_id = stage.id
								) task_item
							), '[]'::json) AS tasks
						FROM bid_assembly_stages stage
						WHERE stage.assembly_id = ba.id AND stage.is_archived = false
					) stage_item
				), '[]'::json) AS stages
			FROM bid_assemblies ba
			LEFT JOIN consortium_intentions ci ON ci.id = ba.consortium_intention_id
			JOIN tenders t ON t.id = ba.tender_id
			JOIN companies lead ON lead.id = ba.lead_company_id
			JOIN bid_assembly_participants access ON access.assembly_id = ba.id
				AND access.company_id = %s::uuid AND access.user_id IS NULL AND access.status = 'active'
			WHERE ba.id = %s::uuid AND ba.status NOT IN ('cancelled', 'paused')
			  AND t.status = 'published'
			LIMIT 1
		) item;
	`, sqlQuote(session.CompanyID), sqlBool(session.canCoordinateAssembly()), sqlBool(session.canWorkOnAssembly()),
		sqlQuote(session.CompanyID), sqlQuote(assemblyID))
}

func buildUpdateAssemblyTaskSQL(assemblyID, taskID string, req updateAssemblyTaskRequest, session sessionUser, dueAtSQL string) string {
	companySQL := nullUUID(req.ResponsibleCompanyID)
	userSQL := nullUUID(req.ResponsibleUserID)
	return fmt.Sprintf(`
		WITH selected AS (
			SELECT task.*, ba.id AS assembly_id, ba.lead_company_id,
				(ba.lead_company_id = %s::uuid AND %s) AS is_leader
			FROM bid_assembly_tasks task
			JOIN bid_assembly_stages stage ON stage.id = task.stage_id
			JOIN bid_assemblies ba ON ba.id = stage.assembly_id
			JOIN bid_assembly_participants access ON access.assembly_id = ba.id
				AND access.company_id = %s::uuid AND access.user_id IS NULL AND access.status = 'active'
			WHERE ba.id = %s::uuid AND task.id = %s::uuid AND ba.status <> 'cancelled'
			  AND (
				(ba.lead_company_id = %s::uuid AND %s)
				OR task.responsible_company_id = %s::uuid
				OR task.responsible_user_id = %s::uuid
			  )
			LIMIT 1
		), updated AS (
			UPDATE bid_assembly_tasks task
			SET title = CASE WHEN selected.is_leader THEN %s ELSE task.title END,
				description = CASE WHEN selected.is_leader THEN NULLIF(%s, '') ELSE task.description END,
				status = CASE WHEN NOT selected.is_leader AND %s = 'completed' THEN 'under_review' ELSE %s END,
				priority = CASE WHEN selected.is_leader THEN %s ELSE task.priority END,
				responsible_company_id = CASE
					WHEN selected.is_leader AND %s IS NULL THEN NULL
					WHEN selected.is_leader AND EXISTS (
						SELECT 1 FROM bid_assembly_participants bap
						WHERE bap.assembly_id = selected.assembly_id AND bap.company_id = %s AND bap.status = 'active'
					) THEN %s ELSE task.responsible_company_id END,
				responsible_user_id = CASE
					WHEN selected.is_leader AND %s IS NULL THEN NULL
					WHEN selected.is_leader AND EXISTS (
						SELECT 1 FROM users u WHERE u.id = %s AND u.company_id = %s AND u.status = 'active' AND u.deleted_at IS NULL
					) THEN %s ELSE task.responsible_user_id END,
				due_at = CASE WHEN selected.is_leader THEN %s ELSE task.due_at END,
				submitted_at = CASE WHEN (CASE WHEN NOT selected.is_leader AND %s = 'completed' THEN 'under_review' ELSE %s END) = 'under_review' THEN now() ELSE task.submitted_at END,
				completed_at = CASE WHEN selected.is_leader AND %s = 'completed' THEN now() WHEN %s <> 'completed' THEN NULL ELSE task.completed_at END,
				updated_at = now()
			FROM selected
			WHERE task.id = selected.id
			RETURNING task.*, selected.assembly_id, selected.is_leader
		), activity AS (
			INSERT INTO bid_assembly_activity_logs (assembly_id, task_id, actor_user_id, actor_company_id, action, description)
			SELECT updated.assembly_id, updated.id, %s::uuid, %s::uuid, 'task_updated',
				'Tarefa atualizada para o status ' || updated.status || '.'
			FROM updated RETURNING id
		), assignment_notification AS (
			INSERT INTO notifications (recipient_user_id, recipient_company_id, type, title, message, destination_screen, related_entity_type, related_entity_id)
			SELECT updated.responsible_user_id, updated.responsible_company_id, 'system', 'Tarefa atribuida na Central de Montagem',
				updated.title, 'assembly-board', 'bid_assembly', updated.assembly_id
			FROM updated
			WHERE updated.responsible_user_id IS NOT NULL AND updated.responsible_user_id <> %s::uuid
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (SELECT updated.id::text AS id, updated.status, updated.updated_at AS "updatedAt" FROM updated) item;
	`, sqlQuote(session.CompanyID), sqlBool(session.canCoordinateAssembly()), sqlQuote(session.CompanyID), sqlQuote(assemblyID), sqlQuote(taskID),
		sqlQuote(session.CompanyID), sqlBool(session.canCoordinateAssembly()), sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(req.Title), sqlQuote(req.Description),
		sqlQuote(req.Status), sqlQuote(req.Status), sqlQuote(req.Priority), companySQL, companySQL, companySQL,
		userSQL, userSQL, companySQL, userSQL, dueAtSQL, sqlQuote(req.Status), sqlQuote(req.Status), sqlQuote(req.Status), sqlQuote(req.Status),
		sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(session.UserID))
}

func buildCreateAssemblyCommentSQL(assemblyID, taskID, content string, session sessionUser) string {
	return fmt.Sprintf(`
		WITH allowed AS (
			SELECT task.id, ba.id AS assembly_id, task.responsible_user_id, ba.lead_company_id
			FROM bid_assembly_tasks task
			JOIN bid_assembly_stages stage ON stage.id = task.stage_id
			JOIN bid_assemblies ba ON ba.id = stage.assembly_id
			JOIN bid_assembly_participants access ON access.assembly_id = ba.id
				AND access.company_id = %s::uuid AND access.user_id IS NULL AND access.status = 'active'
			WHERE ba.id = %s::uuid AND task.id = %s::uuid AND ba.status <> 'cancelled'
		), inserted AS (
			INSERT INTO bid_assembly_task_comments (task_id, user_id, company_id, content)
			SELECT allowed.id, %s::uuid, %s::uuid, %s FROM allowed RETURNING *
		), activity AS (
			INSERT INTO bid_assembly_activity_logs (assembly_id, task_id, actor_user_id, actor_company_id, action, description)
			SELECT allowed.assembly_id, allowed.id, %s::uuid, %s::uuid, 'comment_added', 'Comentario adicionado a tarefa.'
			FROM allowed RETURNING id
		), notified AS (
			INSERT INTO notifications (recipient_user_id, recipient_company_id, type, title, message, destination_screen, related_entity_type, related_entity_id)
			SELECT allowed.responsible_user_id, u.company_id, 'system', 'Novo comentario em sua tarefa', %s,
				'assembly-board', 'bid_assembly', allowed.assembly_id
			FROM allowed JOIN users u ON u.id = allowed.responsible_user_id
			WHERE allowed.responsible_user_id IS NOT NULL AND allowed.responsible_user_id <> %s::uuid
			RETURNING id
		)
		SELECT row_to_json(item) FROM (
			SELECT inserted.id::text AS id, inserted.content, inserted.created_at AS "createdAt" FROM inserted
		) item;
	`, sqlQuote(session.CompanyID), sqlQuote(assemblyID), sqlQuote(taskID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(content),
		sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(content), sqlQuote(session.UserID))
}

func buildCreateAssemblyEvidenceSQL(assemblyID, taskID string, req createAssemblyEvidenceRequest, fileURL string, fileSize int64, session sessionUser) string {
	mediaCTE := `selected_media AS (SELECT NULL::uuid AS id WHERE false),`
	if req.EvidenceType == "file" {
		mediaCTE = fmt.Sprintf(`selected_media AS (
			INSERT INTO media_files (company_id, uploaded_by_user_id, media_type, file_name, file_url, mime_type, file_size, source)
			VALUES (%s::uuid, %s::uuid, 'assembly_document', %s, %s, NULLIF(%s, ''), %d, 'upload')
			RETURNING id
		),`, sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(req.FileName), sqlQuote(fileURL), sqlQuote(req.MimeType), fileSize)
	}
	return fmt.Sprintf(`
		WITH allowed AS (
			SELECT task.id, ba.id AS assembly_id
			FROM bid_assembly_tasks task
			JOIN bid_assembly_stages stage ON stage.id = task.stage_id
			JOIN bid_assemblies ba ON ba.id = stage.assembly_id
			JOIN bid_assembly_participants access ON access.assembly_id = ba.id
				AND access.company_id = %s::uuid AND access.user_id IS NULL AND access.status = 'active'
			WHERE ba.id = %s::uuid AND task.id = %s::uuid AND ba.status <> 'cancelled'
		), %s inserted AS (
			INSERT INTO bid_assembly_task_evidences (
				task_id, media_file_id, evidence_type, title, external_url, note,
				version_number, uploaded_by_user_id, company_id
			)
			SELECT allowed.id, (SELECT id FROM selected_media LIMIT 1), %s, %s, NULLIF(%s, ''), NULLIF(%s, ''),
				COALESCE((SELECT max(version_number) + 1 FROM bid_assembly_task_evidences WHERE task_id = allowed.id AND title = %s), 1),
				%s::uuid, %s::uuid
			FROM allowed RETURNING *
		), activity AS (
			INSERT INTO bid_assembly_activity_logs (assembly_id, task_id, actor_user_id, actor_company_id, action, description)
			SELECT allowed.assembly_id, allowed.id, %s::uuid, %s::uuid, 'evidence_added', 'Evidencia adicionada: ' || %s
			FROM allowed RETURNING id
		)
		SELECT row_to_json(item) FROM (
			SELECT inserted.id::text AS id, inserted.title, inserted.evidence_type AS "evidenceType", inserted.created_at AS "createdAt" FROM inserted
		) item;
	`, sqlQuote(session.CompanyID), sqlQuote(assemblyID), sqlQuote(taskID), mediaCTE, sqlQuote(req.EvidenceType), sqlQuote(req.Title),
		sqlQuote(req.ExternalURL), sqlQuote(req.Note), sqlQuote(req.Title), sqlQuote(session.UserID), sqlQuote(session.CompanyID),
		sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(req.Title))
}

func normalizeAssemblyTaskStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending", "in_progress", "waiting_information", "blocked", "under_review", "returned_for_adjustment", "completed", "not_applicable":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "pending"
	}
}

func normalizeAssemblyTaskPriority(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low", "normal", "high", "urgent":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "normal"
	}
}

func nullUUID(value string) string {
	if strings.TrimSpace(value) == "" {
		return "NULL::uuid"
	}
	return sqlQuote(strings.TrimSpace(value)) + "::uuid"
}

func sqlBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func saveAssemblyFile(dataURL, originalName string) (string, int64, error) {
	_, payload, ok := strings.Cut(dataURL, ",")
	if !ok {
		return "", 0, errors.New("arquivo invalido")
	}
	content, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", 0, errors.New("nao foi possivel ler o arquivo")
	}
	if len(content) > 20*1024*1024 {
		return "", 0, errors.New("arquivo maior que 20MB")
	}
	ext := strings.ToLower(filepath.Ext(originalName))
	allowed := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".zip": true,
		".html": true, ".htm": true, ".txt": true,
	}
	if !allowed[ext] {
		return "", 0, errors.New("tipo de arquivo nao permitido")
	}
	dir := filepath.Join("uploads", "assemblies")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, errors.New("nao foi possivel preparar a pasta de documentos")
	}
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return "", 0, errors.New("nao foi possivel gerar o nome do arquivo")
	}
	fileName := fmt.Sprintf("assembly-%d-%s%s", time.Now().UnixNano(), hex.EncodeToString(random), ext)
	if err := os.WriteFile(filepath.Join(dir, fileName), content, 0644); err != nil {
		return "", 0, errors.New("nao foi possivel salvar o documento")
	}
	url := strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/") + "/uploads/assemblies/" + fileName
	return url, int64(len(content)), nil
}

const assemblySchemaSQL = `
CREATE TABLE IF NOT EXISTS bid_assembly_templates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), company_id uuid REFERENCES companies(id) ON DELETE CASCADE,
  name varchar(180) NOT NULL, description text, is_system boolean NOT NULL DEFAULT false,
  is_active boolean NOT NULL DEFAULT true, created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS bid_assembly_templates_system_name_uk ON bid_assembly_templates(name) WHERE is_system = true;
CREATE TABLE IF NOT EXISTS bid_assembly_template_stages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), template_id uuid NOT NULL REFERENCES bid_assembly_templates(id) ON DELETE CASCADE,
  title varchar(180) NOT NULL, description text, position integer NOT NULL DEFAULT 0, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS bid_assembly_template_tasks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), template_stage_id uuid NOT NULL REFERENCES bid_assembly_template_stages(id) ON DELETE CASCADE,
  title varchar(220) NOT NULL, description text, position integer NOT NULL DEFAULT 0,
  default_weight numeric(8,2) NOT NULL DEFAULT 1, default_days_offset integer, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS bid_assemblies (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), consortium_intention_id uuid REFERENCES consortium_intentions(id) ON DELETE RESTRICT,
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE RESTRICT, template_id uuid REFERENCES bid_assembly_templates(id) ON DELETE SET NULL,
  assembly_type varchar(30) NOT NULL DEFAULT 'consortium', owner_company_id uuid REFERENCES companies(id),
  lead_company_id uuid NOT NULL REFERENCES companies(id), title varchar(220) NOT NULL, status varchar(40) NOT NULL DEFAULT 'preparing', status_before_pause varchar(40),
  start_date date NOT NULL DEFAULT CURRENT_DATE, due_date date, created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bid_assemblies_status_chk CHECK (status IN ('preparing','in_progress','under_review','ready_to_submit','submitted','paused','cancelled')),
  CONSTRAINT bid_assemblies_type_chk CHECK (assembly_type IN ('consortium','individual')),
  CONSTRAINT bid_assemblies_consortium_uk UNIQUE (consortium_intention_id)
);
ALTER TABLE bid_assemblies ALTER COLUMN consortium_intention_id DROP NOT NULL;
ALTER TABLE bid_assemblies ADD COLUMN IF NOT EXISTS assembly_type varchar(30) NOT NULL DEFAULT 'consortium';
ALTER TABLE bid_assemblies ADD COLUMN IF NOT EXISTS owner_company_id uuid REFERENCES companies(id);
UPDATE bid_assemblies SET owner_company_id = lead_company_id WHERE owner_company_id IS NULL;
ALTER TABLE bid_assemblies DROP CONSTRAINT IF EXISTS bid_assemblies_type_chk;
ALTER TABLE bid_assemblies ADD CONSTRAINT bid_assemblies_type_chk CHECK (assembly_type IN ('consortium','individual'));
CREATE UNIQUE INDEX IF NOT EXISTS bid_assemblies_individual_company_tender_uk ON bid_assemblies(tender_id, owner_company_id) WHERE assembly_type = 'individual' AND status <> 'cancelled';
CREATE TABLE IF NOT EXISTS bid_assembly_participants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), assembly_id uuid NOT NULL REFERENCES bid_assemblies(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id), user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  role varchar(40) NOT NULL DEFAULT 'collaborator', status varchar(40) NOT NULL DEFAULT 'active',
  joined_at timestamptz NOT NULL DEFAULT now(), removed_at timestamptz,
  CONSTRAINT bid_assembly_participants_role_chk CHECK (role IN ('coordinator','collaborator','viewer')),
  CONSTRAINT bid_assembly_participants_status_chk CHECK (status IN ('active','removed'))
);
CREATE UNIQUE INDEX IF NOT EXISTS bid_assembly_participants_company_uk ON bid_assembly_participants(assembly_id, company_id) WHERE user_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS bid_assembly_participants_user_uk ON bid_assembly_participants(assembly_id, user_id) WHERE user_id IS NOT NULL;
CREATE TABLE IF NOT EXISTS bid_assembly_stages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), assembly_id uuid NOT NULL REFERENCES bid_assemblies(id) ON DELETE CASCADE,
  source_template_stage_id uuid REFERENCES bid_assembly_template_stages(id) ON DELETE SET NULL,
  title varchar(180) NOT NULL, description text, position integer NOT NULL DEFAULT 0, is_custom boolean NOT NULL DEFAULT false,
  is_archived boolean NOT NULL DEFAULT false, created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS bid_assembly_tasks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), stage_id uuid NOT NULL REFERENCES bid_assembly_stages(id) ON DELETE CASCADE,
  source_template_task_id uuid REFERENCES bid_assembly_template_tasks(id) ON DELETE SET NULL,
  title varchar(220) NOT NULL, description text, position integer NOT NULL DEFAULT 0, status varchar(40) NOT NULL DEFAULT 'pending',
  priority varchar(20) NOT NULL DEFAULT 'normal', weight numeric(8,2) NOT NULL DEFAULT 1,
  responsible_company_id uuid REFERENCES companies(id) ON DELETE SET NULL, responsible_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  due_at timestamptz, submitted_at timestamptz, completed_at timestamptz, created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  is_custom boolean NOT NULL DEFAULT false, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bid_assembly_tasks_status_chk CHECK (status IN ('pending','in_progress','waiting_information','blocked','under_review','returned_for_adjustment','completed','not_applicable')),
  CONSTRAINT bid_assembly_tasks_priority_chk CHECK (priority IN ('low','normal','high','urgent')),
  CONSTRAINT bid_assembly_tasks_weight_chk CHECK (weight > 0)
);
CREATE TABLE IF NOT EXISTS bid_assembly_task_comments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), task_id uuid NOT NULL REFERENCES bid_assembly_tasks(id) ON DELETE CASCADE,
  user_id uuid REFERENCES users(id) ON DELETE SET NULL, company_id uuid REFERENCES companies(id) ON DELETE SET NULL,
  content text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz
);
CREATE TABLE IF NOT EXISTS bid_assembly_task_evidences (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), task_id uuid NOT NULL REFERENCES bid_assembly_tasks(id) ON DELETE CASCADE,
  media_file_id uuid REFERENCES media_files(id) ON DELETE SET NULL, evidence_type varchar(30) NOT NULL,
  title varchar(220) NOT NULL, external_url text, note text, version_number integer NOT NULL DEFAULT 1,
  status varchar(30) NOT NULL DEFAULT 'current', uploaded_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  company_id uuid REFERENCES companies(id) ON DELETE SET NULL, created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bid_assembly_evidence_type_chk CHECK (evidence_type IN ('file','link','note')),
  CONSTRAINT bid_assembly_evidence_status_chk CHECK (status IN ('current','superseded','approved'))
);
CREATE TABLE IF NOT EXISTS bid_assembly_deadline_alerts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), task_id uuid NOT NULL REFERENCES bid_assembly_tasks(id) ON DELETE CASCADE,
  recipient_user_id uuid REFERENCES users(id) ON DELETE CASCADE, alert_type varchar(30) NOT NULL,
  alert_date date NOT NULL DEFAULT CURRENT_DATE, created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bid_assembly_deadline_alerts_type_chk CHECK (alert_type IN ('due_soon','due_today','overdue')),
  CONSTRAINT bid_assembly_deadline_alerts_uk UNIQUE (task_id, recipient_user_id, alert_type, alert_date)
);
CREATE TABLE IF NOT EXISTS bid_assembly_activity_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), assembly_id uuid NOT NULL REFERENCES bid_assemblies(id) ON DELETE CASCADE,
  task_id uuid REFERENCES bid_assembly_tasks(id) ON DELETE CASCADE, actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  actor_company_id uuid REFERENCES companies(id) ON DELETE SET NULL, action varchar(80) NOT NULL,
  description text, metadata jsonb NOT NULL DEFAULT '{}'::jsonb, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_stages_assembly_position ON bid_assembly_stages(assembly_id, position);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_tasks_stage_position ON bid_assembly_tasks(stage_id, position);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_tasks_responsible_due ON bid_assembly_tasks(responsible_user_id, due_at, status);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_comments_task ON bid_assembly_task_comments(task_id, created_at);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_evidences_task ON bid_assembly_task_evidences(task_id, created_at);
`

const defaultAssemblyTemplateSQL = `
INSERT INTO bid_assembly_templates (name, description, is_system, is_active)
VALUES ('Modelo LicitaHub', 'Estrutura padrao de montagem colaborativa de licitacoes em oito fases.', true, true)
ON CONFLICT DO NOTHING;
WITH template AS (SELECT id FROM bid_assembly_templates WHERE name = 'Modelo LicitaHub' AND is_system = true LIMIT 1),
incoming(title, description, position) AS (VALUES
 ('Planejamento da montagem','Leitura do edital, estrategia, calendario e organizacao do trabalho.',1),
 ('Concepcao consorcial','Formalizacao, identidade e responsabilidades do consorcio.',2),
 ('Montagem da peca qualitativa','Desenvolvimento dos conteudos tecnicos e qualitativos da proposta.',3),
 ('Montagem do orcamento','Viabilidade, formacao de preco e planilhas comerciais.',4),
 ('Montagem da equipe tecnica','Definicao da equipe e consolidacao de comprovacoes profissionais.',5),
 ('Montagem das declaracoes','Preparacao e assinatura das declaracoes exigidas.',6),
 ('Certificacoes e quesitos de pontuacao','Organizacao de certificados e comprovacoes adicionais.',7),
 ('Revisao e consolidacao','Conferencia, consolidacao, assinatura e preparacao para envio.',8)
)
INSERT INTO bid_assembly_template_stages (template_id,title,description,position)
SELECT template.id,incoming.title,incoming.description,incoming.position FROM template CROSS JOIN incoming
WHERE NOT EXISTS (SELECT 1 FROM bid_assembly_template_stages s WHERE s.template_id=template.id AND s.title=incoming.title);
WITH template AS (SELECT id FROM bid_assembly_templates WHERE name='Modelo LicitaHub' AND is_system=true LIMIT 1),
incoming(stage_title,title,description,position,weight) AS (VALUES
 ('Concepcao consorcial','Elaborar o termo de constituicao do consorcio','Consolidar participantes, objeto, compromissos e regras de representacao.',1,2),
 ('Concepcao consorcial','Definir nome do consorcio','Registrar a denominacao que sera usada na proposta.',2,1),
 ('Concepcao consorcial','Definir identidade visual do consorcio','Organizar logomarca, capa e padrao visual dos documentos.',3,1),
 ('Concepcao consorcial','Definir papeis e responsabilidades das empresas','Registrar lideranca, entregas e pontos focais de cada consorciada.',4,2),
 ('Planejamento da montagem','Realizar leitura orientada do edital','Mapear entregaveis, criterios, prazos e riscos da proposta.',1,2),
 ('Planejamento da montagem','Montar calendario geral da proposta','Definir marcos internos anteriores a data oficial de entrega.',2,1),
 ('Planejamento da montagem','Distribuir responsabilidades por fase','Atribuir empresas e profissionais responsaveis.',3,1),
 ('Planejamento da montagem','Organizar repositorio compartilhado','Definir estrutura de pastas, nomes e controle de versoes.',4,1),
 ('Montagem da peca qualitativa','Elaborar conhecimento do objeto','Descrever compreensao do contexto, desafios e objetivos do contrato.',1,3),
 ('Montagem da peca qualitativa','Descrever produtos e entregaveis','Consolidar produtos, resultados e criterios de aceitacao.',2,2),
 ('Montagem da peca qualitativa','Desenvolver metodologia','Detalhar abordagem, procedimentos, ferramentas e integracao das disciplinas.',3,4),
 ('Montagem da peca qualitativa','Elaborar plano de trabalho','Organizar atividades, sequencia, interfaces e responsabilidades.',4,3),
 ('Montagem da peca qualitativa','Revisar aderencia aos criterios tecnicos','Conferir atendimento e potencial de pontuacao da peca.',5,2),
 ('Montagem do orcamento','Analisar viabilidade financeira','Avaliar custos, riscos, fluxo e condicoes comerciais.',1,2),
 ('Montagem do orcamento','Definir preco da proposta','Consolidar estrategia de preco e premissas comerciais.',2,2),
 ('Montagem do orcamento','Montar planilha orcamentaria','Preparar quantitativos, custos, encargos e composicoes.',3,3),
 ('Montagem do orcamento','Revisar tributos, BDI e consistencia','Conferir calculos, incidencias e compatibilidade com o edital.',4,2),
 ('Montagem da equipe tecnica','Analisar profissionais disponiveis','Verificar disponibilidade, vinculo e aderencia dos profissionais.',1,2),
 ('Montagem da equipe tecnica','Montar quadro de experiencia profissional','Consolidar experiencias e pontuacoes por profissional.',2,2),
 ('Montagem da equipe tecnica','Disponibilizar CATs dos profissionais','Anexar comprovacoes de responsabilidade e experiencia tecnica.',3,2),
 ('Montagem da equipe tecnica','Disponibilizar documentos de formacao','Anexar diplomas, certificados e demais comprovacoes academicas.',4,1),
 ('Montagem da equipe tecnica','Preparar declaracoes dos profissionais','Consolidar disponibilidade, compromisso e autorizacoes exigidas.',5,1),
 ('Montagem da equipe tecnica','Disponibilizar registros profissionais','Anexar registros e certidoes dos conselhos de classe.',6,1),
 ('Montagem da equipe tecnica','Revisar atendimento e pontuacao da equipe','Conferir lacunas, validade e potencial de pontuacao.',7,2),
 ('Montagem das declaracoes','Mapear declaracoes exigidas','Criar lista completa conforme edital e anexos.',1,1),
 ('Montagem das declaracoes','Elaborar declaracoes','Preencher os modelos e textos aplicaveis ao consorcio.',2,2),
 ('Montagem das declaracoes','Coletar assinaturas e validar poderes','Conferir assinantes, procuracoes e versoes finais.',3,2),
 ('Certificacoes e quesitos de pontuacao','Mapear certificacoes requeridas','Identificar certificados obrigatorios e pontuaveis.',1,1),
 ('Certificacoes e quesitos de pontuacao','Disponibilizar documentos comprobatorios','Anexar certificacoes e demais evidencias de pontuacao.',2,2),
 ('Certificacoes e quesitos de pontuacao','Conferir validade e aderencia','Verificar emissor, prazo, escopo e criterio atendido.',3,1),
 ('Revisao e consolidacao','Consolidar documentos da proposta','Reunir somente versoes atuais e aprovadas.',1,3),
 ('Revisao e consolidacao','Realizar revisao cruzada','Conferir coerencia entre tecnica, equipe, documentos e preco.',2,3),
 ('Revisao e consolidacao','Conferir assinaturas e formatos','Validar assinaturas, extensoes, limites e nomenclaturas.',3,2),
 ('Revisao e consolidacao','Validar dossie final','Confirmar que todas as evidencias obrigatorias estao presentes.',4,2),
 ('Revisao e consolidacao','Registrar protocolo de envio','Guardar comprovante, data, horario e versao submetida.',5,1)
)
INSERT INTO bid_assembly_template_tasks (template_stage_id,title,description,position,default_weight)
SELECT stage.id,incoming.title,incoming.description,incoming.position,incoming.weight FROM incoming JOIN template ON true
JOIN bid_assembly_template_stages stage ON stage.template_id=template.id AND stage.title=incoming.stage_title
WHERE NOT EXISTS (SELECT 1 FROM bid_assembly_template_tasks task WHERE task.template_stage_id=stage.id AND task.title=incoming.title);

WITH system_template AS (
 SELECT id FROM bid_assembly_templates WHERE name='Modelo LicitaHub' AND is_system=true
), reordered_template AS (
 UPDATE bid_assembly_template_stages stage
 SET position = CASE stage.title
  WHEN 'Planejamento da montagem' THEN 1
  WHEN 'Concepcao consorcial' THEN 2
  ELSE stage.position
 END
 FROM system_template template
 WHERE stage.template_id=template.id
   AND stage.title IN ('Planejamento da montagem','Concepcao consorcial')
 RETURNING stage.id, stage.position
)
UPDATE bid_assembly_stages assembly_stage
SET position = reordered_template.position
FROM reordered_template
WHERE assembly_stage.source_template_stage_id=reordered_template.id;
`
