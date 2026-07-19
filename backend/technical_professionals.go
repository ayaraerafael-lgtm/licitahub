package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type technicalProfessionalRequest struct {
	FullName                 string                                  `json:"fullName"`
	Formation                string                                  `json:"formation"`
	ComplementaryEducation   string                                  `json:"complementaryEducation"`
	ProfessionalRegistration string                                  `json:"professionalRegistration"`
	RoleTitle                string                                  `json:"roleTitle"`
	Phone                    string                                  `json:"phone"`
	Email                    string                                  `json:"email"`
	State                    string                                  `json:"state"`
	AvailabilityStatus       string                                  `json:"availabilityStatus"`
	ProfileSummary           string                                  `json:"profileSummary"`
	Status                   string                                  `json:"status"`
	Educations               []technicalProfessionalEducationRequest `json:"educations"`
}

type technicalProfessionalEducationRequest struct {
	Level       string `json:"level"`
	CourseName  string `json:"courseName"`
	Institution string `json:"institution"`
}

func (a *app) handleTechnicalProfessionals(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "os profissionais tecnicos estao disponiveis somente para empresas associadas")
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.listTechnicalProfessionals(w, r, session)
	case http.MethodPost:
		if !session.canManageTechnicalCertificates() {
			writeError(w, http.StatusForbidden, "seu perfil pode consultar, mas nao cadastrar profissionais")
			return
		}
		a.createTechnicalProfessional(w, r, session)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleTechnicalProfessionalByPath(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "profissional tecnico nao disponivel")
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/technical-professionals/"), "/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "profissional nao informado")
		return
	}
	switch r.Method {
	case http.MethodPut:
		if !session.canManageTechnicalCertificates() {
			writeError(w, http.StatusForbidden, "seu perfil nao pode alterar profissionais")
			return
		}
		a.updateTechnicalProfessional(w, r, id, session)
	case http.MethodDelete:
		if !session.canManageTechnicalCertificates() {
			writeError(w, http.StatusForbidden, "seu perfil nao pode remover profissionais")
			return
		}
		a.deleteTechnicalProfessional(w, r, id, session)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) listTechnicalProfessionals(w http.ResponseWriter, r *http.Request, session sessionUser) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && status != "active" && status != "archived" {
		writeError(w, http.StatusBadRequest, "status de profissional invalido")
		return
	}
	statusFilter := ""
	if status != "" {
		statusFilter = fmt.Sprintf("AND p.status = %s", sqlQuote(status))
	}
	searchFilter := ""
	if search != "" {
		needle := sqlQuote("%" + search + "%")
		searchFilter = fmt.Sprintf("AND (concat_ws(' ', p.full_name, p.formation, p.complementary_education, p.professional_registration, p.role_title, p.email, p.profile_summary) ILIKE %s OR EXISTS (SELECT 1 FROM technical_professional_educations e WHERE e.technical_professional_id = p.id AND concat_ws(' ', e.education_level, e.course_name, e.institution) ILIKE %s))", needle, needle)
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json) FROM (
			SELECT p.id::text AS id, p.full_name AS "fullName", COALESCE(p.formation, '') AS formation,
				COALESCE(p.complementary_education, '') AS "complementaryEducation", COALESCE(p.professional_registration, '') AS "professionalRegistration",
				COALESCE(p.role_title, '') AS "roleTitle", COALESCE(p.phone, '') AS phone, COALESCE(p.email, '') AS email,
				COALESCE(p.state, '') AS state, p.availability_status AS "availabilityStatus", COALESCE(p.profile_summary, '') AS "profileSummary",
				p.status, p.created_at AS "createdAt", p.updated_at AS "updatedAt",
				COALESCE((SELECT json_agg(json_build_object('id', e.id::text, 'level', e.education_level, 'courseName', e.course_name, 'institution', COALESCE(e.institution, '')) ORDER BY e.display_order) FROM technical_professional_educations e WHERE e.technical_professional_id = p.id), '[]'::json) AS educations,
				(SELECT count(*) FROM technical_certificates tc WHERE tc.technical_professional_id = p.id AND tc.deleted_at IS NULL) AS "certificateCount"
			FROM technical_professionals p
			WHERE p.company_id = %s::uuid AND p.deleted_at IS NULL %s %s
			ORDER BY p.full_name ASC
		) item;
	`, sqlQuote(session.CompanyID), statusFilter, searchFilter))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar os profissionais tecnicos")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func validateTechnicalProfessional(req *technicalProfessionalRequest) error {
	req.FullName = strings.TrimSpace(req.FullName)
	if req.FullName == "" {
		return fmt.Errorf("nome do profissional e obrigatorio")
	}
	if req.AvailabilityStatus == "" {
		req.AvailabilityStatus = "available"
	}
	if req.AvailabilityStatus != "available" && req.AvailabilityStatus != "limited" && req.AvailabilityStatus != "unavailable" {
		return fmt.Errorf("disponibilidade do profissional invalida")
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Status != "active" && req.Status != "archived" {
		return fmt.Errorf("status do profissional invalido")
	}
	if len(req.FullName) > 255 {
		return fmt.Errorf("nome do profissional muito longo")
	}
	for _, education := range req.Educations {
		if strings.TrimSpace(education.CourseName) == "" {
			continue
		}
		if !validTechnicalEducationLevel(education.Level) {
			return fmt.Errorf("tipo de formacao invalido")
		}
	}
	return nil
}

func professionalValues(req technicalProfessionalRequest) []any {
	return []any{sqlQuote(req.FullName), sqlQuote(req.Formation), sqlQuote(req.ComplementaryEducation), sqlQuote(req.ProfessionalRegistration), sqlQuote(req.RoleTitle), sqlQuote(req.Phone), sqlQuote(strings.ToLower(strings.TrimSpace(req.Email))), sqlQuote(normalizeState(req.State)), sqlQuote(req.AvailabilityStatus), sqlQuote(req.ProfileSummary), sqlQuote(req.Status)}
}

func (a *app) createTechnicalProfessional(w http.ResponseWriter, r *http.Request, session sessionUser) {
	var req technicalProfessionalRequest
	if err := json.NewDecoder(ioLimitReader(r.Body, 2*1024*1024)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dados do profissional invalidos")
		return
	}
	if err := validateTechnicalProfessional(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	v := professionalValues(req)
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH inserted AS (
			INSERT INTO technical_professionals (company_id, full_name, formation, complementary_education, professional_registration, role_title, phone, email, state, availability_status, profile_summary, status, created_by_user_id)
			VALUES (%s::uuid, NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''), %s, NULLIF(%s, ''), %s, %s::uuid)
			RETURNING id::text AS id
		) SELECT row_to_json(inserted) FROM inserted;
	`, sqlQuote(session.CompanyID), v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10], sqlQuote(session.UserID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusInternalServerError, "nao foi possivel cadastrar o profissional")
		return
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(payload, &created)
	if created.ID == "" || a.replaceTechnicalProfessionalEducations(r.Context(), created.ID, req.Educations) != nil {
		writeError(w, http.StatusInternalServerError, "profissional cadastrado, mas nao foi possivel salvar as formacoes")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) updateTechnicalProfessional(w http.ResponseWriter, r *http.Request, id string, session sessionUser) {
	var req technicalProfessionalRequest
	if err := json.NewDecoder(ioLimitReader(r.Body, 2*1024*1024)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "dados do profissional invalidos")
		return
	}
	if err := validateTechnicalProfessional(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	v := professionalValues(req)
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		UPDATE technical_professionals SET full_name = NULLIF(%s, ''), formation = NULLIF(%s, ''), complementary_education = NULLIF(%s, ''), professional_registration = NULLIF(%s, ''), role_title = NULLIF(%s, ''), phone = NULLIF(%s, ''), email = NULLIF(%s, ''), state = NULLIF(%s, ''), availability_status = %s, profile_summary = NULLIF(%s, ''), status = %s, updated_at = now()
		WHERE id = %s::uuid AND company_id = %s::uuid AND deleted_at IS NULL RETURNING id::text AS id;
	`, v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10], sqlQuote(id), sqlQuote(session.CompanyID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "profissional nao encontrado")
		return
	}
	if err := a.replaceTechnicalProfessionalEducations(r.Context(), id, req.Educations); err != nil {
		writeError(w, http.StatusInternalServerError, "profissional atualizado, mas nao foi possivel salvar as formacoes")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func validTechnicalEducationLevel(level string) bool {
	switch strings.TrimSpace(level) {
	case "graduation", "specialization", "mba", "masters", "doctorate", "postdoctorate", "extension", "certification", "other":
		return true
	default:
		return false
	}
}

func (a *app) replaceTechnicalProfessionalEducations(ctx context.Context, professionalID string, educations []technicalProfessionalEducationRequest) error {
	values := make([]string, 0, len(educations))
	for index, education := range educations {
		courseName := strings.TrimSpace(education.CourseName)
		if courseName == "" {
			continue
		}
		if !validTechnicalEducationLevel(education.Level) {
			return fmt.Errorf("tipo de formacao invalido")
		}
		values = append(values, fmt.Sprintf("(%s::uuid, %s, %s, NULLIF(%s, ''), %d)", sqlQuote(professionalID), sqlQuote(education.Level), sqlQuote(courseName), sqlQuote(strings.TrimSpace(education.Institution)), index))
	}
	sql := fmt.Sprintf("BEGIN; DELETE FROM technical_professional_educations WHERE technical_professional_id = %s::uuid;", sqlQuote(professionalID))
	if len(values) > 0 {
		sql += " INSERT INTO technical_professional_educations (technical_professional_id, education_level, course_name, institution, display_order) VALUES " + strings.Join(values, ",") + ";"
	}
	sql += " COMMIT;"
	_, err := a.runPSQL(ctx, sql)
	return err
}

func (a *app) deleteTechnicalProfessional(w http.ResponseWriter, r *http.Request, id string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`UPDATE technical_professionals SET deleted_at = now(), updated_at = now() WHERE id = %s::uuid AND company_id = %s::uuid AND deleted_at IS NULL RETURNING id::text AS id;`, sqlQuote(id), sqlQuote(session.CompanyID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "profissional nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}
