package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type technicalCertificateRequest struct {
	CertificateNumber        string                                `json:"certificateNumber"`
	IssuerName               string                                `json:"issuerName"`
	IssuerDocument           string                                `json:"issuerDocument"`
	ContractedName           string                                `json:"contractedName"`
	ContractNumber           string                                `json:"contractNumber"`
	Object                   string                                `json:"object"`
	ServiceDescription       string                                `json:"serviceDescription"`
	State                    string                                `json:"state"`
	City                     string                                `json:"city"`
	ExecutionStart           string                                `json:"executionStart"`
	ExecutionEnd             string                                `json:"executionEnd"`
	ContractValue            string                                `json:"contractValue"`
	TechnicalManager         string                                `json:"technicalManager"`
	ProfessionalRegistration string                                `json:"professionalRegistration"`
	ArtCatReference          string                                `json:"artCatReference"`
	CATNumber                string                                `json:"catNumber"`
	TechnicalProfessionalID  string                                `json:"technicalProfessionalId"`
	CATProfessional          string                                `json:"catProfessional"`
	ProfessionalRole         string                                `json:"professionalRole"`
	CompletionStatus         string                                `json:"completionStatus"`
	UsageScope               string                                `json:"usageScope"`
	Quantities               []technicalCertificateQuantityRequest `json:"quantities"`
	Tags                     string                                `json:"tags"`
	ContentText              string                                `json:"contentText"`
	Status                   string                                `json:"status"`
	FileDataURL              string                                `json:"fileDataUrl"`
	FileName                 string                                `json:"fileName"`
	MimeType                 string                                `json:"mimeType"`
}

type technicalCertificateQuantityRequest struct {
	Description string `json:"description"`
	Value       string `json:"value"`
	Unit        string `json:"unit"`
	Note        string `json:"note"`
}

func (s sessionUser) canManageTechnicalCertificates() bool {
	return s.CompanyID != "" && s.RoleKey != "reader"
}

