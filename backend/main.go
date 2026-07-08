package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type app struct {
	psqlPath string
	pg       postgresConfig
}

type postgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type createNewsRequest struct {
	Title             string `json:"title"`
	CategoryID        string `json:"categoryId"`
	CategorySlug      string `json:"categorySlug"`
	Status            string `json:"status"`
	Summary           string `json:"summary"`
	Content           string `json:"content"`
	MainImageURL      string `json:"mainImageUrl"`
	MainImageDataURL  string `json:"mainImageDataUrl"`
	MainImageFileName string `json:"mainImageFileName"`
	MainImageMimeType string `json:"mainImageMimeType"`
}

type createInvitationRequest struct {
	TradeName    string `json:"tradeName"`
	CNPJ         string `json:"cnpj"`
	ContactName  string `json:"contactName"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	State        string `json:"state"`
	InternalNote string `json:"internalNote"`
}

type acceptInvitationRequest struct {
	Token                    string `json:"token"`
	InvitationID             string `json:"invitationId"`
	Website                  string `json:"website"`
	InstitutionalDescription string `json:"institutionalDescription"`
	City                     string `json:"city"`
	State                    string `json:"state"`
	AdminFullName            string `json:"adminFullName"`
	AdminEmail               string `json:"adminEmail"`
	AdminPhone               string `json:"adminPhone"`
	AdminJobTitle            string `json:"adminJobTitle"`
	Password                 string `json:"password"`
}

type createCompanyUserRequest struct {
	CompanyID        string `json:"companyId"`
	FullName         string `json:"fullName"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	JobTitle         string `json:"jobTitle"`
	AccessProfileKey string `json:"accessProfileKey"`
}