func (a *app) handleTechnicalCertificates(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "o acervo tecnico esta disponivel somente para empresas associadas")
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.listTechnicalCertificates(w, r, session)
	case http.MethodPost:
		if !session.canManageTechnicalCertificates() {
			writeError(w, http.StatusForbidden, "seu perfil pode consultar, mas nao cadastrar atestados")
			return
		}
		a.createTechnicalCertificate(w, r, session)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleTechnicalCertificateByPath(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "o acervo tecnico esta disponivel somente para empresas associadas")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/technical-certificates/"), "/"), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		writeError(w, http.StatusNotFound, "atestado nao informado")
		return
	}
	id := strings.TrimSpace(parts[0])
	if len(parts) == 2 && parts[1] == "document" && r.Method == http.MethodGet {
		a.downloadTechnicalCertificateDocument(w, r, id, session)
		return
	}
	if len(parts) == 2 && parts[1] == "reprocess-ocr" && r.Method == http.MethodPost {
		if !session.canManageTechnicalCertificates() {
			writeError(w, http.StatusForbidden, "seu perfil pode consultar, mas nao refazer a leitura")
			return
		}
		a.reprocessTechnicalCertificateOCR(w, r, id, session)
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.getTechnicalCertificate(w, r, id, session)
	case http.MethodPut:
		if !session.canManageTechnicalCertificates() {
			writeError(w, http.StatusForbidden, "seu perfil pode consultar, mas nao editar atestados")
			return
		}
		a.updateTechnicalCertificate(w, r, id, session)
	case http.MethodDelete:
		if !session.canManageTechnicalCertificates() {
			writeError(w, http.StatusForbidden, "seu perfil pode consultar, mas nao remover atestados")
			return
		}
		a.deleteTechnicalCertificate(w, r, id, session)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) listTechnicalCertificates(w http.ResponseWriter, r *http.Request, session sessionUser) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && status != "active" && status != "archived" {
		writeError(w, http.StatusBadRequest, "status de atestado invalido")
		return
	}
	statusFilter := ""
	if status != "" {
		statusFilter = fmt.Sprintf("AND tc.status = %s", sqlQuote(status))
	}
	searchFilter := ""
	if search != "" {
		needle := sqlQuote("%" + search + "%")
		searchFilter = fmt.Sprintf(`AND (
			concat_ws(' ', tc.certificate_number, tc.issuer_name, tc.contracted_name, tc.contract_number, tc.object, tc.service_description, tc.tags, tc.technical_manager, tc.art_cat_reference, tc.cat_number, tc.cat_professional, tc.professional_role, tc.extracted_text) ILIKE %s
			OR EXISTS (SELECT 1 FROM technical_certificate_quantities q WHERE q.technical_certificate_id = tc.id AND concat_ws(' ', q.description, q.unit, q.note) ILIKE %s)
		)`, needle, needle)
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT tc.id::text AS id, COALESCE(tc.certificate_number, '') AS "certificateNumber", COALESCE(tc.issuer_name, '') AS "issuerName",
				COALESCE(tc.issuer_document, '') AS "issuerDocument", COALESCE(tc.contracted_name, '') AS "contractedName", COALESCE(tc.contract_number, '') AS "contractNumber", tc.object,
				COALESCE(tc.service_description, '') AS "serviceDescription", COALESCE(tc.state, '') AS state, COALESCE(tc.city, '') AS city,
				tc.execution_start AS "executionStart", tc.execution_end AS "executionEnd", tc.contract_value AS "contractValue",
				COALESCE(tc.technical_manager, '') AS "technicalManager", COALESCE(tc.professional_registration, '') AS "professionalRegistration",
				COALESCE(tc.art_cat_reference, '') AS "artCatReference", COALESCE(tc.cat_number, '') AS "catNumber", COALESCE(tc.technical_professional_id::text, '') AS "technicalProfessionalId", COALESCE(tc.cat_professional, '') AS "catProfessional", COALESCE(tc.professional_role, '') AS "professionalRole", tc.completion_status AS "completionStatus", tc.usage_scope AS "usageScope", COALESCE(tc.tags, '') AS tags, tc.file_name AS "fileName",
				COALESCE(tc.mime_type, '') AS "mimeType", tc.file_size AS "fileSize", tc.extraction_status AS "extractionStatus", tc.status,
				tc.extracted_text AS "contentText", tc.created_at AS "createdAt", tc.updated_at AS "updatedAt",
				COALESCE((SELECT json_agg(json_build_object('id', q.id::text, 'description', q.description, 'value', q.quantity, 'unit', COALESCE(q.unit, ''), 'note', COALESCE(q.note, '')) ORDER BY q.display_order) FROM technical_certificate_quantities q WHERE q.technical_certificate_id = tc.id), '[]'::json) AS quantities,
				'/api/technical-certificates/' || tc.id::text || '/document' AS "documentUrl"
			FROM technical_certificates tc
			WHERE tc.company_id = %s::uuid AND tc.deleted_at IS NULL %s %s
			ORDER BY tc.updated_at DESC
		) item;
	`, sqlQuote(session.CompanyID), statusFilter, searchFilter))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o acervo tecnico")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) getTechnicalCertificate(w http.ResponseWriter, r *http.Request, id string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), technicalCertificateSelectSQL(id, session.CompanyID))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "atestado nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createTechnicalCertificate(w http.ResponseWriter, r *http.Request, session sessionUser) {
	var req technicalCertificateRequest
	if err := json.NewDecoder(ioLimitReader(r.Body, 28*1024*1024)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dados do atestado invalidos")
		return
	}
	if strings.TrimSpace(req.Object) == "" {
		writeError(w, http.StatusBadRequest, "objeto do atestado e obrigatorio")
		return
	}
	if strings.TrimSpace(req.FileDataURL) == "" {
		writeError(w, http.StatusBadRequest, "arquivo PDF do atestado e obrigatorio")
		return
	}
	storagePath, fileName, mimeType, fileSize, err := saveTechnicalCertificateDocument(req.FileDataURL, req.FileName, req.MimeType)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	extractedText := strings.TrimSpace(req.ContentText)
	extractionStatus := "manual"
	if extractedText == "" {
		extractedText, extractionStatus = extractTechnicalCertificateText(r.Context(), storagePath)
	}
	extractedTextPath := ""
	if extractedText != "" {
		var textErr error
		extractedTextPath, textErr = saveTechnicalCertificateText(storagePath, extractedText)
		if textErr != nil {
			_ = os.Remove(storagePath)
			writeError(w, http.StatusInternalServerError, "nao foi possivel salvar o texto lido do atestado")
			return
		}
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Status != "active" && req.Status != "archived" {
		writeError(w, http.StatusBadRequest, "status de atestado invalido")
		return
	}
	if req.CompletionStatus == "" {
		req.CompletionStatus = "final"
	}
	if req.CompletionStatus != "final" && req.CompletionStatus != "partial" {
		writeError(w, http.StatusBadRequest, "situacao de execucao invalida")
		return
	}
	if req.UsageScope == "" {
		req.UsageScope = "both"
	}
	if req.UsageScope != "company" && req.UsageScope != "professional" && req.UsageScope != "both" {
		writeError(w, http.StatusBadRequest, "uso permitido do atestado invalido")
		return
	}
	if !validTechnicalCertificateExecutionDates(req.ExecutionStart, req.ExecutionEnd) {
		writeError(w, http.StatusBadRequest, "a data de inicio da execucao deve ser anterior ou igual a data final")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH inserted AS (
			INSERT INTO technical_certificates (
				company_id, certificate_number, issuer_name, issuer_document, contracted_name, contract_number, object, service_description, state, city,
				execution_start, execution_end, contract_value, technical_manager, professional_registration, art_cat_reference, tags,
				cat_number, technical_professional_id, cat_professional, professional_role, completion_status, usage_scope, document_url, file_name, mime_type, file_size, extracted_text, extracted_text_file_path, extraction_status, status, uploaded_by_user_id
			) VALUES (
				%s::uuid, NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), %s, NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''),
				%s, %s, %s, NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), %s,
				NULLIF(%s, ''), NULLIF(%s, '')::uuid, NULLIF(%s, ''), NULLIF(%s, ''), %s, %s, %s, %s, NULLIF(%s, ''), %d, %s, NULLIF(%s, ''), %s, %s, %s::uuid
			) RETURNING id
		)
		SELECT row_to_json(item) FROM (
			SELECT id::text AS id FROM inserted
		) item;
	`, sqlQuote(session.CompanyID), sqlQuote(req.CertificateNumber), sqlQuote(req.IssuerName), sqlQuote(req.IssuerDocument), sqlQuote(req.ContractedName), sqlQuote(req.ContractNumber), sqlQuote(req.Object), sqlQuote(req.ServiceDescription), sqlQuote(normalizeState(req.State)), sqlQuote(req.City), nullableCertificateDateSQL(req.ExecutionStart), nullableCertificateDateSQL(req.ExecutionEnd), nullableCertificateValueSQL(req.ContractValue), sqlQuote(req.TechnicalManager), sqlQuote(req.ProfessionalRegistration), sqlQuote(req.ArtCatReference), sqlQuote(req.Tags), sqlQuote(req.CATNumber), sqlQuote(req.TechnicalProfessionalID), sqlQuote(req.CATProfessional), sqlQuote(req.ProfessionalRole), sqlQuote(req.CompletionStatus), sqlQuote(req.UsageScope), sqlQuote(storagePath), sqlQuote(fileName), sqlQuote(mimeType), fileSize, sqlQuote(extractedText), sqlQuote(extractedTextPath), sqlQuote(extractionStatus), sqlQuote(req.Status), sqlQuote(session.UserID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusInternalServerError, "nao foi possivel registrar o atestado")
		return
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(payload, &created)
	if created.ID == "" {
		writeError(w, http.StatusInternalServerError, "nao foi possivel identificar o atestado registrado")
		return
	}
	if err := a.replaceTechnicalCertificateQuantities(r.Context(), created.ID, req.Quantities); err != nil {
		writeError(w, http.StatusInternalServerError, "atestado registrado, mas nao foi possivel salvar os quantitativos")
		return
	}
	logCertificateAudit(r.Context(), a, session, created.ID, "created", "Atestado técnico incluído no acervo da empresa.")
	fullPayload, err := a.queryJSON(r.Context(), technicalCertificateSelectSQL(created.ID, session.CompanyID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "atestado registrado, mas nao foi possivel carregar o resultado")
		return
	}
	writeRawJSON(w, http.StatusCreated, fullPayload)
}

func (a *app) updateTechnicalCertificate(w http.ResponseWriter, r *http.Request, id string, session sessionUser) {
	var req technicalCertificateRequest
	if err := json.NewDecoder(ioLimitReader(r.Body, 3*1024*1024)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dados do atestado invalidos")
		return
	}
	if strings.TrimSpace(req.Object) == "" {
		writeError(w, http.StatusBadRequest, "objeto do atestado e obrigatorio")
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Status != "active" && req.Status != "archived" {
		writeError(w, http.StatusBadRequest, "status de atestado invalido")
		return
	}
	if req.CompletionStatus == "" {
		req.CompletionStatus = "final"
	}
	if req.CompletionStatus != "final" && req.CompletionStatus != "partial" {
		writeError(w, http.StatusBadRequest, "situacao de execucao invalida")
		return
	}
	if req.UsageScope == "" {
		req.UsageScope = "both"
	}
	if req.UsageScope != "company" && req.UsageScope != "professional" && req.UsageScope != "both" {
		writeError(w, http.StatusBadRequest, "uso permitido do atestado invalido")
		return
	}
	if !validTechnicalCertificateExecutionDates(req.ExecutionStart, req.ExecutionEnd) {
		writeError(w, http.StatusBadRequest, "a data de inicio da execucao deve ser anterior ou igual a data final")
		return
	}
	extractionStatus := "pending_ocr"
	if strings.TrimSpace(req.ContentText) != "" {
		extractionStatus = "manual"
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		UPDATE technical_certificates
		SET certificate_number = NULLIF(%s, ''), issuer_name = NULLIF(%s, ''), issuer_document = NULLIF(%s, ''), contracted_name = NULLIF(%s, ''), contract_number = NULLIF(%s, ''),
			object = %s, service_description = NULLIF(%s, ''), state = NULLIF(%s, ''), city = NULLIF(%s, ''), execution_start = %s, execution_end = %s,
			contract_value = %s, technical_manager = NULLIF(%s, ''), professional_registration = NULLIF(%s, ''), art_cat_reference = NULLIF(%s, ''),
			tags = %s, cat_number = NULLIF(%s, ''), technical_professional_id = NULLIF(%s, '')::uuid, cat_professional = NULLIF(%s, ''), professional_role = NULLIF(%s, ''), completion_status = %s, usage_scope = %s, extracted_text = %s, extraction_status = %s, status = %s, updated_at = now()
		WHERE id = %s::uuid AND company_id = %s::uuid AND deleted_at IS NULL
		RETURNING id::text AS id, document_url AS "documentUrl";
	`, sqlQuote(req.CertificateNumber), sqlQuote(req.IssuerName), sqlQuote(req.IssuerDocument), sqlQuote(req.ContractedName), sqlQuote(req.ContractNumber), sqlQuote(req.Object), sqlQuote(req.ServiceDescription), sqlQuote(normalizeState(req.State)), sqlQuote(req.City), nullableCertificateDateSQL(req.ExecutionStart), nullableCertificateDateSQL(req.ExecutionEnd), nullableCertificateValueSQL(req.ContractValue), sqlQuote(req.TechnicalManager), sqlQuote(req.ProfessionalRegistration), sqlQuote(req.ArtCatReference), sqlQuote(req.Tags), sqlQuote(req.CATNumber), sqlQuote(req.TechnicalProfessionalID), sqlQuote(req.CATProfessional), sqlQuote(req.ProfessionalRole), sqlQuote(req.CompletionStatus), sqlQuote(req.UsageScope), sqlQuote(strings.TrimSpace(req.ContentText)), sqlQuote(extractionStatus), sqlQuote(req.Status), sqlQuote(id), sqlQuote(session.CompanyID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "atestado nao encontrado")
		return
	}
	var updated struct {
		DocumentURL string `json:"documentUrl"`
	}
	_ = json.Unmarshal(payload, &updated)
	if updated.DocumentURL != "" && strings.TrimSpace(req.ContentText) != "" {
		textPath, textErr := saveTechnicalCertificateText(updated.DocumentURL, strings.TrimSpace(req.ContentText))
		if textErr != nil {
			writeError(w, http.StatusInternalServerError, "atestado atualizado, mas nao foi possivel salvar o texto corrigido")
			return
		}
		_, _ = a.runPSQL(r.Context(), fmt.Sprintf(`UPDATE technical_certificates SET extracted_text_file_path = %s, updated_at = now() WHERE id = %s::uuid AND company_id = %s::uuid;`, sqlQuote(textPath), sqlQuote(id), sqlQuote(session.CompanyID)))
	}
	if err := a.replaceTechnicalCertificateQuantities(r.Context(), id, req.Quantities); err != nil {
		writeError(w, http.StatusInternalServerError, "dados atualizados, mas nao foi possivel salvar os quantitativos")
		return
	}
	logCertificateAudit(r.Context(), a, session, id, "updated", "Dados do atestado técnico atualizados.")
	fullPayload, err := a.queryJSON(r.Context(), technicalCertificateSelectSQL(id, session.CompanyID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o atestado atualizado")
		return
	}
	writeRawJSON(w, http.StatusOK, fullPayload)
}

func validTechnicalCertificateExecutionDates(start, end string) bool {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if start == "" || end == "" {
		return true
	}
	startDate, startErr := time.Parse("2006-01-02", start)
	endDate, endErr := time.Parse("2006-01-02", end)
	return startErr == nil && endErr == nil && !startDate.After(endDate)
}

func (a *app) replaceTechnicalCertificateQuantities(ctx context.Context, certificateID string, quantities []technicalCertificateQuantityRequest) error {
	values := make([]string, 0, len(quantities))
	for index, quantity := range quantities {
		description := strings.TrimSpace(quantity.Description)
		if description == "" {
			continue
		}
		values = append(values, fmt.Sprintf("(%s::uuid, %s, %s, NULLIF(%s, ''), NULLIF(%s, ''), %d)", sqlQuote(certificateID), sqlQuote(description), nullableCertificateQuantitySQL(quantity.Value), sqlQuote(strings.TrimSpace(quantity.Unit)), sqlQuote(strings.TrimSpace(quantity.Note)), index))
	}
	sql := fmt.Sprintf("BEGIN; DELETE FROM technical_certificate_quantities WHERE technical_certificate_id = %s::uuid;", sqlQuote(certificateID))
	if len(values) > 0 {
		sql += " INSERT INTO technical_certificate_quantities (technical_certificate_id, description, quantity, unit, note, display_order) VALUES " + strings.Join(values, ",") + ";"
	}
	sql += " COMMIT;"
	_, err := a.runPSQL(ctx, sql)
	return err
}

func nullableCertificateQuantitySQL(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, ",") {
		value = strings.ReplaceAll(strings.ReplaceAll(value, ".", ""), ",", ".")
	}
	if value == "" {
		return "NULL"
	}
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return "NULL"
	}
	return sqlQuote(value) + "::numeric"
}

func (a *app) deleteTechnicalCertificate(w http.ResponseWriter, r *http.Request, id string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		UPDATE technical_certificates SET deleted_at = now(), updated_at = now()
		WHERE id = %s::uuid AND company_id = %s::uuid AND deleted_at IS NULL
		RETURNING id::text AS id;
	`, sqlQuote(id), sqlQuote(session.CompanyID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "atestado nao encontrado")
		return
	}
	logCertificateAudit(r.Context(), a, session, id, "deleted", "Atestado técnico removido do acervo da empresa.")
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

func (a *app) reprocessTechnicalCertificateOCR(w http.ResponseWriter, r *http.Request, id string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item) FROM (
			SELECT document_url AS path
			FROM technical_certificates
			WHERE id = %s::uuid AND company_id = %s::uuid AND deleted_at IS NULL
		) item;
	`, sqlQuote(id), sqlQuote(session.CompanyID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "atestado nao encontrado")
		return
	}
	var document struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(payload, &document); err != nil || strings.TrimSpace(document.Path) == "" {
		writeError(w, http.StatusInternalServerError, "arquivo do atestado nao encontrado")
		return
	}
	if _, err := os.Stat(document.Path); err != nil {
		writeError(w, http.StatusNotFound, "arquivo do atestado nao esta mais disponivel")
		return
	}
	text, status := extractTechnicalCertificateText(r.Context(), document.Path)
	textPath := ""
	if text != "" {
		var saveErr error
		textPath, saveErr = saveTechnicalCertificateText(document.Path, text)
		if saveErr != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel salvar o texto lido")
			return
		}
	}
	_, err = a.runPSQL(r.Context(), fmt.Sprintf(`
		UPDATE technical_certificates
		SET extracted_text = %s, extracted_text_file_path = NULLIF(%s, ''), extraction_status = %s, updated_at = now()
		WHERE id = %s::uuid AND company_id = %s::uuid AND deleted_at IS NULL;
	`, sqlQuote(text), sqlQuote(textPath), sqlQuote(status), sqlQuote(id), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar a leitura do atestado")
		return
	}
	logCertificateAudit(r.Context(), a, session, id, "ocr_reprocessed", "Leitura do atestado refeita automaticamente.")
	fullPayload, err := a.queryJSON(r.Context(), technicalCertificateSelectSQL(id, session.CompanyID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "leitura concluida, mas nao foi possivel carregar o atestado")
		return
	}
	writeRawJSON(w, http.StatusOK, fullPayload)
}

func (a *app) downloadTechnicalCertificateDocument(w http.ResponseWriter, r *http.Request, id string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item) FROM (
			SELECT document_url AS path, file_name AS "fileName", COALESCE(mime_type, 'application/pdf') AS "mimeType"
			FROM technical_certificates
			WHERE id = %s::uuid AND company_id = %s::uuid AND deleted_at IS NULL
		) item;
	`, sqlQuote(id), sqlQuote(session.CompanyID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "arquivo do atestado nao encontrado")
		return
	}
	var file struct {
		Path     string `json:"path"`
		FileName string `json:"fileName"`
		MimeType string `json:"mimeType"`
	}
	if err := json.Unmarshal(payload, &file); err != nil || file.Path == "" {
		writeError(w, http.StatusNotFound, "arquivo do atestado nao encontrado")
		return
	}
	if _, err := os.Stat(file.Path); err != nil {
		writeError(w, http.StatusNotFound, "arquivo do atestado nao esta mais disponivel")
		return
	}
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", file.FileName))
	http.ServeFile(w, r, file.Path)
}