type updateCompanyUserRequest struct {
	FullName         string `json:"fullName"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	JobTitle         string `json:"jobTitle"`
	AccessProfileKey string `json:"accessProfileKey"`
	Status           string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	application := &app{
		psqlPath: getenv("PSQL_PATH", `C:\Program Files\PostgreSQL\17\bin\psql.exe`),
		pg: postgresConfig{
			Host:     getenv("PGHOST", "localhost"),
			Port:     getenv("PGPORT", "5432"),
			User:     getenv("PGUSER", "postgres"),
			Password: os.Getenv("PGPASSWORD"),
			Database: getenv("PGDATABASE", "licitahub_dev"),
		},
	}

	if err := application.checkDatabase(context.Background()); err != nil {
		log.Fatal(err)
	}
	if err := application.ensureDatabaseMigrations(context.Background()); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", application.handleHealth)
	mux.HandleFunc("/api/access-profiles", application.handleAccessProfiles)
	mux.HandleFunc("/api/company-invitations", application.handleCompanyInvitations)
	mux.HandleFunc("/api/company-invitations/", application.handleCompanyInvitationByPath)
	mux.HandleFunc("/api/companies", application.handleCompanies)
	mux.HandleFunc("/api/company-users", application.handleCompanyUsers)
	mux.HandleFunc("/api/company-users/", application.handleCompanyUserByPath)
	mux.HandleFunc("/api/news/categories", application.handleNewsCategories)
	mux.HandleFunc("/api/news", application.handleNews)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	mux.Handle("/", http.FileServer(http.Dir("..")))

	port := getenv("APP_PORT", "8080")
	log.Printf("LicitaHub API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, withCORS(mux)))
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "licitahub-api",
	})
}

func (a *app) handleAccessProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}

	payload, err := a.queryJSON(r.Context(), `
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT id::text AS id, key, name, COALESCE(description, '') AS description
			FROM access_profiles
			ORDER BY
				CASE key
					WHEN 'company_admin' THEN 1
					WHEN 'commercial' THEN 2
					WHEN 'technical' THEN 3
					WHEN 'reader' THEN 4
					ELSE 9
				END,
				name
		) item;
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar os perfis de acesso")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleCompanyInvitations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listCompanyInvitations(w, r)
	case http.MethodPost:
		a.createCompanyInvitation(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleCompanyInvitationByPath(w http.ResponseWriter, r *http.Request) {
	id, action := splitResourcePath(r.URL.Path, "/api/company-invitations/")
	if id == "" {
		writeError(w, http.StatusNotFound, "convite nao encontrado")
		return
	}

	if action == "" && r.Method == http.MethodGet {
		a.getCompanyInvitation(w, r, id)
		return
	}

	if action == "accept" && r.Method == http.MethodPost {
		a.acceptCompanyInvitation(w, r, id)
		return
	}

	if action == "cancel" && r.Method == http.MethodPatch {
		a.changeCompanyInvitationStatus(w, r, id, "cancelled")
		return
	}

	writeError(w, http.StatusNotFound, "rota do convite nao encontrada")
}

func (a *app) getCompanyInvitation(w http.ResponseWriter, r *http.Request, pathInvitationID string) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	filter := fmt.Sprintf("i.id = %s::uuid", sqlQuote(pathInvitationID))
	if pathInvitationID == "by-token" && token != "" {
		filter = fmt.Sprintf("i.invitation_token = %s", sqlQuote(token))
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				i.id::text AS id,
				i.company_id::text AS "companyId",
				i.trade_name AS "tradeName",
				i.cnpj,
				i.contact_name AS "contactName",
				i.email,
				i.phone,
				COALESCE(i.state, '') AS state,
				COALESCE(i.internal_note, '') AS "internalNote",
				i.status,
				i.invitation_token AS token,
				i.sent_at AS "sentAt",
				i.accepted_at AS "acceptedAt",
				i.expires_at AS "expiresAt",
				i.created_at AS "createdAt",
				i.updated_at AS "updatedAt"
			FROM company_invitations i
			WHERE %s
			LIMIT 1
		) item;
	`, filter))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o convite")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) listCompanyInvitations(w http.ResponseWriter, r *http.Request) {
	payload, err := a.queryJSON(r.Context(), `
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				i.id::text AS id,
				i.company_id::text AS "companyId",
				i.trade_name AS "tradeName",
				i.cnpj,
				i.contact_name AS "contactName",
				i.email,
				i.phone,
				COALESCE(i.state, '') AS state,
				COALESCE(i.internal_note, '') AS "internalNote",
				i.status,
				i.invitation_token AS token,
				i.sent_at AS "sentAt",
				i.accepted_at AS "acceptedAt",
				i.expires_at AS "expiresAt",
				i.created_at AS "createdAt",
				i.updated_at AS "updatedAt"
			FROM company_invitations i
			ORDER BY i.created_at DESC
		) item;
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar os convites")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createCompanyInvitation(w http.ResponseWriter, r *http.Request) {
	var req createInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.TradeName = strings.TrimSpace(req.TradeName)
	req.CNPJ = normalizeCNPJ(req.CNPJ)
	req.ContactName = strings.TrimSpace(req.ContactName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.State = normalizeState(req.State)
	req.InternalNote = strings.TrimSpace(req.InternalNote)

	if req.TradeName == "" {
		writeError(w, http.StatusBadRequest, "nome fantasia e obrigatorio")
		return
	}
	if req.CNPJ == "" {
		writeError(w, http.StatusBadRequest, "cnpj e obrigatorio")
		return
	}
	if req.ContactName == "" {
		writeError(w, http.StatusBadRequest, "nome do responsavel e obrigatorio")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email e obrigatorio")
		return
	}
	if req.Phone == "" {
		writeError(w, http.StatusBadRequest, "telefone e obrigatorio")
		return
	}

	token, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel gerar o token do convite")
		return
	}

	payload, err := a.queryJSON(r.Context(), buildCreateCompanyInvitationSQL(req, token))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel criar o convite: "+humanizeConstraintError(err.Error()))
		return
	}

	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) acceptCompanyInvitation(w http.ResponseWriter, r *http.Request, pathInvitationID string) {
	var req acceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.InvitationID = strings.TrimSpace(req.InvitationID)
	if req.InvitationID == "" {
		req.InvitationID = pathInvitationID
	}
	req.Token = strings.TrimSpace(req.Token)
	req.Website = strings.TrimSpace(req.Website)
	req.InstitutionalDescription = strings.TrimSpace(req.InstitutionalDescription)
	req.City = strings.TrimSpace(req.City)
	req.State = normalizeState(req.State)
	req.AdminFullName = strings.TrimSpace(req.AdminFullName)
	req.AdminEmail = strings.TrimSpace(strings.ToLower(req.AdminEmail))
	req.AdminPhone = strings.TrimSpace(req.AdminPhone)
	req.AdminJobTitle = strings.TrimSpace(req.AdminJobTitle)
	req.Password = strings.TrimSpace(req.Password)

	if pathInvitationID == "by-token" {
		req.InvitationID = ""
	}

	if req.InvitationID == "" && req.Token == "" {
		writeError(w, http.StatusBadRequest, "convite ou token e obrigatorio")
		return
	}
	if req.AdminFullName == "" {
		writeError(w, http.StatusBadRequest, "nome do administrador e obrigatorio")
		return
	}
	if req.AdminEmail == "" {
		writeError(w, http.StatusBadRequest, "email do administrador e obrigatorio")
		return
	}
	if req.AdminPhone == "" {
		writeError(w, http.StatusBadRequest, "telefone do administrador e obrigatorio")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "senha deve ter pelo menos 8 caracteres")
		return
	}

	payload, err := a.queryJSON(r.Context(), buildAcceptCompanyInvitationSQL(req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel aceitar o convite: "+humanizeConstraintError(err.Error()))
		return
	}

	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "convite nao encontrado, expirado ou ja utilizado")
		return
	}

	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) changeCompanyInvitationStatus(w http.ResponseWriter, r *http.Request, invitationID string, status string) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		UPDATE company_invitations
		SET status = %s
		WHERE id = %s::uuid
		RETURNING row_to_json(company_invitations);
	`, sqlQuote(status), sqlQuote(invitationID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o convite")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleCompanies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}

	payload, err := a.queryJSON(r.Context(), `
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				c.id::text AS id,
				c.trade_name AS "tradeName",
				c.cnpj,
				c.status,
				COALESCE(c.main_contact_name, '') AS "mainContactName",
				COALESCE(c.main_contact_email, '') AS "mainContactEmail",
				COALESCE(c.main_contact_phone, '') AS "mainContactPhone",
				COALESCE(c.state, '') AS state,
				COALESCE(c.city, '') AS city,
				COALESCE(p.website, '') AS website,
				COALESCE(p.institutional_description, '') AS "institutionalDescription",
				c.created_at AS "createdAt",
				c.updated_at AS "updatedAt"
			FROM companies c
			LEFT JOIN company_profiles p ON p.company_id = c.id
			WHERE c.deleted_at IS NULL
			ORDER BY c.created_at DESC
		) item;
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar as empresas")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleCompanyUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listCompanyUsers(w, r)
	case http.MethodPost:
		a.createCompanyUser(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleCompanyUserByPath(w http.ResponseWriter, r *http.Request) {
	userID, action := splitResourcePath(r.URL.Path, "/api/company-users/")
	if userID == "" {
		writeError(w, http.StatusNotFound, "usuario nao encontrado")
		return
	}

	switch {
	case action == "" && r.Method == http.MethodPut:
		a.updateCompanyUser(w, r, userID)
	case action == "block" && r.Method == http.MethodPatch:
		a.changeCompanyUserStatus(w, r, userID, "blocked")
	case action == "unblock" && r.Method == http.MethodPatch:
		a.changeCompanyUserStatus(w, r, userID, "active")
	case action == "remove" && (r.Method == http.MethodPatch || r.Method == http.MethodDelete):
		a.changeCompanyUserStatus(w, r, userID, "removed")
	default:
		writeError(w, http.StatusNotFound, "rota do usuario nao encontrada")
	}
}

func (a *app) listCompanyUsers(w http.ResponseWriter, r *http.Request) {
	companyID := strings.TrimSpace(r.URL.Query().Get("companyId"))
	filter := "u.company_id IS NOT NULL"
	if companyID != "" {
		filter = fmt.Sprintf("u.company_id = %s::uuid", sqlQuote(companyID))
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				u.id::text AS id,
				u.company_id::text AS "companyId",
				c.trade_name AS "companyTradeName",
				u.full_name AS "fullName",
				u.email,
				COALESCE(u.phone, '') AS phone,
				COALESCE(u.job_title, '') AS "jobTitle",
				p.id::text AS "accessProfileId",
				p.key AS "accessProfileKey",
				p.name AS "accessProfileName",
				u.status,
				u.created_at AS "createdAt",
				u.updated_at AS "updatedAt"
			FROM users u
			LEFT JOIN companies c ON c.id = u.company_id
			LEFT JOIN access_profiles p ON p.id = u.access_profile_id
			WHERE %s
			  AND u.deleted_at IS NULL
			  AND u.status <> 'removed'
			ORDER BY u.created_at DESC
		) item;
	`, filter))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar os usuarios")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createCompanyUser(w http.ResponseWriter, r *http.Request) {
	var req createCompanyUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.CompanyID = strings.TrimSpace(req.CompanyID)
	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.JobTitle = strings.TrimSpace(req.JobTitle)
	req.AccessProfileKey = normalizeAccessProfileKey(req.AccessProfileKey)

	if req.CompanyID == "" {
		writeError(w, http.StatusBadRequest, "empresa e obrigatoria")
		return
	}
	if req.FullName == "" {
		writeError(w, http.StatusBadRequest, "nome do usuario e obrigatorio")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email do usuario e obrigatorio")
		return
	}
	if req.Phone == "" {
		writeError(w, http.StatusBadRequest, "telefone do usuario e obrigatorio")
		return
	}
	if req.AccessProfileKey == "" {
		req.AccessProfileKey = "commercial"
	}

	payload, err := a.queryJSON(r.Context(), buildCreateCompanyUserSQL(req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel criar o usuario: "+humanizeConstraintError(err.Error()))
		return
	}

	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) updateCompanyUser(w http.ResponseWriter, r *http.Request, userID string) {
	var req updateCompanyUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.JobTitle = strings.TrimSpace(req.JobTitle)
	req.AccessProfileKey = normalizeAccessProfileKey(req.AccessProfileKey)
	req.Status = normalizeUserStatus(req.Status)

	if req.FullName == "" {
		writeError(w, http.StatusBadRequest, "nome do usuario e obrigatorio")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email do usuario e obrigatorio")
		return
	}
	if req.Phone == "" {
		writeError(w, http.StatusBadRequest, "telefone do usuario e obrigatorio")
		return
	}
	if req.AccessProfileKey == "" {
		req.AccessProfileKey = "commercial"
	}
	if req.Status == "" {
		req.Status = "active"
	}

	payload, err := a.queryJSON(r.Context(), buildUpdateCompanyUserSQL(userID, req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o usuario: "+humanizeConstraintError(err.Error()))
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) changeCompanyUserStatus(w http.ResponseWriter, r *http.Request, userID string, status string) {
	status = normalizeUserStatus(status)
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		UPDATE users
		SET
			status = %s,
			blocked_at = CASE WHEN %s = 'blocked' THEN now() ELSE NULL END,
			removed_at = CASE WHEN %s = 'removed' THEN now() ELSE removed_at END
		WHERE id = %s::uuid
		  AND deleted_at IS NULL
		RETURNING row_to_json(users);
	`, sqlQuote(status), sqlQuote(status), sqlQuote(status), sqlQuote(userID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o status do usuario")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleNewsCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}

	sql := `
		SELECT COALESCE(json_agg(row_to_json(category)), '[]'::json)
		FROM (
			SELECT id::text AS id, name, slug
			FROM news_categories
			WHERE is_active = true
			ORDER BY name
		) category;
	`

	payload, err := a.queryJSON(r.Context(), sql)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar as categorias")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleNews(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listNews(w, r)
	case http.MethodPost:
		a.createNews(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) listNews(w http.ResponseWriter, r *http.Request) {
	sql := `
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				n.id::text AS id,
				n.title,
				c.id::text AS "categoryId",
				c.name AS "categoryName",
				c.slug AS "categorySlug",
				n.status,
				COALESCE(n.summary, '') AS summary,
				COALESCE(n.content, '') AS content,
				COALESCE(m.file_url, '') AS "mainImageUrl",
				n.published_at AS "publishedAt",
				n.created_at AS "createdAt",
				n.updated_at AS "updatedAt"
			FROM news n
			LEFT JOIN news_categories c ON c.id = n.category_id
			LEFT JOIN media_files m ON m.id = n.main_image_media_id
			WHERE n.deleted_at IS NULL
			  AND n.status IN ('published', 'featured')
			ORDER BY COALESCE(n.published_at, n.created_at) DESC
		) item;
	`

	payload, err := a.queryJSON(r.Context(), sql)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar as noticias")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createNews(w http.ResponseWriter, r *http.Request) {
	var req createNewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.CategoryID = strings.TrimSpace(req.CategoryID)
	req.CategorySlug = strings.TrimSpace(strings.ToLower(req.CategorySlug))
	req.Summary = strings.TrimSpace(req.Summary)
	req.Content = strings.TrimSpace(req.Content)
	req.MainImageURL = strings.TrimSpace(req.MainImageURL)
	req.MainImageDataURL = strings.TrimSpace(req.MainImageDataURL)
	req.MainImageFileName = strings.TrimSpace(req.MainImageFileName)
	req.MainImageMimeType = strings.TrimSpace(req.MainImageMimeType)

	status, err := normalizeNewsStatus(req.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "titulo e obrigatorio")
		return
	}

	if req.CategoryID == "" && req.CategorySlug == "" {
		writeError(w, http.StatusBadRequest, "categoria e obrigatoria")
		return
	}

	if req.MainImageDataURL != "" {
		imageURL, err := saveNewsImage(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.MainImageURL = imageURL
	}

	payload, err := a.queryJSON(r.Context(), buildCreateNewsSQL(req, status))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel cadastrar a noticia: "+err.Error())
		return
	}

	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) checkDatabase(ctx context.Context) error {
	if _, err := os.Stat(a.psqlPath); err != nil {
		return fmt.Errorf("psql nao encontrado em %s", a.psqlPath)
	}

	if a.pg.Password == "" {
		return errors.New("PGPASSWORD nao informado")
	}

	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	_, err := a.runPSQL(ctx, "SELECT 1;")
	return err
}

func (a *app) ensureDatabaseMigrations(ctx context.Context) error {
	_, err := a.runPSQL(ctx, `
		ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash text;
	`)
	return err
}

func (a *app) queryJSON(ctx context.Context, sql string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	output, err := a.runPSQL(ctx, sql)
	if err != nil {
		return nil, err
	}

	payload := strings.TrimSpace(output)
	payload = strings.TrimPrefix(payload, "SET\n")
	payload = strings.TrimPrefix(payload, "SET\r\n")
	payload = strings.TrimSpace(payload)
	payload = firstJSONPayload(payload)
	if payload == "" {
		payload = "null"
	}
	if !strings.HasPrefix(payload, "{") && !strings.HasPrefix(payload, "[") && payload != "null" {
		return nil, errors.New(payload)
	}

	return []byte(payload), nil
}

func (a *app) runPSQL(ctx context.Context, sql string) (string, error) {
	tmp, err := os.CreateTemp("", "licitahub-*.sql")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.WriteString("SET client_encoding = 'UTF8';\n" + sql); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(
		ctx,
		a.psqlPath,
		"-h", a.pg.Host,
		"-p", a.pg.Port,
		"-U", a.pg.User,
		"-d", a.pg.Database,
		"-tA",
		"-v", "ON_ERROR_STOP=1",
		"-f", tmpPath,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+a.pg.Password, "PGCLIENTENCODING=UTF8")

	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		log.Printf("psql error: %s", message)
		return "", fmt.Errorf("%w: %s", err, message)
	}

	return string(output), nil
}

func firstJSONPayload(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return ""
	}

	start := strings.IndexAny(output, "[{")
	if start < 0 {
		return output
	}

	inString := false
	escaped := false
	depth := 0
	var opening rune
	var closing rune

	for index, char := range output[start:] {
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if char == '"' {
				inString = false
			}
			continue
		}

		if char == '"' {
			inString = true
			continue
		}

		if char == '{' || char == '[' {
			if depth == 0 {
				opening = char
				if opening == '{' {
					closing = '}'
				} else {
					closing = ']'
				}
			}
			depth++
			continue
		}

		if depth > 0 && char == closing {
			depth--
			if depth == 0 {
				return strings.TrimSpace(output[start : start+index+1])
			}
		}
	}

	return output[start:]
}