func technicalCertificateSelectSQL(id, companyID string) string {
	return fmt.Sprintf(`
		SELECT row_to_json(item) FROM (
			SELECT tc.id::text AS id, COALESCE(tc.certificate_number, '') AS "certificateNumber", COALESCE(tc.issuer_name, '') AS "issuerName",
				COALESCE(tc.issuer_document, '') AS "issuerDocument", COALESCE(tc.contracted_name, '') AS "contractedName", COALESCE(tc.contract_number, '') AS "contractNumber", tc.object,
				COALESCE(tc.service_description, '') AS "serviceDescription", COALESCE(tc.state, '') AS state, COALESCE(tc.city, '') AS city,
				tc.execution_start AS "executionStart", tc.execution_end AS "executionEnd", tc.contract_value AS "contractValue",
				COALESCE(tc.technical_manager, '') AS "technicalManager", COALESCE(tc.professional_registration, '') AS "professionalRegistration",
				COALESCE(tc.art_cat_reference, '') AS "artCatReference", COALESCE(tc.cat_number, '') AS "catNumber", COALESCE(tc.technical_professional_id::text, '') AS "technicalProfessionalId", COALESCE(tc.cat_professional, '') AS "catProfessional", COALESCE(tc.professional_role, '') AS "professionalRole", tc.completion_status AS "completionStatus", tc.usage_scope AS "usageScope", COALESCE(tc.tags, '') AS tags, tc.file_name AS "fileName",
				COALESCE(tc.mime_type, '') AS "mimeType", tc.file_size AS "fileSize", tc.extracted_text AS "contentText", tc.extraction_status AS "extractionStatus",
				tc.status, tc.created_at AS "createdAt", tc.updated_at AS "updatedAt", COALESCE((SELECT json_agg(json_build_object('id', q.id::text, 'description', q.description, 'value', q.quantity, 'unit', COALESCE(q.unit, ''), 'note', COALESCE(q.note, '')) ORDER BY q.display_order) FROM technical_certificate_quantities q WHERE q.technical_certificate_id = tc.id), '[]'::json) AS quantities, '/api/technical-certificates/' || tc.id::text || '/document' AS "documentUrl"
			FROM technical_certificates tc
			WHERE tc.id = %s::uuid AND tc.company_id = %s::uuid AND tc.deleted_at IS NULL
		) item;
	`, sqlQuote(id), sqlQuote(companyID))
}