func buildCreateCompanyInvitationSQL(req createInvitationRequest, token string) string {
	return fmt.Sprintf(`
		INSERT INTO company_invitations (
			trade_name,
			cnpj,
			contact_name,
			email,
			phone,
			state,
			internal_note,
			status,
			invitation_token,
			sent_at,
			expires_at
		)
		VALUES (
			%s,
			%s,
			%s,
			%s,
			%s,
			NULLIF(%s, ''),
			NULLIF(%s, ''),
			'sent',
			%s,
			now(),
			now() + interval '30 days'
		)
		RETURNING row_to_json(company_invitations);
	`,
		sqlQuote(req.TradeName),
		sqlQuote(req.CNPJ),
		sqlQuote(req.ContactName),
		sqlQuote(req.Email),
		sqlQuote(req.Phone),
		sqlQuote(req.State),
		sqlQuote(req.InternalNote),
		sqlQuote(token),
	)
}

func buildAcceptCompanyInvitationSQL(req acceptInvitationRequest) string {
	filter := fmt.Sprintf("i.id = %s::uuid", sqlQuote(req.InvitationID))
	if req.InvitationID == "" {
		filter = fmt.Sprintf("i.invitation_token = %s", sqlQuote(req.Token))
	}
	if req.InvitationID != "" && req.Token != "" {
		filter = fmt.Sprintf("i.id = %s::uuid AND i.invitation_token = %s", sqlQuote(req.InvitationID), sqlQuote(req.Token))
	}

	return fmt.Sprintf(`
		WITH selected_invitation AS (
			SELECT i.*
			FROM company_invitations i
			WHERE %s
			  AND i.status = 'sent'
			  AND (i.expires_at IS NULL OR i.expires_at > now())
			LIMIT 1
		),
		inserted_company AS (
			INSERT INTO companies (
				trade_name,
				cnpj,
				status,
				main_contact_name,
				main_contact_email,
				main_contact_phone,
				state,
				city
			)
			SELECT
				i.trade_name,
				i.cnpj,
				'pending_review',
				%s,
				%s,
				%s,
				COALESCE(NULLIF(%s, ''), i.state),
				NULLIF(%s, '')
			FROM selected_invitation i
			RETURNING *
		),
		upsert_profile AS (
			INSERT INTO company_profiles (
				company_id,
				website,
				institutional_description,
				state,
				city,
				public_profile_slug
			)
			SELECT
				c.id,
				NULLIF(%s, ''),
				NULLIF(%s, ''),
				c.state,
				c.city,
				lower(regexp_replace(c.trade_name, '[^a-zA-Z0-9]+', '-', 'g'))
			FROM inserted_company c
			ON CONFLICT (company_id) DO UPDATE
			SET
				website = EXCLUDED.website,
				institutional_description = EXCLUDED.institutional_description,
				state = EXCLUDED.state,
				city = EXCLUDED.city
			RETURNING *
		),
		selected_profile AS (
			SELECT id FROM access_profiles WHERE key = 'company_admin' LIMIT 1
		),
		inserted_user AS (
			INSERT INTO users (
				company_id,
				access_profile_id,
				full_name,
				email,
				phone,
				job_title,
				password_hash,
				status
			)
			SELECT
				c.id,
				p.id,
				%s,
				%s,
				%s,
				NULLIF(%s, ''),
				%s,
				'active'
			FROM inserted_company c
			CROSS JOIN selected_profile p
			RETURNING *
		),
		updated_invitation AS (
			UPDATE company_invitations i
			SET
				company_id = c.id,
				status = 'pending_review',
				accepted_at = now()
			FROM inserted_company c, selected_invitation si
			WHERE i.id = si.id
			RETURNING i.*
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				c.id::text AS "companyId",
				c.trade_name AS "tradeName",
				c.cnpj,
				c.status AS "companyStatus",
				u.id::text AS "adminUserId",
				u.full_name AS "adminFullName",
				u.email AS "adminEmail",
				i.id::text AS "invitationId",
				i.status AS "invitationStatus",
				i.accepted_at AS "acceptedAt"
			FROM inserted_company c
			JOIN inserted_user u ON u.company_id = c.id
			JOIN updated_invitation i ON i.company_id = c.id
		) item;
	`,
		filter,
		sqlQuote(req.AdminFullName),
		sqlQuote(req.AdminEmail),
		sqlQuote(req.AdminPhone),
		sqlQuote(req.State),
		sqlQuote(req.City),
		sqlQuote(req.Website),
		sqlQuote(req.InstitutionalDescription),
		sqlQuote(req.AdminFullName),
		sqlQuote(req.AdminEmail),
		sqlQuote(req.AdminPhone),
		sqlQuote(req.AdminJobTitle),
		sqlQuote(hashPassword(req.Password)),
	)
}

func buildCreateCompanyUserSQL(req createCompanyUserRequest) string {
	return fmt.Sprintf(`
		WITH selected_profile AS (
			SELECT id FROM access_profiles WHERE key = %s LIMIT 1
		),
		inserted_user AS (
			INSERT INTO users (
				company_id,
				access_profile_id,
				full_name,
				email,
				phone,
				job_title,
				status
			)
			SELECT
				%s::uuid,
				p.id,
				%s,
				%s,
				%s,
				NULLIF(%s, ''),
				'pending_invite'
			FROM selected_profile p
			RETURNING *
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				u.id::text AS id,
				u.company_id::text AS "companyId",
				u.full_name AS "fullName",
				u.email,
				COALESCE(u.phone, '') AS phone,
				COALESCE(u.job_title, '') AS "jobTitle",
				p.key AS "accessProfileKey",
				p.name AS "accessProfileName",
				u.status,
				u.created_at AS "createdAt",
				u.updated_at AS "updatedAt"
			FROM inserted_user u
			JOIN access_profiles p ON p.id = u.access_profile_id
		) item;
	`,
		sqlQuote(req.AccessProfileKey),
		sqlQuote(req.CompanyID),
		sqlQuote(req.FullName),
		sqlQuote(req.Email),
		sqlQuote(req.Phone),
		sqlQuote(req.JobTitle),
	)
}