func saveTechnicalCertificateDocument(dataURL, originalName, mimeType string) (string, string, string, int, error) {
	header, payload, ok := strings.Cut(strings.TrimSpace(dataURL), ",")
	if !ok || !strings.HasPrefix(strings.ToLower(header), "data:") {
		return "", "", "", 0, errors.New("arquivo do atestado invalido")
	}
	originalName = strings.TrimSpace(filepath.Base(originalName))
	if strings.ToLower(filepath.Ext(originalName)) != ".pdf" {
		return "", "", "", 0, errors.New("envie o atestado no formato PDF")
	}
	content, err := base64.StdEncoding.DecodeString(payload)
	if err != nil || len(content) == 0 {
		return "", "", "", 0, errors.New("nao foi possivel ler o PDF do atestado")
	}
	if len(content) > 25*1024*1024 {
		return "", "", "", 0, errors.New("cada atestado pode ter no maximo 25 MB")
	}
	if mimeType == "" {
		mimeType = strings.TrimPrefix(strings.TrimSuffix(header, ";base64"), "data:")
	}
	dir := filepath.Join("private_uploads", "technical_certificates")
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", "", "", 0, errors.New("nao foi possivel preparar a pasta privada do acervo")
	}
	filePath := filepath.Join(dir, fmt.Sprintf("certificate-%d.pdf", time.Now().UnixNano()))
	if err := os.WriteFile(filePath, content, 0640); err != nil {
		return "", "", "", 0, errors.New("nao foi possivel salvar o PDF do atestado")
	}
	if len(originalName) > 220 {
		originalName = originalName[:220]
	}
	return filePath, originalName, mimeType, len(content), nil
}