func buildUpdateCompanyUserSQL(userID string, req updateCompanyUserRequest) string {
	return fmt.Sprintf(`
		WITH selected_profile AS (
			SELECT id FROM access_profiles WHERE key = %s LIMIT 1
		),
		updated_user AS (
			UPDATE users u
			SET
				access_profile_id = p.id,
				full_name = %s,
				email = %s,
				phone = %s,
				job_title = NULLIF(%s, ''),
				status = %s,
				blocked_at = CASE WHEN %s = 'blocked' THEN COALESCE(u.blocked_at, now()) ELSE NULL END,
				removed_at = CASE WHEN %s = 'removed' THEN COALESCE(u.removed_at, now()) ELSE NULL END
			FROM selected_profile p
			WHERE u.id = %s::uuid
			  AND u.deleted_at IS NULL
			RETURNING u.*
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				u.id::text AS id,
				u.company_id::text AS "companyId",
				u.full_name AS "fullName",
				u.email,
				COALESCE(u.phone, '') AS phone,
				COALESCE(u.job_title, '') AS "jobTitle",
				p.key AS "accessProfileKey",
				p.name AS "accessProfileName",
				u.status,
				u.created_at AS "createdAt",
				u.updated_at AS "updatedAt"
			FROM updated_user u
			JOIN access_profiles p ON p.id = u.access_profile_id
		) item;
	`,
		sqlQuote(req.AccessProfileKey),
		sqlQuote(req.FullName),
		sqlQuote(req.Email),
		sqlQuote(req.Phone),
		sqlQuote(req.JobTitle),
		sqlQuote(req.Status),
		sqlQuote(req.Status),
		sqlQuote(req.Status),
		sqlQuote(userID),
	)
}