func extractTechnicalCertificateText(ctx context.Context, filePath string) (string, string) {
	command, err := exec.LookPath("pdftotext")
	if err == nil {
		extractionCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		output, runErr := exec.CommandContext(extractionCtx, command, "-layout", filePath, "-").Output()
		cancel()
		text := strings.TrimSpace(string(output))
		if runErr == nil && text != "" {
			return text, "extracted"
		}
	}
	return extractTechnicalCertificateOCR(ctx, filePath)
}

func extractTechnicalCertificateOCR(ctx context.Context, filePath string) (string, string) {
	converter, converterErr := findPDFToImageCommand()
	tesseract, tesseractErr := findTesseractCommand()
	if converterErr != nil || tesseractErr != nil {
		return "", "pending_ocr"
	}
	tempDir, err := os.MkdirTemp("", "licitahub-ocr-*")
	if err != nil {
		return "", "failed"
	}
	defer os.RemoveAll(tempDir)
	prefix := filepath.Join(tempDir, "page")
	ocrCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if err := exec.CommandContext(ocrCtx, converter, "-r", "220", "-png", filePath, prefix).Run(); err != nil {
		return "", "failed"
	}
	pages, err := filepath.Glob(prefix + "-*.png")
	if err != nil || len(pages) == 0 {
		return "", "failed"
	}
	parts := make([]string, 0, len(pages))
	for _, page := range pages {
		output, runErr := exec.CommandContext(ocrCtx, tesseract, page, "stdout", "-l", "por").Output()
		if runErr != nil {
			continue
		}
		if text := strings.TrimSpace(string(output)); text != "" {
			parts = append(parts, text)
		}
	}
	if len(parts) == 0 {
		return "", "failed"
	}
	return strings.Join(parts, "\n\n"), "ocr_extracted"
}