func buildCreateNewsSQL(req createNewsRequest, status string) string {
	categoryFilter := fmt.Sprintf("slug = %s", sqlQuote(req.CategorySlug))
	if req.CategoryID != "" {
		categoryFilter = fmt.Sprintf("id = %s::uuid", sqlQuote(req.CategoryID))
	}

	imageURL := nullOrQuote(req.MainImageURL)
	imageFileName := req.MainImageFileName
	if imageFileName == "" {
		imageFileName = "imagem-noticia"
	}

	return fmt.Sprintf(`
		WITH selected_category AS (
			SELECT id FROM news_categories WHERE %s AND is_active = true LIMIT 1
		),
		inserted_media AS (
			INSERT INTO media_files (media_type, file_name, file_url, mime_type, source)
			SELECT 'image', %s, %s, %s, 'external_link'
			WHERE %s IS NOT NULL
			RETURNING id, file_url
		),
		inserted_news AS (
			INSERT INTO news (
				category_id,
				title,
				status,
				summary,
				content,
				main_image_media_id,
				published_at
			)
			SELECT
				selected_category.id,
				%s,
				%s,
				NULLIF(%s, ''),
				NULLIF(%s, ''),
				(SELECT id FROM inserted_media),
				CASE WHEN %s = 'draft' THEN NULL ELSE now() END
			FROM selected_category
			RETURNING *
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				n.id::text AS id,
				n.title,
				c.id::text AS "categoryId",
				c.name AS "categoryName",
				c.slug AS "categorySlug",
				n.status,
				COALESCE(n.summary, '') AS summary,
				COALESCE(n.content, '') AS content,
				COALESCE((SELECT file_url FROM inserted_media), '') AS "mainImageUrl",
				n.published_at AS "publishedAt",
				n.created_at AS "createdAt",
				n.updated_at AS "updatedAt"
			FROM inserted_news n
			JOIN news_categories c ON c.id = n.category_id
		) item;
	`, categoryFilter,
		sqlQuote(imageFileName),
		imageURL,
		nullOrQuote(req.MainImageMimeType),
		imageURL,
		sqlQuote(req.Title),
		sqlQuote(status),
		sqlQuote(req.Summary),
		sqlQuote(req.Content),
		sqlQuote(status),
	)
}

func saveNewsImage(req createNewsRequest) (string, error) {
	header, payload, ok := strings.Cut(req.MainImageDataURL, ",")
	if !ok || !strings.HasPrefix(header, "data:image/") {
		return "", errors.New("imagem invalida")
	}

	mimeType := req.MainImageMimeType
	if mimeType == "" {
		mimeType = strings.TrimPrefix(strings.TrimSuffix(header, ";base64"), "data:")
	}

	ext := imageExtension(mimeType)
	if ext == "" {
		return "", errors.New("tipo de imagem nao permitido")
	}

	bytes, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", errors.New("nao foi possivel ler a imagem")
	}

	if len(bytes) > 5*1024*1024 {
		return "", errors.New("imagem maior que 5MB")
	}

	dir := filepath.Join("uploads", "news")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", errors.New("nao foi possivel preparar a pasta de imagens")
	}

	fileName := fmt.Sprintf("news-%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(dir, fileName)
	if err := os.WriteFile(path, bytes, 0644); err != nil {
		return "", errors.New("nao foi possivel salvar a imagem")
	}

	return strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/") + "/uploads/news/" + fileName, nil
}

func imageExtension(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func normalizeNewsStatus(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "draft", "rascunho":
		return "draft", nil
	case "published", "publicado", "publicada":
		return "published", nil
	case "featured", "destaque", "destaque principal":
		return "featured", nil
	default:
		return "", errors.New("status da noticia invalido")
	}
}