// findPDFToImageCommand checks the portable converter shipped with the backend
// before relying on a machine-wide Poppler installation.
func findPDFToImageCommand() (string, error) {
	if command, err := exec.LookPath("pdftoppm"); err == nil {
		return command, nil
	}
	if executable, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(executable), "tools", "poppler", "Library", "bin", "pdftoppm.exe")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", errors.New("conversor de PDF nao encontrado")
}

// findTesseractCommand also supports the standard Windows installer location.
// That keeps OCR available even when Windows has not refreshed PATH yet.
func findTesseractCommand() (string, error) {
	if command, err := exec.LookPath("tesseract"); err == nil {
		return command, nil
	}
	for _, root := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramW6432"), os.Getenv("ProgramFiles(x86)")} {
		if strings.TrimSpace(root) == "" {
			continue
		}
		candidate := filepath.Join(root, "Tesseract-OCR", "tesseract.exe")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", errors.New("tesseract nao encontrado")
}

func saveTechnicalCertificateText(pdfPath, text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil
	}
	textPath := strings.TrimSuffix(pdfPath, filepath.Ext(pdfPath)) + ".txt"
	if err := os.WriteFile(textPath, []byte(text), 0640); err != nil {
		return "", err
	}
	return textPath, nil
}

func nullableCertificateDateSQL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "NULL"
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return "NULL"
	}
	return sqlQuote(value) + "::date"
}

func nullableCertificateValueSQL(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, ".", ""))
	value = strings.ReplaceAll(value, ",", ".")
	if value == "" {
		return "NULL"
	}
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return "NULL"
	}
	return sqlQuote(value) + "::numeric"
}

func logCertificateAudit(ctx context.Context, a *app, session sessionUser, certificateID, action, description string) {
	_, _ = a.runPSQL(ctx, fmt.Sprintf(`
		INSERT INTO audit_logs (actor_user_id, company_id, module, action, entity_type, entity_id, description)
		VALUES (%s::uuid, %s::uuid, 'technical_certificates', %s, 'technical_certificate', %s::uuid, %s);
	`, sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(action), sqlQuote(certificateID), sqlQuote(description)))
}

func ioLimitReader(body io.ReadCloser, limit int64) *io.LimitedReader {
	return &io.LimitedReader{R: body, N: limit}
}