func splitResourcePath(path string, prefix string) (string, string) {
	remainder := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if remainder == "" {
		return "", ""
	}
	parts := strings.Split(remainder, "/")
	id := strings.TrimSpace(parts[0])
	action := ""
	if len(parts) > 1 {
		action = strings.TrimSpace(parts[1])
	}
	return id, action
}

func randomToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func normalizeCNPJ(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func normalizeState(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) > 2 {
		return value[:2]
	}
	return value
}

func normalizeAccessProfileKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "/", "_")

	switch value {
	case "", "commercial", "comercial", "comercial_relacionamento", "comercial___relacionamento":
		return "commercial"
	case "company_admin", "administrador_da_empresa", "administrador":
		return "company_admin"
	case "technical", "tecnico", "técnico":
		return "technical"
	case "reader", "leitor":
		return "reader"
	default:
		return value
	}
}

func normalizeUserStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "active", "ativo":
		return "active"
	case "blocked", "bloqueado":
		return "blocked"
	case "inactive", "inativo":
		return "inactive"
	case "removed", "removido", "remove":
		return "removed"
	case "pending_invite", "convite_pendente", "pendente":
		return "pending_invite"
	default:
		return ""
	}
}

func humanizeConstraintError(message string) string {
	switch {
	case strings.Contains(message, "company_invitations_trade_name_uk"), strings.Contains(message, "companies_trade_name_uk"):
		return "nome da empresa ja cadastrado"
	case strings.Contains(message, "company_invitations_cnpj_uk"), strings.Contains(message, "companies_cnpj_uk"):
		return "cnpj ja cadastrado"
	case strings.Contains(message, "invalid input syntax for type uuid"):
		return "identificador invalido"
	default:
		return message
	}
}

func nullOrQuote(value string) string {
	if strings.TrimSpace(value) == "" {
		return "NULL"
	}
	return sqlQuote(value)
}

func sqlQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Private-Network", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("could not write json response: %v", err)
	}
}

func writeRawJSON(w http.ResponseWriter, status int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(payload); err != nil {
		log.Printf("could not write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
