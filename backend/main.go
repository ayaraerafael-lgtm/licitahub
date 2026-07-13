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
	"sync"
	"time"
)

type app struct {
	psqlPath string
	pg       postgresConfig
	chatHub  *chatHub
}

type chatHub struct {
	mu      sync.Mutex
	clients map[string]map[chan []byte]bool
}

func newChatHub() *chatHub {
	return &chatHub{clients: make(map[string]map[chan []byte]bool)}
}

func (h *chatHub) add(companyID string) chan []byte {
	ch := make(chan []byte, 16)
	if companyID == "" {
		return ch
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[companyID] == nil {
		h.clients[companyID] = make(map[chan []byte]bool)
	}
	h.clients[companyID][ch] = true
	return ch
}

func (h *chatHub) remove(companyID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if companyClients := h.clients[companyID]; companyClients != nil {
		delete(companyClients, ch)
		if len(companyClients) == 0 {
			delete(h.clients, companyID)
		}
	}
	close(ch)
}

func (h *chatHub) broadcast(companyID string, payload []byte) {
	if companyID == "" || len(payload) == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients[companyID] {
		select {
		case ch <- payload:
		default:
		}
	}
}

type postgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type sessionUser struct {
	UserID    string
	CompanyID string
	RoleKey   string
}

func (s sessionUser) canManageCompany() bool {
	return s.RoleKey == "company_admin" && s.CompanyID != ""
}

func (s sessionUser) canManageCompanyUsers() bool {
	return s.RoleKey == "company_admin" && s.CompanyID != ""
}

func (s sessionUser) canManagePlatform() bool {
	return s.RoleKey == "platform_admin"
}

func (s sessionUser) canUseChat() bool {
	return s.CompanyID != "" && (s.RoleKey == "company_admin" || s.RoleKey == "commercial")
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
	ExpiresAt         string `json:"expiresAt"`
	MainImageURL      string `json:"mainImageUrl"`
	MainImageDataURL  string `json:"mainImageDataUrl"`
	MainImageFileName string `json:"mainImageFileName"`
	MainImageMimeType string `json:"mainImageMimeType"`
}

type updateNewsStatusRequest struct {
	Status    string `json:"status"`
	ExpiresAt string `json:"expiresAt"`
}

type createPostRequest struct {
	Title             string `json:"title"`
	CategorySlug      string `json:"categorySlug"`
	Visibility        string `json:"visibility"`
	Content           string `json:"content"`
	MainImageURL      string `json:"mainImageUrl"`
	MainImageDataURL  string `json:"mainImageDataUrl"`
	MainImageFileName string `json:"mainImageFileName"`
	MainImageMimeType string `json:"mainImageMimeType"`
}

type updatePostRequest struct {
	Title             string `json:"title"`
	CategorySlug      string `json:"categorySlug"`
	Visibility        string `json:"visibility"`
	Content           string `json:"content"`
	MainImageURL      string `json:"mainImageUrl"`
	MainImageDataURL  string `json:"mainImageDataUrl"`
	MainImageFileName string `json:"mainImageFileName"`
	MainImageMimeType string `json:"mainImageMimeType"`
}

type createPostCommentRequest struct {
	Content string `json:"content"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type createTenderRequest struct {
	Agency            string `json:"agency"`
	Number            string `json:"number"`
	Object            string `json:"object"`
	Modality          string `json:"modality"`
	JudgmentCriterion string `json:"judgmentCriterion"`
	EstimatedValue    string `json:"estimatedValue"`
	State             string `json:"state"`
	City              string `json:"city"`
	OpeningDate       string `json:"openingDate"`
	Status            string `json:"status"`
	CloudFolderURL    string `json:"cloudFolderUrl"`
	AnalysisDataURL   string `json:"analysisDataUrl"`
	AnalysisFileName  string `json:"analysisFileName"`
	AnalysisMimeType  string `json:"analysisMimeType"`
}

type updateTenderAnalysisRequest struct {
	AnalysisDataURL  string `json:"analysisDataUrl"`
	AnalysisFileName string `json:"analysisFileName"`
	AnalysisMimeType string `json:"analysisMimeType"`
}

type tenderInterestRequirementRequest struct {
	RequirementKey string `json:"requirementKey"`
	StatusKey      string `json:"statusKey"`
	WhatWeHave     string `json:"whatWeHave"`
	WhatWeSeek     string `json:"whatWeSeek"`
}

type createTenderInterestRequest struct {
	GeneralPosition string                             `json:"generalPosition"`
	DesiredRole     string                             `json:"desiredRole"`
	PublicSummary   string                             `json:"publicSummary"`
	InternalNote    string                             `json:"internalNote"`
	Requirements    []tenderInterestRequirementRequest `json:"requirements"`
}

type partnerEvaluationRequest struct {
	Decision string `json:"decision"`
}

type updatePartnershipAdRequest struct {
	OfferSummary string `json:"offerSummary"`
	SeekSummary  string `json:"seekSummary"`
}

type updateConsortiumLeaderRequest struct {
	LeadCompanyID string `json:"leadCompanyId"`
	Notes         string `json:"notes"`
}

type createConsortiumAdRequest struct {
	NeedSummary string `json:"needSummary"`
	Notes       string `json:"notes"`
}

type acceptConsortiumApplicationRequest struct {
	ApplicationID string `json:"applicationId"`
}

type withdrawFromConsortiumRequest struct {
	SuccessorCompanyID string `json:"successorCompanyId"`
}

type startChatRequest struct {
	AdID string `json:"adId"`
}

type createChatMessageRequest struct {
	Content string `json:"content"`
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
	ProfilePhotoURL          string `json:"profilePhotoUrl"`
	ProfilePhotoDataURL      string `json:"profilePhotoDataUrl"`
	ProfilePhotoFileName     string `json:"profilePhotoFileName"`
	ProfilePhotoMimeType     string `json:"profilePhotoMimeType"`
}

type reviewCompanyInvitationRequest struct {
	Decision          string `json:"decision"`
	AdjustmentRequest string `json:"adjustmentRequest"`
	ReviewNote        string `json:"reviewNote"`
}

type createCompanyUserRequest struct {
	CompanyID            string `json:"companyId"`
	FullName             string `json:"fullName"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	JobTitle             string `json:"jobTitle"`
	AccessProfileKey     string `json:"accessProfileKey"`
	ProfilePhotoDataURL  string `json:"profilePhotoDataUrl"`
	ProfilePhotoFileName string `json:"profilePhotoFileName"`
	ProfilePhotoMimeType string `json:"profilePhotoMimeType"`
	ProfilePhotoURL      string `json:"profilePhotoUrl"`
}

type updateCompanyUserRequest struct {
	FullName             string `json:"fullName"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	JobTitle             string `json:"jobTitle"`
	AccessProfileKey     string `json:"accessProfileKey"`
	Status               string `json:"status"`
	ProfilePhotoDataURL  string `json:"profilePhotoDataUrl"`
	ProfilePhotoFileName string `json:"profilePhotoFileName"`
	ProfilePhotoMimeType string `json:"profilePhotoMimeType"`
	ProfilePhotoURL      string `json:"profilePhotoUrl"`
}

type updateCompanyProfileRequest struct {
	Website                  string `json:"website"`
	CompanySize              string `json:"companySize"`
	InstitutionalDescription string `json:"institutionalDescription"`
	State                    string `json:"state"`
	City                     string `json:"city"`
	NationalCoverage         bool   `json:"nationalCoverage"`
	LogoDataURL              string `json:"logoDataUrl"`
	LogoFileName             string `json:"logoFileName"`
	LogoMimeType             string `json:"logoMimeType"`
}

type updateMyUserProfileRequest struct {
	FullName             string `json:"fullName"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	ProfilePhotoDataURL  string `json:"profilePhotoDataUrl"`
	ProfilePhotoFileName string `json:"profilePhotoFileName"`
	ProfilePhotoMimeType string `json:"profilePhotoMimeType"`
	ProfilePhotoURL      string `json:"profilePhotoUrl"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	application := &app{
		psqlPath: getenv("PSQL_PATH", `C:\Program Files\PostgreSQL\17\bin\psql.exe`),
		chatHub:  newChatHub(),
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
	mux.HandleFunc("/api/auth/login", application.handleLogin)
	mux.HandleFunc("/api/auth/session", application.handleAuthSession)
	mux.HandleFunc("/api/auth/logout", application.handleLogout)
	mux.HandleFunc("/api/auth/forgot-password", application.handleForgotPassword)
	mux.HandleFunc("/api/auth/reset-password", application.handleResetPassword)
	mux.HandleFunc("/api/users/me", application.handleMyUserProfile)
	mux.HandleFunc("/api/notifications/read-all", application.handleNotificationsReadAll)
	mux.HandleFunc("/api/notifications", application.handleNotifications)
	mux.HandleFunc("/api/access-profiles", application.handleAccessProfiles)
	mux.HandleFunc("/api/company-invitations", application.handleCompanyInvitations)
	mux.HandleFunc("/api/company-invitations/", application.handleCompanyInvitationByPath)
	mux.HandleFunc("/api/companies/", application.handleCompanyByPath)
	mux.HandleFunc("/api/companies", application.handleCompanies)
	mux.HandleFunc("/api/company-users", application.handleCompanyUsers)
	mux.HandleFunc("/api/company-users/", application.handleCompanyUserByPath)
	mux.HandleFunc("/api/news/categories", application.handleNewsCategories)
	mux.HandleFunc("/api/news/admin", application.handleAdminNews)
	mux.HandleFunc("/api/news/", application.handleNewsByPath)
	mux.HandleFunc("/api/news", application.handleNews)
	mux.HandleFunc("/api/community/categories", application.handlePostCategories)
	mux.HandleFunc("/api/community/posts/", application.handleCommunityPostByPath)
	mux.HandleFunc("/api/community/posts", application.handleCommunityPosts)
	mux.HandleFunc("/api/tenders/", application.handleTenderByPath)
	mux.HandleFunc("/api/tenders", application.handleTenders)
	mux.HandleFunc("/api/partnership-ads/", application.handlePartnershipAdByPath)
	mux.HandleFunc("/api/partnership-ads", application.handlePartnershipAds)
	mux.HandleFunc("/api/chats/stream", application.handleChatStream)
	mux.HandleFunc("/api/chats/", application.handleChatByPath)
	mux.HandleFunc("/api/chats", application.handleChats)
	mux.HandleFunc("/api/matches/", application.handleMatchByPath)
	mux.HandleFunc("/api/matches", application.handleMatches)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	mux.Handle("/", frontendFileServer())

	port := getenv("APP_PORT", "8080")
	log.Printf("LicitaHub API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, withCORS(mux)))
}

func frontendFileServer() http.Handler {
	distDir := filepath.Clean(filepath.Join("..", "dist"))
	indexPath := filepath.Join(distDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return http.FileServer(http.Dir(".."))
	}

	fileServer := http.FileServer(http.Dir(distDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/preview-react-cdn.html" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		relativePath := filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/"))
		requestedPath := filepath.Join(distDir, relativePath)
		if info, err := os.Stat(requestedPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, indexPath)
	})
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

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "informe email e senha")
		return
	}

	token, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel iniciar a sessao")
		return
	}
	tokenHash := hashSessionToken(token)

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_user AS (
			SELECT
				u.id,
				u.company_id,
				u.full_name,
				u.email,
				u.status AS user_status,
				u.password_hash,
				ap.key AS role_key,
				ap.name AS role_name,
				COALESCE(c.trade_name, 'LicitaHub') AS company_name,
				COALESCE(c.status, 'active') AS company_status,
				COALESCE(photo.file_url, '') AS profile_photo_url,
				COALESCE(company_logo.file_url, '') AS company_logo_url
			FROM users u
			JOIN access_profiles ap ON ap.id = u.access_profile_id
			LEFT JOIN companies c ON c.id = u.company_id
			LEFT JOIN media_files photo ON photo.id = u.profile_photo_media_id
			LEFT JOIN company_profiles cp ON cp.company_id = c.id
			LEFT JOIN media_files company_logo ON company_logo.id = cp.logo_media_id
			WHERE lower(u.email) = %s
			  AND u.deleted_at IS NULL
			ORDER BY u.created_at
			LIMIT 1
		),
		valid_user AS (
			SELECT *
			FROM selected_user
			WHERE password_hash = %s
			  AND user_status = 'active'
			  AND (role_key = 'platform_admin' OR company_status = 'active')
		),
		created_session AS (
			INSERT INTO auth_sessions (user_id, token_hash, expires_at)
			SELECT id, %s, now() + interval '12 hours'
			FROM valid_user
			RETURNING id, user_id, expires_at
		),
		updated_user AS (
			UPDATE users u
			SET last_login_at = now()
			FROM valid_user v
			WHERE u.id = v.id
			RETURNING u.id
		)
		SELECT json_build_object(
			'authenticated', EXISTS (SELECT 1 FROM created_session),
			'reason', CASE
				WHEN NOT EXISTS (SELECT 1 FROM selected_user) THEN 'invalid_credentials'
				WHEN NOT EXISTS (SELECT 1 FROM selected_user WHERE password_hash = %s) THEN 'invalid_credentials'
				WHEN EXISTS (SELECT 1 FROM selected_user WHERE user_status = 'pending_invite' OR company_status = 'pending_review') THEN 'pending_approval'
				WHEN EXISTS (SELECT 1 FROM selected_user WHERE user_status = 'blocked') THEN 'user_blocked'
				WHEN EXISTS (SELECT 1 FROM selected_user WHERE company_status IN ('blocked', 'rejected', 'inactive')) THEN 'company_blocked'
				ELSE 'access_denied'
			END,
			'user', (
				SELECT json_build_object(
					'id', id::text,
					'fullName', full_name,
					'email', email,
					'roleKey', role_key,
					'roleName', role_name,
					'companyId', company_id::text,
					'companyName', company_name,
					'profilePhotoUrl', profile_photo_url,
					'companyLogoUrl', company_logo_url
				)
				FROM valid_user
			)
		);
	`,
		sqlQuote(req.Email),
		sqlQuote(hashPassword(req.Password)),
		sqlQuote(tokenHash),
		sqlQuote(hashPassword(req.Password)),
	))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel entrar no sistema")
		return
	}

	var result struct {
		Authenticated bool   `json:"authenticated"`
		Reason        string `json:"reason"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		writeError(w, http.StatusInternalServerError, "resposta de login invalida")
		return
	}
	if !result.Authenticated {
		message := "email ou senha invalidos"
		status := http.StatusUnauthorized
		switch result.Reason {
		case "pending_approval":
			message = "cadastro aguardando aprovacao da LicitaHub"
			status = http.StatusForbidden
		case "user_blocked":
			message = "usuario bloqueado"
			status = http.StatusForbidden
		case "company_blocked":
			message = "acesso da empresa nao esta liberado"
			status = http.StatusForbidden
		}
		writeError(w, status, message)
		return
	}

	setSessionCookie(w, token, time.Now().Add(12*time.Hour))
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleAuthSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	cookie, err := r.Cookie("licitahub_session")
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		writeError(w, http.StatusUnauthorized, "sessao nao encontrada")
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				u.id::text AS id,
				u.full_name AS "fullName",
				u.email,
				ap.key AS "roleKey",
				ap.name AS "roleName",
				u.company_id::text AS "companyId",
				COALESCE(c.trade_name, 'LicitaHub') AS "companyName",
				COALESCE(photo.file_url, '') AS "profilePhotoUrl",
				COALESCE(company_logo.file_url, '') AS "companyLogoUrl"
			FROM auth_sessions s
			JOIN users u ON u.id = s.user_id
			JOIN access_profiles ap ON ap.id = u.access_profile_id
			LEFT JOIN companies c ON c.id = u.company_id
			LEFT JOIN media_files photo ON photo.id = u.profile_photo_media_id
			LEFT JOIN company_profiles cp ON cp.company_id = c.id
			LEFT JOIN media_files company_logo ON company_logo.id = cp.logo_media_id
			WHERE s.token_hash = %s
			  AND s.revoked_at IS NULL
			  AND s.expires_at > now()
			  AND u.status = 'active'
			  AND u.deleted_at IS NULL
			  AND (ap.key = 'platform_admin' OR c.status = 'active')
			LIMIT 1
		) item;
	`, sqlQuote(hashSessionToken(cookie.Value))))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel validar a sessao")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		clearSessionCookie(w)
		writeError(w, http.StatusUnauthorized, "sessao expirada")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if cookie, err := r.Cookie("licitahub_session"); err == nil && cookie.Value != "" {
		_, _ = a.runPSQL(r.Context(), fmt.Sprintf(`
			UPDATE auth_sessions SET revoked_at = now()
			WHERE token_hash = %s AND revoked_at IS NULL;
		`, sqlQuote(hashSessionToken(cookie.Value))))
	}
	clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"loggedOut": true})
}

func (a *app) handleMyUserProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getMyUserProfile(w, r)
	case http.MethodPut:
		a.updateMyUserProfile(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) getMyUserProfile(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(myUserProfileSQL(), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar seu perfil")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) updateMyUserProfile(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}

	var req updateMyUserProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.ProfilePhotoDataURL = strings.TrimSpace(req.ProfilePhotoDataURL)
	req.ProfilePhotoFileName = strings.TrimSpace(req.ProfilePhotoFileName)
	req.ProfilePhotoMimeType = strings.TrimSpace(req.ProfilePhotoMimeType)
	if req.FullName == "" {
		writeError(w, http.StatusBadRequest, "nome e obrigatorio")
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
	if req.ProfilePhotoDataURL != "" {
		photoURL, err := saveProfilePhoto(req.ProfilePhotoDataURL, req.ProfilePhotoMimeType)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.ProfilePhotoURL = photoURL
	}

	payload, err := a.queryJSON(r.Context(), buildUpdateMyUserProfileSQL(session, req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel salvar seu perfil: "+humanizeConstraintError(err.Error()))
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) currentSessionUser(w http.ResponseWriter, r *http.Request) (sessionUser, bool) {
	cookie, err := r.Cookie("licitahub_session")
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		writeError(w, http.StatusUnauthorized, "sessao nao encontrada")
		return sessionUser{}, false
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				u.id::text AS "userId",
				COALESCE(u.company_id::text, '') AS "companyId",
				ap.key AS "roleKey"
			FROM auth_sessions s
			JOIN users u ON u.id = s.user_id
			JOIN access_profiles ap ON ap.id = u.access_profile_id
			LEFT JOIN companies c ON c.id = u.company_id
			WHERE s.token_hash = %s
			  AND s.revoked_at IS NULL
			  AND s.expires_at > now()
			  AND u.status = 'active'
			  AND u.deleted_at IS NULL
			  AND (ap.key = 'platform_admin' OR c.status = 'active')
			LIMIT 1
		) item;
	`, sqlQuote(hashSessionToken(cookie.Value))))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel validar a sessao")
		return sessionUser{}, false
	}
	if strings.TrimSpace(string(payload)) == "null" {
		clearSessionCookie(w)
		writeError(w, http.StatusUnauthorized, "sessao expirada")
		return sessionUser{}, false
	}

	var session sessionUser
	if err := json.Unmarshal(payload, &session); err != nil {
		writeError(w, http.StatusInternalServerError, "sessao invalida")
		return sessionUser{}, false
	}
	return session, true
}

func (a *app) handleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				n.id::text AS id,
				n.type,
				n.title,
				COALESCE(n.message, '') AS message,
				COALESCE(n.destination_screen, '') AS "destinationScreen",
				COALESCE(n.related_entity_type, '') AS "relatedEntityType",
				COALESCE(n.related_entity_id::text, '') AS "relatedEntityId",
				n.created_at AS "createdAt"
			FROM notifications n
			WHERE n.is_read = false
			  AND (
				n.recipient_user_id = %s::uuid
				OR (
					NULLIF(%s, '')::uuid IS NOT NULL
					AND n.recipient_company_id = NULLIF(%s, '')::uuid
				)
			  )
			ORDER BY n.created_at DESC
			LIMIT 30
		) item;
	`, sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar notificacoes")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleNotificationsReadAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH read_notifications AS (
			UPDATE notifications n
			SET is_read = true, read_at = now()
			WHERE n.is_read = false
			  AND (
				n.recipient_user_id = %s::uuid
				OR (
					NULLIF(%s, '')::uuid IS NOT NULL
					AND n.recipient_company_id = NULLIF(%s, '')::uuid
				)
			  )
			RETURNING id
		)
		SELECT json_build_object('readCount', (SELECT count(*) FROM read_notifications));
	`, sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel zerar notificacoes")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "informe o email")
		return
	}
	token, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel gerar o link")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_user AS (
			SELECT id FROM users
			WHERE lower(email) = %s AND deleted_at IS NULL
			ORDER BY created_at LIMIT 1
		),
		invalidated AS (
			UPDATE password_reset_tokens p
			SET used_at = now()
			FROM selected_user u
			WHERE p.user_id = u.id AND p.used_at IS NULL
			RETURNING p.id
		),
		created_token AS (
			INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
			SELECT id, %s, now() + interval '1 hour'
			FROM selected_user
			RETURNING id
		)
		SELECT json_build_object('created', EXISTS (SELECT 1 FROM created_token));
	`, sqlQuote(req.Email), sqlQuote(hashSessionToken(token))))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel gerar o link")
		return
	}
	var result struct {
		Created bool `json:"created"`
	}
	_ = json.Unmarshal(payload, &result)
	response := map[string]any{"message": "Se o e-mail estiver cadastrado, a recuperação foi gerada."}
	if result.Created {
		baseURL := strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/")
		response["resetUrl"] = baseURL + "/#reset-password?token=" + token
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *app) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	req.Password = strings.TrimSpace(req.Password)
	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "link de recuperacao invalido")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "senha deve ter pelo menos 8 caracteres")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_token AS (
			SELECT id, user_id
			FROM password_reset_tokens
			WHERE token_hash = %s
			  AND used_at IS NULL
			  AND expires_at > now()
			LIMIT 1
		),
		updated_user AS (
			UPDATE users u
			SET
				password_hash = %s,
				status = CASE WHEN u.status = 'pending_invite' THEN 'active' ELSE u.status END,
				updated_at = now()
			FROM selected_token t
			WHERE u.id = t.user_id
			RETURNING u.id
		),
		used_token AS (
			UPDATE password_reset_tokens p
			SET used_at = now()
			FROM selected_token t
			WHERE p.id = t.id
			RETURNING p.id
		),
		revoked_sessions AS (
			UPDATE auth_sessions s
			SET revoked_at = now()
			FROM selected_token t
			WHERE s.user_id = t.user_id AND s.revoked_at IS NULL
			RETURNING s.id
		)
		SELECT json_build_object('reset', EXISTS (SELECT 1 FROM updated_user));
	`, sqlQuote(hashSessionToken(req.Token)), sqlQuote(hashPassword(req.Password))))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel redefinir a senha")
		return
	}
	var result struct {
		Reset bool `json:"reset"`
	}
	_ = json.Unmarshal(payload, &result)
	if !result.Reset {
		writeError(w, http.StatusBadRequest, "link expirado ou ja utilizado")
		return
	}
	clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"reset": true})
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
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode gerenciar convites")
		return
	}

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
	if action == "resubmit" && r.Method == http.MethodPatch {
		a.resubmitCompanyInvitation(w, r, id)
		return
	}

	if action == "cancel" && r.Method == http.MethodPatch {
		session, ok := a.currentSessionUser(w, r)
		if !ok {
			return
		}
		if !session.canManagePlatform() {
			writeError(w, http.StatusForbidden, "somente administrador da plataforma pode cancelar convites")
			return
		}
		a.changeCompanyInvitationStatus(w, r, id, "cancelled")
		return
	}

	if action == "review" && r.Method == http.MethodPatch {
		session, ok := a.currentSessionUser(w, r)
		if !ok {
			return
		}
		if !session.canManagePlatform() {
			writeError(w, http.StatusForbidden, "somente administrador da plataforma pode analisar empresas")
			return
		}
		a.reviewCompanyInvitation(w, r, id)
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
				i.updated_at AS "updatedAt",
				COALESCE(c.city, '') AS city,
				COALESCE(p.website, '') AS website,
				COALESCE(p.institutional_description, '') AS "institutionalDescription",
				COALESCE(admin_user.full_name, '') AS "adminFullName",
				COALESCE(admin_user.email, '') AS "adminEmail",
				COALESCE(admin_user.phone, '') AS "adminPhone",
				COALESCE(admin_user.job_title, '') AS "adminJobTitle",
				COALESCE(admin_user.status, '') AS "adminUserStatus",
				COALESCE(latest_review.status, '') AS "reviewStatus",
				COALESCE(latest_review.adjustment_request, '') AS "adjustmentRequest"
			FROM company_invitations i
			LEFT JOIN companies c ON c.id = i.company_id
			LEFT JOIN company_profiles p ON p.company_id = c.id
			LEFT JOIN LATERAL (
				SELECT u.full_name, u.email, u.phone, u.job_title, u.status
				FROM users u
				JOIN access_profiles ap ON ap.id = u.access_profile_id
				WHERE u.company_id = c.id
				  AND ap.key = 'company_admin'
				  AND u.deleted_at IS NULL
				ORDER BY u.created_at
				LIMIT 1
			) admin_user ON true
			LEFT JOIN LATERAL (
				SELECT r.status, r.adjustment_request
				FROM company_reviews r
				WHERE r.company_id = c.id
				ORDER BY r.created_at DESC
				LIMIT 1
			) latest_review ON true
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

func (a *app) reviewCompanyInvitation(w http.ResponseWriter, r *http.Request, invitationID string) {
	var req reviewCompanyInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.Decision = strings.ToLower(strings.TrimSpace(req.Decision))
	req.AdjustmentRequest = strings.TrimSpace(req.AdjustmentRequest)
	req.ReviewNote = strings.TrimSpace(req.ReviewNote)
	if req.Decision != "approved" && req.Decision != "adjustment_requested" && req.Decision != "rejected" {
		writeError(w, http.StatusBadRequest, "decisao de analise invalida")
		return
	}
	if req.Decision == "adjustment_requested" && req.AdjustmentRequest == "" {
		writeError(w, http.StatusBadRequest, "informe o ajuste solicitado")
		return
	}

	companyStatus := "pending_review"
	invitationStatus := "pending_review"
	userStatus := "pending_invite"
	notificationTitle := "Cadastro em análise"
	notificationMessage := "A LicitaHub solicitou ajustes no cadastro da empresa."
	if req.Decision == "approved" {
		companyStatus = "active"
		invitationStatus = "accepted"
		userStatus = "active"
		notificationTitle = "Empresa aprovada"
		notificationMessage = "O cadastro foi aprovado e o acesso à LicitaHub está liberado."
	} else if req.Decision == "rejected" {
		companyStatus = "rejected"
		invitationStatus = "rejected"
		userStatus = "inactive"
		notificationTitle = "Cadastro não aprovado"
		notificationMessage = "O cadastro da empresa não foi aprovado pela LicitaHub."
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_invitation AS (
			SELECT id, company_id
			FROM company_invitations
			WHERE id = %s::uuid
			  AND status = 'pending_review'
			  AND company_id IS NOT NULL
			LIMIT 1
		),
		inserted_review AS (
			INSERT INTO company_reviews (
				company_id, status, adjustment_request, review_note
			)
			SELECT
				company_id, %s, NULLIF(%s, ''), NULLIF(%s, '')
			FROM selected_invitation
			RETURNING id, company_id, status
		),
		updated_company AS (
			UPDATE companies c
			SET status = %s
			FROM selected_invitation si
			WHERE c.id = si.company_id
			RETURNING c.id, c.trade_name, c.status
		),
		updated_users AS (
			UPDATE users u
			SET status = %s
			FROM selected_invitation si
			WHERE u.company_id = si.company_id
			  AND u.deleted_at IS NULL
			RETURNING u.id, u.company_id, u.status
		),
		updated_invitation AS (
			UPDATE company_invitations i
			SET status = %s
			FROM selected_invitation si
			WHERE i.id = si.id
			RETURNING i.id, i.company_id, i.status
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id, u.company_id, 'system', %s, %s,
				'company-dashboard', 'company', u.company_id
			FROM updated_users u
			RETURNING id
		)
		SELECT json_build_object(
			'invitationId', i.id::text,
			'invitationStatus', i.status,
			'companyId', c.id::text,
			'companyName', c.trade_name,
			'companyStatus', c.status,
			'decision', r.status
		)
		FROM updated_invitation i
		JOIN updated_company c ON c.id = i.company_id
		JOIN inserted_review r ON r.company_id = c.id;
	`,
		sqlQuote(invitationID),
		sqlQuote(req.Decision),
		sqlQuote(req.AdjustmentRequest),
		sqlQuote(req.ReviewNote),
		sqlQuote(companyStatus),
		sqlQuote(userStatus),
		sqlQuote(invitationStatus),
		sqlQuote(notificationTitle),
		sqlQuote(notificationMessage),
	))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel concluir a analise da empresa")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "convite aguardando analise nao encontrado")
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

func (a *app) resubmitCompanyInvitation(w http.ResponseWriter, r *http.Request, invitationID string) {
	var req acceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	req.Website = strings.TrimSpace(req.Website)
	req.InstitutionalDescription = strings.TrimSpace(req.InstitutionalDescription)
	req.City = strings.TrimSpace(req.City)
	req.State = normalizeState(req.State)
	req.AdminFullName = strings.TrimSpace(req.AdminFullName)
	req.AdminEmail = strings.ToLower(strings.TrimSpace(req.AdminEmail))
	req.AdminPhone = strings.TrimSpace(req.AdminPhone)
	req.AdminJobTitle = strings.TrimSpace(req.AdminJobTitle)
	if req.Token == "" || req.AdminFullName == "" || req.AdminEmail == "" || req.AdminPhone == "" {
		writeError(w, http.StatusBadRequest, "preencha os dados obrigatorios")
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_invitation AS (
			SELECT i.id, i.company_id
			FROM company_invitations i
			WHERE i.id = %s::uuid
			  AND i.invitation_token = %s
			  AND i.status = 'pending_review'
			  AND EXISTS (
				SELECT 1 FROM company_reviews r
				WHERE r.company_id = i.company_id
				  AND r.status = 'adjustment_requested'
				  AND r.created_at = (SELECT max(r2.created_at) FROM company_reviews r2 WHERE r2.company_id = i.company_id)
			  )
		),
		updated_company AS (
			UPDATE companies c SET
				main_contact_name = %s,
				main_contact_email = %s,
				main_contact_phone = %s,
				state = NULLIF(%s, ''),
				city = NULLIF(%s, ''),
				updated_at = now()
			FROM selected_invitation si
			WHERE c.id = si.company_id
			RETURNING c.id, c.trade_name
		),
		updated_profile AS (
			UPDATE company_profiles p SET
				website = NULLIF(%s, ''),
				institutional_description = NULLIF(%s, ''),
				state = NULLIF(%s, ''),
				city = NULLIF(%s, ''),
				updated_at = now()
			FROM selected_invitation si
			WHERE p.company_id = si.company_id
			RETURNING p.company_id
		),
		updated_user AS (
			UPDATE users u SET
				full_name = %s,
				email = %s,
				phone = %s,
				job_title = NULLIF(%s, ''),
				updated_at = now()
			FROM selected_invitation si, access_profiles ap
			WHERE u.company_id = si.company_id
			  AND u.access_profile_id = ap.id
			  AND ap.key = 'company_admin'
			RETURNING u.id
		),
		created_review AS (
			INSERT INTO company_reviews (company_id, status, review_note)
			SELECT company_id, 'resubmitted', 'Cadastro corrigido e reenviado para nova análise.'
			FROM selected_invitation
			RETURNING id
		)
		SELECT json_build_object(
			'resubmitted', EXISTS (SELECT 1 FROM updated_company),
			'companyName', (SELECT trade_name FROM updated_company LIMIT 1)
		);
	`,
		sqlQuote(invitationID), sqlQuote(req.Token),
		sqlQuote(req.AdminFullName), sqlQuote(req.AdminEmail), sqlQuote(req.AdminPhone),
		sqlQuote(req.State), sqlQuote(req.City),
		sqlQuote(req.Website), sqlQuote(req.InstitutionalDescription), sqlQuote(req.State), sqlQuote(req.City),
		sqlQuote(req.AdminFullName), sqlQuote(req.AdminEmail), sqlQuote(req.AdminPhone), sqlQuote(req.AdminJobTitle),
	))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel reenviar o cadastro")
		return
	}
	var result struct {
		Resubmitted bool `json:"resubmitted"`
	}
	_ = json.Unmarshal(payload, &result)
	if !result.Resubmitted {
		writeError(w, http.StatusBadRequest, "nao ha ajuste pendente para este convite")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
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
	req.ProfilePhotoURL = strings.TrimSpace(req.ProfilePhotoURL)
	req.ProfilePhotoDataURL = strings.TrimSpace(req.ProfilePhotoDataURL)
	req.ProfilePhotoFileName = strings.TrimSpace(req.ProfilePhotoFileName)
	req.ProfilePhotoMimeType = strings.TrimSpace(req.ProfilePhotoMimeType)

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
	if req.ProfilePhotoDataURL != "" {
		photoURL, err := saveProfilePhoto(req.ProfilePhotoDataURL, req.ProfilePhotoMimeType)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.ProfilePhotoURL = photoURL
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
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode listar empresas")
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

func (a *app) handleCompanyByPath(w http.ResponseWriter, r *http.Request) {
	id, action := splitResourcePath(r.URL.Path, "/api/companies/")
	if id != "me" {
		writeError(w, http.StatusNotFound, "empresa nao encontrada")
		return
	}

	if action == "public-profile" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.getMyCompanyPublicProfile(w, r)
		return
	}

	if action != "" {
		writeError(w, http.StatusNotFound, "empresa nao encontrada")
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.getMyCompanyProfile(w, r)
	case http.MethodPut:
		a.updateMyCompanyProfile(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) getMyCompanyProfile(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(companyProfileSQL(), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o perfil da empresa")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) getMyCompanyPublicProfile(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(companyPublicProfileSQL(), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o perfil publico da empresa")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) updateMyCompanyProfile(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManageCompany() {
		writeError(w, http.StatusForbidden, "somente administrador da empresa pode editar o perfil da empresa")
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}

	var req updateCompanyProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.Website = strings.TrimSpace(req.Website)
	req.CompanySize = normalizeCompanySize(req.CompanySize)
	req.InstitutionalDescription = strings.TrimSpace(req.InstitutionalDescription)
	req.State = normalizeState(req.State)
	req.City = strings.TrimSpace(req.City)
	req.LogoDataURL = strings.TrimSpace(req.LogoDataURL)
	req.LogoFileName = strings.TrimSpace(req.LogoFileName)
	req.LogoMimeType = strings.TrimSpace(req.LogoMimeType)

	logoURL := ""
	if req.LogoDataURL != "" {
		var err error
		logoURL, err = saveCompanyLogo(req.LogoDataURL, req.LogoMimeType)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	payload, err := a.queryJSON(r.Context(), buildUpdateCompanyProfileSQL(session.CompanyID, session.UserID, req, logoURL))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel salvar o perfil: "+humanizeConstraintError(err.Error()))
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
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManageCompanyUsers() {
		writeError(w, http.StatusForbidden, "somente administrador da empresa pode listar usuarios")
		return
	}

	companyID := strings.TrimSpace(r.URL.Query().Get("companyId"))
	if session.RoleKey != "platform_admin" {
		companyID = session.CompanyID
	}
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
				COALESCE(photo.file_url, '') AS "profilePhotoUrl",
				u.status,
				u.created_at AS "createdAt",
				u.updated_at AS "updatedAt"
			FROM users u
			LEFT JOIN companies c ON c.id = u.company_id
			LEFT JOIN access_profiles p ON p.id = u.access_profile_id
			LEFT JOIN media_files photo ON photo.id = u.profile_photo_media_id
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
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManageCompanyUsers() {
		writeError(w, http.StatusForbidden, "somente administrador da empresa pode cadastrar usuarios")
		return
	}

	var req createCompanyUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	req.CompanyID = strings.TrimSpace(req.CompanyID)
	if session.RoleKey != "platform_admin" {
		req.CompanyID = session.CompanyID
	}
	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.JobTitle = strings.TrimSpace(req.JobTitle)
	req.AccessProfileKey = normalizeAccessProfileKey(req.AccessProfileKey)
	req.ProfilePhotoDataURL = strings.TrimSpace(req.ProfilePhotoDataURL)
	req.ProfilePhotoFileName = strings.TrimSpace(req.ProfilePhotoFileName)
	req.ProfilePhotoMimeType = strings.TrimSpace(req.ProfilePhotoMimeType)

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
	if req.ProfilePhotoDataURL == "" {
		writeError(w, http.StatusBadRequest, "foto do profissional e obrigatoria")
		return
	}
	if req.ProfilePhotoDataURL != "" {
		photoURL, err := saveProfilePhoto(req.ProfilePhotoDataURL, req.ProfilePhotoMimeType)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.ProfilePhotoURL = photoURL
	}

	token, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel gerar o link de acesso")
		return
	}

	payload, err := a.queryJSON(r.Context(), buildCreateCompanyUserSQL(req, hashSessionToken(token)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel criar o usuario: "+humanizeConstraintError(err.Error()))
		return
	}

	var result map[string]any
	_ = json.Unmarshal(payload, &result)
	baseURL := strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/")
	result["setupUrl"] = baseURL + "/#reset-password?token=" + token
	writeJSON(w, http.StatusCreated, result)
}

func (a *app) updateCompanyUser(w http.ResponseWriter, r *http.Request, userID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManageCompanyUsers() {
		writeError(w, http.StatusForbidden, "somente administrador da empresa pode editar usuarios")
		return
	}

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
	req.ProfilePhotoDataURL = strings.TrimSpace(req.ProfilePhotoDataURL)
	req.ProfilePhotoFileName = strings.TrimSpace(req.ProfilePhotoFileName)
	req.ProfilePhotoMimeType = strings.TrimSpace(req.ProfilePhotoMimeType)

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
	if req.ProfilePhotoDataURL == "" {
		hasPhoto, err := a.companyUserHasPhoto(r.Context(), userID, session)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel validar a foto do usuario")
			return
		}
		if !hasPhoto {
			writeError(w, http.StatusBadRequest, "foto do profissional e obrigatoria")
			return
		}
	}
	if req.ProfilePhotoDataURL != "" {
		photoURL, err := saveProfilePhoto(req.ProfilePhotoDataURL, req.ProfilePhotoMimeType)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.ProfilePhotoURL = photoURL
	}

	payload, err := a.queryJSON(r.Context(), buildUpdateCompanyUserSQL(userID, session, req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o usuario: "+humanizeConstraintError(err.Error()))
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) companyUserHasPhoto(ctx context.Context, userID string, session sessionUser) (bool, error) {
	companyFilter := ""
	if session.RoleKey != "platform_admin" {
		companyFilter = fmt.Sprintf("AND company_id = %s::uuid", sqlQuote(session.CompanyID))
	}
	payload, err := a.queryJSON(ctx, fmt.Sprintf(`
		SELECT json_build_object(
			'hasPhoto',
			EXISTS (
				SELECT 1
				FROM users
				WHERE id = %s::uuid
				  %s
				  AND profile_photo_media_id IS NOT NULL
				  AND deleted_at IS NULL
			)
		);
	`, sqlQuote(userID), companyFilter))
	if err != nil {
		return false, err
	}
	var result struct {
		HasPhoto bool `json:"hasPhoto"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return false, err
	}
	return result.HasPhoto, nil
}

func (a *app) changeCompanyUserStatus(w http.ResponseWriter, r *http.Request, userID string, status string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManageCompanyUsers() {
		writeError(w, http.StatusForbidden, "somente administrador da empresa pode alterar acesso de usuarios")
		return
	}

	status = normalizeUserStatus(status)
	companyFilter := ""
	if session.RoleKey != "platform_admin" {
		companyFilter = fmt.Sprintf("AND company_id = %s::uuid", sqlQuote(session.CompanyID))
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		UPDATE users
		SET
			status = %s,
			blocked_at = CASE WHEN %s = 'blocked' THEN now() ELSE NULL END,
			removed_at = CASE WHEN %s = 'removed' THEN now() ELSE removed_at END
		WHERE id = %s::uuid
		  %s
		  AND deleted_at IS NULL
		RETURNING row_to_json(users);
	`, sqlQuote(status), sqlQuote(status), sqlQuote(status), sqlQuote(userID), companyFilter))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o status do usuario")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleTenders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listTenders(w, r)
	case http.MethodPost:
		a.createTender(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleTenderByPath(w http.ResponseWriter, r *http.Request) {
	id, action := splitResourcePath(r.URL.Path, "/api/tenders/")
	if id == "" {
		writeError(w, http.StatusNotFound, "edital nao encontrado")
		return
	}
	if action == "interests" {
		switch r.Method {
		case http.MethodGet:
			a.listTenderInterests(w, r, id)
		case http.MethodPost:
			a.createTenderInterest(w, r, id)
		default:
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		}
		return
	}
	if action == "analysis" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.updateTenderAnalysis(w, r, id)
		return
	}
	if action != "" {
		writeError(w, http.StatusNotFound, "edital nao encontrado")
		return
	}
	if r.Method == http.MethodPut {
		a.updateTender(w, r, id)
		return
	}
	if r.Method == http.MethodDelete {
		a.deleteTender(w, r, id)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if err := a.refreshOccurredTenders(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar editais ocorridos")
		return
	}
	payload, err := a.queryJSON(r.Context(), tenderDetailSQL(id, session))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o edital")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "edital nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) updateTenderAnalysis(w http.ResponseWriter, r *http.Request, tenderID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.RoleKey != "platform_admin" {
		writeError(w, http.StatusForbidden, "apenas administrador da plataforma pode anexar analise")
		return
	}

	var req updateTenderAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.AnalysisDataURL = strings.TrimSpace(req.AnalysisDataURL)
	req.AnalysisFileName = strings.TrimSpace(req.AnalysisFileName)
	if req.AnalysisDataURL == "" {
		writeError(w, http.StatusBadRequest, "selecione o arquivo HTML da analise")
		return
	}
	analysisURL, err := saveTenderHTML(req.AnalysisDataURL, "analysis")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_tender AS (
			SELECT id
			FROM tenders
			WHERE id = %s::uuid AND deleted_at IS NULL
			LIMIT 1
		),
		inserted_file AS (
			INSERT INTO tender_files (
				tender_id, file_type, title, file_url, mime_type, is_downloadable, uploaded_by_user_id
			)
			SELECT
				id,
				'analysis_html',
				COALESCE(NULLIF(%s, ''), 'Analise do edital'),
				%s,
				'text/html',
				true,
				%s::uuid
			FROM selected_tender
			RETURNING *
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				u.company_id,
				'system',
				'Pre-analise adicionada',
				'A pre-analise HTML de um edital foi adicionada pela LicitaHub.',
				'tender-detail',
				'tender',
				st.id
			FROM selected_tender st
			JOIN users u ON u.company_id IS NOT NULL
			WHERE u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				tf.id::text AS id,
				tf.tender_id::text AS "tenderId",
				tf.file_url AS "analysisUrl",
				tf.title,
				tf.created_at AS "createdAt"
			FROM inserted_file tf
		) item;
	`, sqlQuote(tenderID), sqlQuote(req.AnalysisFileName), sqlQuote(analysisURL), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel anexar a analise: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "edital nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) listTenders(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if err := a.refreshOccurredTenders(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar editais ocorridos")
		return
	}
	myInterestSQL := "false"
	if strings.TrimSpace(session.CompanyID) != "" {
		myInterestSQL = fmt.Sprintf(`EXISTS (
					SELECT 1
					FROM tender_interests ti
					WHERE ti.tender_id = t.id
					  AND ti.company_id = %s::uuid
					  AND ti.deleted_at IS NULL
					  AND ti.status <> 'withdrawn'
				)`, sqlQuote(session.CompanyID))
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				t.id::text AS id,
				t.agency,
				t.number,
				t.object,
				COALESCE(t.modality, '') AS modality,
				COALESCE(t.judgment_criterion, '') AS "judgmentCriterion",
				t.estimated_value AS "estimatedValue",
				COALESCE(t.state, '') AS state,
				COALESCE(t.city, '') AS city,
				t.opening_date AS "openingDate",
				t.status,
				COALESCE(t.cloud_folder_url, '') AS "cloudFolderUrl",
				t.created_at AS "createdAt",
				%s AS "hasMyInterest"
			FROM tenders t
			WHERE t.deleted_at IS NULL
			ORDER BY COALESCE(t.opening_date, t.created_at) DESC
		) item;
	`, myInterestSQL))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar os editais")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createTender(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.RoleKey != "platform_admin" {
		writeError(w, http.StatusForbidden, "apenas administrador da plataforma pode cadastrar edital")
		return
	}
	var req createTenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Agency = strings.TrimSpace(req.Agency)
	req.Number = strings.TrimSpace(req.Number)
	req.Object = strings.TrimSpace(req.Object)
	req.Modality = strings.TrimSpace(req.Modality)
	req.JudgmentCriterion = strings.TrimSpace(req.JudgmentCriterion)
	req.EstimatedValue = strings.TrimSpace(req.EstimatedValue)
	req.State = normalizeState(req.State)
	req.City = strings.TrimSpace(req.City)
	req.OpeningDate = strings.TrimSpace(req.OpeningDate)
	req.Status = normalizeTenderStatus(req.Status)
	req.CloudFolderURL = strings.TrimSpace(req.CloudFolderURL)
	if req.Agency == "" || req.Number == "" || req.Object == "" {
		writeError(w, http.StatusBadRequest, "orgao, numero e objeto sao obrigatorios")
		return
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if !isValidTenderStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "status do edital invalido")
		return
	}
	analysisURL := ""
	if strings.TrimSpace(req.AnalysisDataURL) != "" {
		url, err := saveTenderHTML(req.AnalysisDataURL, "analysis")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		analysisURL = url
	}
	valueSQL := "NULL"
	if req.EstimatedValue != "" {
		valueSQL = sqlQuote(req.EstimatedValue) + "::numeric"
	}
	openingSQL := "NULL"
	if req.OpeningDate != "" {
		openingSQL = sqlQuote(req.OpeningDate+"T12:00:00-03:00") + "::timestamptz"
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH inserted_tender AS (
			INSERT INTO tenders (
				agency, number, object, modality, judgment_criterion,
				estimated_value, state, city, opening_date, status, cloud_folder_url, created_by_user_id
			) VALUES (
				%s, %s, %s, NULLIF(%s, ''), NULLIF(%s, ''),
				%s, NULLIF(%s, ''), NULLIF(%s, ''), %s, %s, NULLIF(%s, ''), %s::uuid
			)
			RETURNING *
		),
		inserted_file AS (
			INSERT INTO tender_files (tender_id, file_type, title, file_url, mime_type, is_downloadable)
			SELECT id, 'analysis_html', %s, %s, 'text/html', true
			FROM inserted_tender
			WHERE %s <> ''
			RETURNING id
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				u.company_id,
				'system',
				'Novo edital publicado',
				'Edital ' || it.number || ' foi publicado pela LicitaHub.',
				'tender-list',
				'tender',
				it.id
			FROM inserted_tender it
			JOIN users u ON u.company_id IS NOT NULL
			WHERE it.status = 'published'
			  AND u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT id::text AS id, agency, number, object, status
			FROM inserted_tender
		) item;
	`,
		sqlQuote(req.Agency), sqlQuote(req.Number), sqlQuote(req.Object),
		sqlQuote(req.Modality), sqlQuote(req.JudgmentCriterion), valueSQL,
		sqlQuote(req.State), sqlQuote(req.City), openingSQL, sqlQuote(req.Status),
		sqlQuote(req.CloudFolderURL), sqlQuote(session.UserID), sqlQuote(req.AnalysisFileName),
		sqlQuote(analysisURL), sqlQuote(analysisURL),
	))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel cadastrar o edital: "+err.Error())
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) updateTender(w http.ResponseWriter, r *http.Request, tenderID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.RoleKey != "platform_admin" {
		writeError(w, http.StatusForbidden, "apenas administrador da plataforma pode editar edital")
		return
	}
	var req createTenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Agency = strings.TrimSpace(req.Agency)
	req.Number = strings.TrimSpace(req.Number)
	req.Object = strings.TrimSpace(req.Object)
	req.Modality = strings.TrimSpace(req.Modality)
	req.JudgmentCriterion = strings.TrimSpace(req.JudgmentCriterion)
	req.EstimatedValue = strings.TrimSpace(req.EstimatedValue)
	req.State = normalizeState(req.State)
	req.City = strings.TrimSpace(req.City)
	req.OpeningDate = strings.TrimSpace(req.OpeningDate)
	req.Status = normalizeTenderStatus(req.Status)
	req.CloudFolderURL = strings.TrimSpace(req.CloudFolderURL)
	if req.Agency == "" || req.Number == "" || req.Object == "" {
		writeError(w, http.StatusBadRequest, "orgao, numero e objeto sao obrigatorios")
		return
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if !isValidTenderStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "status do edital invalido")
		return
	}
	valueSQL := "NULL"
	if req.EstimatedValue != "" {
		valueSQL = sqlQuote(req.EstimatedValue) + "::numeric"
	}
	openingSQL := "NULL"
	if req.OpeningDate != "" {
		openingSQL = sqlQuote(req.OpeningDate+"T12:00:00-03:00") + "::timestamptz"
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH updated_tender AS (
			UPDATE tenders
			SET
				agency = %s,
				number = %s,
				object = %s,
				modality = NULLIF(%s, ''),
				judgment_criterion = NULLIF(%s, ''),
				estimated_value = %s,
				state = NULLIF(%s, ''),
				city = NULLIF(%s, ''),
				opening_date = %s,
				status = %s,
				cloud_folder_url = NULLIF(%s, ''),
				updated_at = now()
			WHERE id = %s::uuid
			  AND deleted_at IS NULL
			RETURNING *
		)
		SELECT row_to_json(item)
		FROM (
			SELECT id::text AS id, agency, number, object, status
			FROM updated_tender
		) item;
	`,
		sqlQuote(req.Agency), sqlQuote(req.Number), sqlQuote(req.Object),
		sqlQuote(req.Modality), sqlQuote(req.JudgmentCriterion), valueSQL,
		sqlQuote(req.State), sqlQuote(req.City), openingSQL, sqlQuote(req.Status),
		sqlQuote(req.CloudFolderURL), sqlQuote(tenderID),
	))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel editar o edital: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "edital nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) deleteTender(w http.ResponseWriter, r *http.Request, tenderID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.RoleKey != "platform_admin" {
		writeError(w, http.StatusForbidden, "apenas administrador da plataforma pode excluir edital")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH deleted_tender AS (
			UPDATE tenders
			SET deleted_at = now(), updated_at = now(), status = 'cancelled'
			WHERE id = %s::uuid
			  AND deleted_at IS NULL
			RETURNING *
		)
		SELECT row_to_json(item)
		FROM (
			SELECT id::text AS id, number, status
			FROM deleted_tender
		) item;
	`, sqlQuote(tenderID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel excluir o edital")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "edital nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) refreshOccurredTenders(ctx context.Context) error {
	_, err := a.runPSQL(ctx, `
		UPDATE tenders
		SET status = 'occurred', updated_at = now()
		WHERE status IN ('published', 'under_review', 'suspended', 'challenged')
		  AND opening_date IS NOT NULL
		  AND opening_date < now()
		  AND deleted_at IS NULL;
	`)
	return err
}

func tenderDetailSQL(id string, session sessionUser) string {
	myInterestSQL := "false"
	if strings.TrimSpace(session.CompanyID) != "" {
		myInterestSQL = fmt.Sprintf(`EXISTS (
					SELECT 1
					FROM tender_interests ti
					WHERE ti.tender_id = t.id
					  AND ti.company_id = %s::uuid
					  AND ti.deleted_at IS NULL
					  AND ti.status <> 'withdrawn'
				)`, sqlQuote(session.CompanyID))
	}
	return fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				t.id::text AS id, t.agency, t.number, t.object,
				COALESCE(t.modality, '') AS modality,
				COALESCE(t.judgment_criterion, '') AS "judgmentCriterion",
				t.estimated_value AS "estimatedValue",
				COALESCE(t.state, '') AS state,
				COALESCE(t.city, '') AS city,
				t.opening_date AS "openingDate",
				t.status,
				COALESCE(t.cloud_folder_url, '') AS "cloudFolderUrl",
				COALESCE((SELECT file_url FROM tender_files WHERE tender_id=t.id AND file_type='analysis_html' ORDER BY created_at DESC LIMIT 1), '') AS "analysisUrl",
				%s AS "hasMyInterest"
			FROM tenders t
			WHERE t.id = %s::uuid AND t.deleted_at IS NULL
			LIMIT 1
		) item;
	`, myInterestSQL, sqlQuote(id))
}

func (a *app) listTenderInterests(w http.ResponseWriter, r *http.Request, tenderID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), partnershipAdsSQL(fmt.Sprintf("AND pa.tender_id = %s::uuid", sqlQuote(tenderID)), session))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar empresas interessadas")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createTenderInterest(w http.ResponseWriter, r *http.Request, tenderID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}

	var req createTenderInterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.GeneralPosition = normalizeInterestGeneralPosition(req.GeneralPosition)
	req.DesiredRole = normalizeInterestDesiredRole(req.DesiredRole)
	req.PublicSummary = strings.TrimSpace(req.PublicSummary)
	req.InternalNote = strings.TrimSpace(req.InternalNote)
	if req.PublicSummary == "" {
		writeError(w, http.StatusBadRequest, "resumo do anuncio e obrigatorio")
		return
	}

	requirementsJSON, err := json.Marshal(req.Requirements)
	if err != nil {
		writeError(w, http.StatusBadRequest, "requisitos invalidos")
		return
	}
	payload, err := a.queryJSON(r.Context(), buildCreateTenderInterestSQL(tenderID, session, req, string(requirementsJSON)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel registrar interesse: "+err.Error())
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) handlePartnershipAds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), partnershipAdsSQL("", session))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar a vitrine de parceiros")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handlePartnershipAdByPath(w http.ResponseWriter, r *http.Request) {
	id, action := splitResourcePath(r.URL.Path, "/api/partnership-ads/")
	if id == "" {
		writeError(w, http.StatusNotFound, "anuncio nao encontrado")
		return
	}
	if action == "evaluate" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.evaluatePartnershipAd(w, r, id)
		return
	}
	if action != "" {
		writeError(w, http.StatusNotFound, "anuncio nao encontrado")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if r.Method == http.MethodPut {
		a.updatePartnershipAd(w, r, id, session)
		return
	}
	if r.Method == http.MethodDelete {
		a.deletePartnershipAd(w, r, id, session)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	payload, err := a.queryJSON(r.Context(), partnershipAdDetailSQL(id))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o anuncio")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "anuncio nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) updatePartnershipAd(w http.ResponseWriter, r *http.Request, adID string, session sessionUser) {
	if session.CompanyID == "" || !session.canUseChat() {
		writeError(w, http.StatusForbidden, "seu perfil nao pode editar anuncios")
		return
	}
	var req updatePartnershipAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.OfferSummary = strings.TrimSpace(req.OfferSummary)
	req.SeekSummary = strings.TrimSpace(req.SeekSummary)
	if req.OfferSummary == "" && req.SeekSummary == "" {
		writeError(w, http.StatusBadRequest, "informe o que sua empresa oferece ou busca")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		UPDATE partnership_ads pa
		SET offer_summary = %s,
			seek_summary = %s,
			updated_at = now()
		WHERE pa.id = %s::uuid
		  AND pa.company_id = %s::uuid
		  AND pa.status = 'published'
		  AND pa.deleted_at IS NULL
		  AND COALESCE(pa.ad_type, 'company') = 'company'
		RETURNING row_to_json(pa);
	`, sqlQuote(req.OfferSummary), sqlQuote(req.SeekSummary), sqlQuote(adID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel editar o anuncio")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "somente anuncios ativos da sua empresa podem ser editados aqui")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) deletePartnershipAd(w http.ResponseWriter, r *http.Request, adID string, session sessionUser) {
	if session.CompanyID == "" || !session.canUseChat() {
		writeError(w, http.StatusForbidden, "seu perfil nao pode excluir anuncios")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		UPDATE partnership_ads pa
		SET status = 'closed', deleted_at = now(), updated_at = now()
		WHERE pa.id = %s::uuid
		  AND pa.company_id = %s::uuid
		  AND pa.status = 'published'
		  AND pa.deleted_at IS NULL
		  AND COALESCE(pa.ad_type, 'company') = 'company'
		RETURNING json_build_object('id', pa.id::text, 'status', pa.status);
	`, sqlQuote(adID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel excluir o anuncio")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "somente anuncios ativos da sua empresa podem ser excluidos aqui")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) evaluatePartnershipAd(w http.ResponseWriter, r *http.Request, adID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}
	var req partnerEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Decision = strings.ToLower(strings.TrimSpace(req.Decision))
	switch req.Decision {
	case "liked", "rejected", "later":
	default:
		writeError(w, http.StatusBadRequest, "decisao invalida")
		return
	}
	adPayload, err := a.queryJSON(r.Context(), partnershipAdKindSQL(adID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar tipo do anuncio")
		return
	}
	if strings.TrimSpace(string(adPayload)) == "null" {
		writeError(w, http.StatusNotFound, "anuncio nao encontrado")
		return
	}
	eligibilityPayload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT json_build_object(
			'hasPublishedCompanyAd', EXISTS (
				SELECT 1
				FROM partnership_ads own_ad
				JOIN partnership_ads target_ad ON target_ad.id = %s::uuid
				WHERE own_ad.company_id = %s::uuid
				  AND own_ad.tender_id = target_ad.tender_id
				  AND own_ad.deleted_at IS NULL
				  AND own_ad.status = 'published'
				  AND COALESCE(own_ad.ad_type, 'company') IN ('company', 'consortium')
			)
		);
	`, sqlQuote(adID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel validar o interesse da empresa")
		return
	}
	var eligibility struct {
		HasPublishedCompanyAd bool `json:"hasPublishedCompanyAd"`
	}
	if json.Unmarshal(eligibilityPayload, &eligibility) != nil || !eligibility.HasPublishedCompanyAd {
		writeError(w, http.StatusConflict, "registre o interesse da sua empresa neste edital antes de avaliar candidatas")
		return
	}
	var adKind struct {
		AdType string `json:"adType"`
	}
	if err := json.Unmarshal(adPayload, &adKind); err != nil {
		writeError(w, http.StatusInternalServerError, "tipo do anuncio invalido")
		return
	}
	var payload []byte
	if adKind.AdType == "consortium" {
		payload, err = a.queryJSON(r.Context(), buildEvaluateConsortiumAdSQL(adID, session, req.Decision))
	} else {
		payload, err = a.queryJSON(r.Context(), buildEvaluatePartnershipAdSQL(adID, session, req.Decision))
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel registrar avaliacao: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusConflict, a.partnershipEvaluationUnavailableReason(r.Context(), adID, session))
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) partnershipEvaluationUnavailableReason(ctx context.Context, adID string, session sessionUser) string {
	payload, err := a.queryJSON(ctx, fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				COALESCE(current_company.trade_name, 'sem empresa vinculada') AS "currentCompanyName",
				COALESCE(owner_company.trade_name, '') AS "adOwnerCompanyName",
				COALESCE(pa.status, 'nao encontrado') AS "adStatus",
				COALESCE(pa.ad_type, '') AS "adType",
				EXISTS (
					SELECT 1
					FROM consortium_members cm
					WHERE cm.consortium_intention_id = pa.consortium_intention_id
					  AND cm.company_id = %s::uuid
				) AS "alreadyMember",
				(pa.company_id = %s::uuid) AS "isOwnAd"
			FROM partnership_ads pa
			LEFT JOIN companies owner_company ON owner_company.id = pa.company_id
			LEFT JOIN companies current_company ON current_company.id = %s::uuid
			WHERE pa.id = %s::uuid
			LIMIT 1
		) item;
	`, sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(adID)))
	if err != nil || strings.TrimSpace(string(payload)) == "null" {
		return "este anuncio nao pode receber sua avaliacao agora"
	}
	var diagnostic struct {
		CurrentCompanyName string `json:"currentCompanyName"`
		AdOwnerCompanyName string `json:"adOwnerCompanyName"`
		AdStatus           string `json:"adStatus"`
		AlreadyMember      bool   `json:"alreadyMember"`
		IsOwnAd            bool   `json:"isOwnAd"`
	}
	if json.Unmarshal(payload, &diagnostic) != nil {
		return "este anuncio nao pode receber sua avaliacao agora"
	}
	if diagnostic.IsOwnAd {
		return "voce esta conectado como " + diagnostic.CurrentCompanyName + ". Este e o anuncio da propria empresa, por isso nao pode ser avaliado"
	}
	if diagnostic.AlreadyMember {
		return "voce esta conectado como " + diagnostic.CurrentCompanyName + ", que ja faz parte deste consorcio"
	}
	if diagnostic.AdStatus != "published" {
		return "o anuncio nao esta mais publicado para receber candidaturas"
	}
	return "o sistema reconheceu voce como " + diagnostic.CurrentCompanyName + ", mas nao conseguiu registrar a candidatura. Tente atualizar a pagina e abrir novamente o anuncio do consorcio"
}

func (a *app) handleChats(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canUseChat() {
		writeError(w, http.StatusForbidden, "perfil sem permissao para usar conversas de parceria")
		return
	}
	switch r.Method {
	case http.MethodGet:
		payload, err := a.queryJSON(r.Context(), chatThreadsSQL(session, ""))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel carregar conversas")
			return
		}
		writeRawJSON(w, http.StatusOK, payload)
	case http.MethodPost:
		var req startChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "json invalido")
			return
		}
		req.AdID = strings.TrimSpace(req.AdID)
		if req.AdID == "" {
			writeError(w, http.StatusBadRequest, "anuncio obrigatorio")
			return
		}
		payload, err := a.queryJSON(r.Context(), buildStartChatSQL(session, req.AdID))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel iniciar conversa: "+err.Error())
			return
		}
		if strings.TrimSpace(string(payload)) == "null" {
			writeError(w, http.StatusNotFound, "anuncio indisponivel para conversa")
			return
		}
		a.broadcastChatThreadPayload(payload)
		writeRawJSON(w, http.StatusCreated, payload)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleChatByPath(w http.ResponseWriter, r *http.Request) {
	threadID, action := splitResourcePath(r.URL.Path, "/api/chats/")
	if threadID == "" {
		writeError(w, http.StatusNotFound, "conversa nao encontrada")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canUseChat() {
		writeError(w, http.StatusForbidden, "perfil sem permissao para usar conversas de parceria")
		return
	}
	if action != "messages" {
		writeError(w, http.StatusNotFound, "acao de conversa nao encontrada")
		return
	}
	switch r.Method {
	case http.MethodGet:
		payload, err := a.queryJSON(r.Context(), chatMessagesSQL(session, threadID))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel carregar mensagens")
			return
		}
		writeRawJSON(w, http.StatusOK, payload)
	case http.MethodPost:
		var req createChatMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "json invalido")
			return
		}
		req.Content = strings.TrimSpace(req.Content)
		if req.Content == "" {
			writeError(w, http.StatusBadRequest, "mensagem obrigatoria")
			return
		}
		if len([]rune(req.Content)) > 2000 {
			writeError(w, http.StatusBadRequest, "mensagem muito longa")
			return
		}
		payload, err := a.queryJSON(r.Context(), buildCreateChatMessageSQL(session, threadID, req.Content))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel enviar mensagem: "+err.Error())
			return
		}
		if strings.TrimSpace(string(payload)) == "null" {
			writeError(w, http.StatusNotFound, "conversa indisponivel")
			return
		}
		a.broadcastChatMessagePayload(payload)
		writeRawJSON(w, http.StatusCreated, payload)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canUseChat() {
		writeError(w, http.StatusForbidden, "perfil sem permissao para usar conversas de parceria")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "stream nao suportado")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := a.chatHub.add(session.CompanyID)
	defer a.chatHub.remove(session.CompanyID, ch)
	fmt.Fprint(w, "event: ready\ndata: {}\n\n")
	flusher.Flush()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case payload := <-ch:
			fmt.Fprintf(w, "event: chat-message\ndata: %s\n\n", payload)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (a *app) broadcastChatThreadPayload(payload []byte) {
	var data struct {
		CompanyAID string `json:"companyAId"`
		CompanyBID string `json:"companyBId"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return
	}
	a.chatHub.broadcast(data.CompanyAID, payload)
	a.chatHub.broadcast(data.CompanyBID, payload)
}

func (a *app) broadcastChatMessagePayload(payload []byte) {
	var data struct {
		CompanyAID string `json:"companyAId"`
		CompanyBID string `json:"companyBId"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return
	}
	a.chatHub.broadcast(data.CompanyAID, payload)
	a.chatHub.broadcast(data.CompanyBID, payload)
}

func (a *app) handleMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), matchesSQL(session, ""))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar matches")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleMatchByPath(w http.ResponseWriter, r *http.Request) {
	id, action := splitResourcePath(r.URL.Path, "/api/matches/")
	if id == "" {
		writeError(w, http.StatusNotFound, "match nao encontrado")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if action == "leader" {
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.updateConsortiumLeader(w, r, id, session)
		return
	}
	if action == "consortium-ad" {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.createConsortiumAd(w, r, id, session)
		return
	}
	if action == "application-accept" {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.acceptConsortiumApplication(w, r, id, session)
		return
	}
	if action == "withdraw" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.withdrawFromConsortium(w, r, id, session)
		return
	}
	if action != "" || r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "match nao encontrado")
		return
	}
	payload, err := a.queryJSON(r.Context(), matchesSQL(session, fmt.Sprintf("AND m.id = %s::uuid", sqlQuote(id))))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar match")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) updateConsortiumLeader(w http.ResponseWriter, r *http.Request, matchID string, session sessionUser) {
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}
	var req updateConsortiumLeaderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.LeadCompanyID = strings.TrimSpace(req.LeadCompanyID)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.LeadCompanyID == "" {
		writeError(w, http.StatusBadRequest, "selecione a empresa lider do consorcio")
		return
	}
	payload, err := a.queryJSON(r.Context(), buildUpdateConsortiumLeaderSQL(matchID, session, req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel definir lider do consorcio: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "match nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) acceptConsortiumApplication(w http.ResponseWriter, r *http.Request, matchID string, session sessionUser) {
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}
	var req acceptConsortiumApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.ApplicationID = strings.TrimSpace(req.ApplicationID)
	if req.ApplicationID == "" {
		writeError(w, http.StatusBadRequest, "candidatura obrigatoria")
		return
	}
	payload, err := a.queryJSON(r.Context(), buildAcceptConsortiumApplicationSQL(matchID, session, req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel aceitar candidata: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "apenas a lider pode aceitar esta candidata")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) withdrawFromConsortium(w http.ResponseWriter, r *http.Request, matchID string, session sessionUser) {
	if !session.canManageCompany() {
		writeError(w, http.StatusForbidden, "somente o administrador da empresa pode desistir do consorcio")
		return
	}
	var req withdrawFromConsortiumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.SuccessorCompanyID = strings.TrimSpace(req.SuccessorCompanyID)
	payload, err := a.queryJSON(r.Context(), buildWithdrawFromConsortiumSQL(matchID, session, req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel registrar a desistência do consorcio: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusBadRequest, "para a lider desistir, escolha antes outra empresa ativa como lider do consorcio")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createConsortiumAd(w http.ResponseWriter, r *http.Request, matchID string, session sessionUser) {
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}
	var req createConsortiumAdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.NeedSummary = strings.TrimSpace(req.NeedSummary)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.NeedSummary == "" {
		writeError(w, http.StatusBadRequest, "descreva o que falta complementar no consorcio")
		return
	}
	payload, err := a.queryJSON(r.Context(), buildCreateConsortiumAdSQL(matchID, session, req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel criar anuncio do consorcio: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "apenas a empresa lider pode criar anuncio para este consorcio")
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
				n.expires_at AS "expiresAt",
				n.created_at AS "createdAt",
				n.updated_at AS "updatedAt"
			FROM news n
			LEFT JOIN news_categories c ON c.id = n.category_id
			LEFT JOIN media_files m ON m.id = n.main_image_media_id
			WHERE n.deleted_at IS NULL
			  AND n.status IN ('published', 'featured')
			  AND n.expires_at >= now()
			ORDER BY
				CASE WHEN n.status = 'featured' THEN 0 ELSE 1 END,
				COALESCE(n.published_at, n.created_at) DESC
		) item;
	`

	payload, err := a.queryJSON(r.Context(), sql)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar as noticias")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleAdminNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode gerenciar noticias")
		return
	}

	payload, err := a.queryJSON(r.Context(), `
		WITH expired_news AS (
			UPDATE news
			SET status = 'expired', archived_at = COALESCE(archived_at, now()), updated_at = now()
			WHERE deleted_at IS NULL
			  AND status IN ('published', 'featured')
			  AND expires_at < now()
			RETURNING id
		)
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				n.id::text AS id,
				n.title,
				COALESCE(c.name, '') AS "categoryName",
				n.status,
				n.published_at AS "publishedAt",
				n.expires_at AS "expiresAt",
				n.created_at AS "createdAt",
				n.updated_at AS "updatedAt"
			FROM news n
			LEFT JOIN news_categories c ON c.id = n.category_id
			WHERE n.deleted_at IS NULL
			ORDER BY n.created_at DESC
		) item;
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar todas as noticias")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleNewsByPath(w http.ResponseWriter, r *http.Request) {
	newsID, action := splitResourcePath(r.URL.Path, "/api/news/")
	if newsID == "" || action != "status" || r.Method != http.MethodPatch {
		writeError(w, http.StatusNotFound, "rota nao encontrada")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode alterar noticias")
		return
	}

	var req updateNewsStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}

	status, err := normalizeNewsStatus(req.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	expiresSQL := "NULL"
	if status == "published" || status == "featured" {
		expiresAt, err := parseNewsExpiresAt(req.ExpiresAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if expiresAt.Before(time.Now()) {
			writeError(w, http.StatusBadRequest, "a data final de publicacao deve ser hoje ou uma data futura")
			return
		}
		expiresSQL = sqlQuote(expiresAt.Format(time.RFC3339))
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH previous_featured AS (
			UPDATE news
			SET status = 'published', updated_at = now()
			WHERE %s = 'featured'
			  AND status = 'featured'
			  AND id <> %s::uuid
			  AND deleted_at IS NULL
			RETURNING id
		),
		updated_news AS (
			UPDATE news
			SET
				status = %s,
				expires_at = %s::timestamptz,
				published_at = CASE
					WHEN %s IN ('published', 'featured') THEN COALESCE(published_at, now())
					ELSE published_at
				END,
				archived_at = CASE
					WHEN %s IN ('archived', 'expired') THEN now()
					ELSE NULL
				END,
				updated_at = now()
			WHERE id = %s::uuid
			  AND deleted_at IS NULL
			RETURNING id, status, expires_at, updated_at
		)
		SELECT json_build_object(
			'id', id::text,
			'status', status,
			'expiresAt', expires_at,
			'updatedAt', updated_at
		)
		FROM updated_news;
	`, sqlQuote(status), sqlQuote(newsID), sqlQuote(status), expiresSQL, sqlQuote(status), sqlQuote(status), sqlQuote(newsID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o status da noticia")
		return
	}

	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createNews(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode cadastrar noticias")
		return
	}

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

	if status == "published" || status == "featured" {
		expiresAt, err := parseNewsExpiresAt(req.ExpiresAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if expiresAt.Before(time.Now()) {
			writeError(w, http.StatusBadRequest, "a data final de publicacao deve ser hoje ou uma data futura")
			return
		}
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

func (a *app) handlePostCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if _, ok := a.currentSessionUser(w, r); !ok {
		return
	}

	payload, err := a.queryJSON(r.Context(), `
		SELECT COALESCE(json_agg(row_to_json(category)), '[]'::json)
		FROM (
			SELECT id::text AS id, name, slug
			FROM post_categories
			WHERE is_active = true
			ORDER BY CASE slug
				WHEN 'equipe-comercial' THEN 1
				WHEN 'noticias' THEN 2
				WHEN 'atividades' THEN 3
				WHEN 'eventos' THEN 4
				WHEN 'conquistas' THEN 5
				WHEN 'conteudo-tecnico' THEN 6
				WHEN 'destaque' THEN 7
				ELSE 99
			END, name
		) category;
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar categorias da comunidade")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleCommunityPosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listCommunityPosts(w, r)
	case http.MethodPost:
		a.createCommunityPost(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) listCommunityPosts(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}

	var filters []string
	query := r.URL.Query()
	scope := strings.ToLower(strings.TrimSpace(query.Get("scope")))
	if scope == "mine" {
		if session.CompanyID == "" {
			writeRawJSON(w, http.StatusOK, []byte("[]"))
			return
		}
		filters = append(filters, fmt.Sprintf("p.company_id = %s::uuid", sqlQuote(session.CompanyID)))
	} else {
		filters = append(filters, "p.visibility IN ('community', 'both')")
	}

	if category := strings.TrimSpace(query.Get("categorySlug")); category != "" && category != "all" {
		filters = append(filters, fmt.Sprintf("pc.slug = %s", sqlQuote(category)))
	}
	if company := strings.TrimSpace(query.Get("company")); company != "" {
		filters = append(filters, fmt.Sprintf("lower(c.trade_name) LIKE lower(%s)", sqlQuote("%"+company+"%")))
	}
	if state := normalizeState(query.Get("state")); state != "" && state != "BR" {
		filters = append(filters, fmt.Sprintf("COALESCE(cp.state, c.state, '') = %s", sqlQuote(state)))
	}

	whereExtra := ""
	if len(filters) > 0 {
		whereExtra = "AND " + strings.Join(filters, " AND ")
	}

	payload, err := a.queryJSON(r.Context(), communityPostsSQL(session, whereExtra))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar publicacoes da comunidade: "+err.Error())
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createCommunityPost(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.CategorySlug = strings.TrimSpace(strings.ToLower(req.CategorySlug))
	req.Visibility = normalizePostVisibility(req.Visibility)
	req.Content = strings.TrimSpace(req.Content)
	req.MainImageURL = strings.TrimSpace(req.MainImageURL)
	req.MainImageDataURL = strings.TrimSpace(req.MainImageDataURL)
	req.MainImageFileName = strings.TrimSpace(req.MainImageFileName)
	req.MainImageMimeType = strings.TrimSpace(req.MainImageMimeType)

	if req.CategorySlug == "" {
		writeError(w, http.StatusBadRequest, "tipo de publicacao e obrigatorio")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "texto da publicacao e obrigatorio")
		return
	}

	if req.MainImageDataURL != "" {
		imageURL, err := saveCommunityImage(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.MainImageURL = imageURL
	}

	payload, err := a.queryJSON(r.Context(), buildCreateCommunityPostSQL(session, req))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel publicar na comunidade: "+err.Error())
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeJSON(w, http.StatusCreated, map[string]string{"status": "published"})
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) handleCommunityPostByPath(w http.ResponseWriter, r *http.Request) {
	postID, action := splitResourcePath(r.URL.Path, "/api/community/posts/")
	if postID == "" {
		writeError(w, http.StatusNotFound, "rota nao encontrada")
		return
	}

	if action == "" {
		switch r.Method {
		case http.MethodPut:
			a.updateCommunityPost(w, r, postID)
		case http.MethodDelete:
			a.deleteCommunityPost(w, r, postID)
		default:
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		}
		return
	}

	if action == "comments" {
		remainder := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/community/posts/"+postID+"/comments"), "/")
		if remainder != "" {
			commentID := strings.Split(remainder, "/")[0]
			switch r.Method {
			case http.MethodPut:
				a.updatePostComment(w, r, postID, commentID)
			case http.MethodDelete:
				a.deletePostComment(w, r, postID, commentID)
			default:
				writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			}
			return
		}
	}

	switch action {
	case "like":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.togglePostLike(w, r, postID)
	case "favorite":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.togglePostFavorite(w, r, postID)
	case "comments":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.createPostComment(w, r, postID)
	case "archive":
		if r.Method != http.MethodPatch {
			writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
			return
		}
		a.archiveCommunityPost(w, r, postID)
	default:
		writeError(w, http.StatusNotFound, "rota nao encontrada")
	}
}

func (a *app) updateCommunityPost(w http.ResponseWriter, r *http.Request, postID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "usuario sem empresa vinculada")
		return
	}
	var req updatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.CategorySlug = strings.TrimSpace(strings.ToLower(req.CategorySlug))
	req.Visibility = normalizePostVisibility(req.Visibility)
	req.Content = strings.TrimSpace(req.Content)
	req.MainImageURL = strings.TrimSpace(req.MainImageURL)
	req.MainImageDataURL = strings.TrimSpace(req.MainImageDataURL)
	req.MainImageFileName = strings.TrimSpace(req.MainImageFileName)
	req.MainImageMimeType = strings.TrimSpace(req.MainImageMimeType)
	if req.CategorySlug == "" {
		writeError(w, http.StatusBadRequest, "tipo de publicacao e obrigatorio")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "texto da publicacao e obrigatorio")
		return
	}
	if req.MainImageDataURL != "" {
		imageURL, err := saveCommunityImage(createPostRequest{
			MainImageDataURL:  req.MainImageDataURL,
			MainImageFileName: req.MainImageFileName,
			MainImageMimeType: req.MainImageMimeType,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.MainImageURL = imageURL
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_category AS (
			SELECT id
			FROM post_categories
			WHERE is_active = true
			  AND slug IN (%s, 'noticias')
			ORDER BY CASE WHEN slug = %s THEN 0 ELSE 1 END
			LIMIT 1
		),
		inserted_media AS (
			INSERT INTO media_files (company_id, uploaded_by_user_id, media_type, file_name, file_url, mime_type, source)
			SELECT %s::uuid, %s::uuid, 'image', %s, %s, %s, 'upload'
			WHERE %s IS NOT NULL
			RETURNING id, file_url
		),
		updated_post AS (
			UPDATE posts p
			SET
				category_id = (SELECT id FROM selected_category),
				title = %s,
				content = %s,
				visibility = %s,
				main_image_media_id = COALESCE((SELECT id FROM inserted_media), p.main_image_media_id),
				updated_at = now()
			WHERE p.id = %s::uuid
			  AND p.company_id = %s::uuid
			  AND p.deleted_at IS NULL
			  AND p.status = 'published'
			  AND EXISTS (SELECT 1 FROM selected_category)
			RETURNING p.id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				p.id::text AS id,
				p.title,
				p.content AS text,
				p.visibility,
				pc.name AS category,
				pc.slug AS "categorySlug",
				COALESCE(media.file_url, '') AS "imageUrl",
				p.updated_at AS "updatedAt"
			FROM updated_post up
			JOIN posts p ON p.id = up.id
			LEFT JOIN post_categories pc ON pc.id = p.category_id
			LEFT JOIN media_files media ON media.id = p.main_image_media_id
		) item;
	`,
		sqlQuote(req.CategorySlug),
		sqlQuote(req.CategorySlug),
		sqlQuote(session.CompanyID),
		sqlQuote(session.UserID),
		sqlQuote(firstNonEmpty(req.MainImageFileName, "imagem-publicacao")),
		nullOrQuote(req.MainImageURL),
		nullOrQuote(req.MainImageMimeType),
		nullOrQuote(req.MainImageURL),
		nullOrQuote(req.Title),
		sqlQuote(req.Content),
		sqlQuote(req.Visibility),
		sqlQuote(postID),
		sqlQuote(session.CompanyID),
	))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel editar publicacao")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "publicacao nao encontrada para sua empresa")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) archiveCommunityPost(w http.ResponseWriter, r *http.Request, postID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH archived_post AS (
			UPDATE posts
			SET status = 'archived', updated_at = now()
			WHERE id = %s::uuid
			  AND company_id = %s::uuid
			  AND deleted_at IS NULL
			RETURNING id, status, updated_at
		)
		SELECT row_to_json(item)
		FROM (
			SELECT id::text AS id, status, updated_at AS "updatedAt"
			FROM archived_post
		) item;
	`, sqlQuote(postID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel arquivar publicacao")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "publicacao nao encontrada para sua empresa")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) deleteCommunityPost(w http.ResponseWriter, r *http.Request, postID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH deleted_post AS (
			UPDATE posts
			SET deleted_at = now(), updated_at = now()
			WHERE id = %s::uuid
			  AND company_id = %s::uuid
			  AND deleted_at IS NULL
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT id::text AS id, true AS deleted
			FROM deleted_post
		) item;
	`, sqlQuote(postID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel excluir publicacao")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "publicacao nao encontrada para sua empresa")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) togglePostLike(w http.ResponseWriter, r *http.Request, postID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH target_post AS (
			SELECT id FROM posts WHERE id = %s::uuid AND deleted_at IS NULL AND status = 'published'
		),
		deleted_like AS (
			DELETE FROM post_likes
			WHERE post_id = (SELECT id FROM target_post)
			  AND user_id = %s::uuid
			RETURNING id
		),
		inserted_like AS (
			INSERT INTO post_likes (post_id, user_id, company_id)
			SELECT id, %s::uuid, NULLIF(%s, '')::uuid
			FROM target_post
			WHERE NOT EXISTS (SELECT 1 FROM deleted_like)
			RETURNING id
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				owner_user.id,
				p.company_id,
				'post_like',
				'Nova curtida',
				COALESCE(liker_company.trade_name, liker_user.full_name, 'Uma empresa') || ' curtiu uma publicacao da sua empresa.',
				'publication-list',
				'post',
				p.id
			FROM inserted_like il
			JOIN posts p ON p.id = %s::uuid
			JOIN users owner_user ON owner_user.company_id = p.company_id
			JOIN users liker_user ON liker_user.id = %s::uuid
			LEFT JOIN companies liker_company ON liker_company.id = NULLIF(%s, '')::uuid
			WHERE owner_user.status = 'active'
			  AND owner_user.deleted_at IS NULL
			  AND owner_user.id <> %s::uuid
			RETURNING id
		)
		SELECT json_build_object(
			'liked', EXISTS (SELECT 1 FROM inserted_like),
			'likeCount', (SELECT count(*) FROM post_likes WHERE post_id = %s::uuid)
		);
	`, sqlQuote(postID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(postID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(postID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar curtida")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) togglePostFavorite(w http.ResponseWriter, r *http.Request, postID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH target_post AS (
			SELECT id FROM posts WHERE id = %s::uuid AND deleted_at IS NULL AND status = 'published'
		),
		deleted_favorite AS (
			DELETE FROM post_favorites
			WHERE post_id = (SELECT id FROM target_post)
			  AND user_id = %s::uuid
			RETURNING id
		),
		inserted_favorite AS (
			INSERT INTO post_favorites (post_id, user_id)
			SELECT id, %s::uuid
			FROM target_post
			WHERE NOT EXISTS (SELECT 1 FROM deleted_favorite)
			RETURNING id
		)
		SELECT json_build_object(
			'favorited', EXISTS (SELECT 1 FROM inserted_favorite)
		);
	`, sqlQuote(postID), sqlQuote(session.UserID), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar favorito")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) createPostComment(w http.ResponseWriter, r *http.Request, postID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	var req createPostCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "comentario vazio")
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH target_post AS (
			SELECT id FROM posts WHERE id = %s::uuid AND deleted_at IS NULL AND status = 'published'
		),
		inserted_comment AS (
			INSERT INTO post_comments (post_id, user_id, company_id, content)
			SELECT id, %s::uuid, NULLIF(%s, '')::uuid, %s
			FROM target_post
			RETURNING *
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				owner_user.id,
				p.company_id,
				'post_comment',
				'Novo comentario',
				COALESCE(comment_company.trade_name, comment_user.full_name, 'Uma empresa') || ' comentou uma publicacao da sua empresa.',
				'publication-list',
				'post',
				p.id
			FROM inserted_comment ic
			JOIN posts p ON p.id = ic.post_id
			JOIN users owner_user ON owner_user.company_id = p.company_id
			JOIN users comment_user ON comment_user.id = ic.user_id
			LEFT JOIN companies comment_company ON comment_company.id = ic.company_id
			WHERE owner_user.status = 'active'
			  AND owner_user.deleted_at IS NULL
			  AND owner_user.id <> ic.user_id
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				pc.id::text AS id,
				COALESCE(c.trade_name, u.full_name, 'Usuario') AS company,
				u.full_name AS "userName",
				pc.content AS text,
				(pc.user_id = %s::uuid) AS "canEdit",
				true AS "canDelete",
				pc.created_at AS "createdAt"
			FROM inserted_comment pc
			JOIN users u ON u.id = pc.user_id
			LEFT JOIN companies c ON c.id = pc.company_id
		) item;
	`, sqlQuote(postID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(req.Content), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel comentar")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "publicacao nao encontrada")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) updatePostComment(w http.ResponseWriter, r *http.Request, postID string, commentID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	var req createPostCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "comentario vazio")
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH updated_comment AS (
			UPDATE post_comments pc
			SET content = %s, updated_at = now()
			WHERE pc.id = %s::uuid
			  AND pc.post_id = %s::uuid
			  AND pc.user_id = %s::uuid
			  AND pc.deleted_at IS NULL
			RETURNING *
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				pc.id::text AS id,
				COALESCE(c.trade_name, u.full_name, 'Usuario') AS company,
				u.full_name AS "userName",
				pc.content AS text,
				true AS "canEdit",
				true AS "canDelete",
				pc.created_at AS "createdAt",
				pc.updated_at AS "updatedAt"
			FROM updated_comment pc
			JOIN users u ON u.id = pc.user_id
			LEFT JOIN companies c ON c.id = pc.company_id
		) item;
	`, sqlQuote(req.Content), sqlQuote(commentID), sqlQuote(postID), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel editar comentario")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "voce so pode editar seus proprios comentarios")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) deletePostComment(w http.ResponseWriter, r *http.Request, postID string, commentID string) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}

	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH deleted_comment AS (
			UPDATE post_comments pc
			SET deleted_at = now(), updated_at = now()
			FROM posts p
			WHERE pc.id = %s::uuid
			  AND pc.post_id = p.id
			  AND p.id = %s::uuid
			  AND pc.deleted_at IS NULL
			  AND (
				pc.user_id = %s::uuid
				OR p.company_id = NULLIF(%s, '')::uuid
			  )
			RETURNING pc.id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT id::text AS id, true AS deleted
			FROM deleted_comment
		) item;
	`, sqlQuote(commentID), sqlQuote(postID), sqlQuote(session.UserID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel excluir comentario")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "comentario nao encontrado ou sem permissao")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
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
	bootstrapAdminPasswordSQL := "NULL"
	if value := strings.TrimSpace(getenv("LICITAHUB_BOOTSTRAP_ADMIN_PASSWORD", "")); value != "" {
		bootstrapAdminPasswordSQL = sqlQuote(hashPassword(value))
	}

	_, err := a.runPSQL(ctx, fmt.Sprintf(`
		ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash text;
		CREATE TABLE IF NOT EXISTS auth_sessions (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash varchar(128) NOT NULL UNIQUE,
			expires_at timestamptz NOT NULL,
			revoked_at timestamptz,
			created_at timestamptz NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS idx_auth_sessions_token
			ON auth_sessions(token_hash, expires_at)
			WHERE revoked_at IS NULL;
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash varchar(128) NOT NULL UNIQUE,
			expires_at timestamptz NOT NULL,
			used_at timestamptz,
			created_at timestamptz NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens
			ON password_reset_tokens(token_hash, expires_at)
			WHERE used_at IS NULL;
		ALTER TABLE company_reviews DROP CONSTRAINT IF EXISTS company_reviews_status_chk;
		ALTER TABLE company_reviews ADD CONSTRAINT company_reviews_status_chk
			CHECK (status IN ('approved', 'adjustment_requested', 'resubmitted', 'rejected'));
		INSERT INTO users (
			access_profile_id, full_name, email, phone, job_title, password_hash, status
		)
		SELECT
			ap.id,
			'Administrador LicitaHub',
			'admin@licitahub.local',
			'',
			'Administrador da plataforma',
			%s,
			'active'
		FROM access_profiles ap
		WHERE ap.key = 'platform_admin'
		  AND NOT EXISTS (
			SELECT 1 FROM users u
			JOIN access_profiles existing_ap ON existing_ap.id = u.access_profile_id
			WHERE existing_ap.key = 'platform_admin'
			  AND u.deleted_at IS NULL
		  );
		ALTER TABLE news ADD COLUMN IF NOT EXISTS expires_at timestamptz;
		ALTER TABLE news ADD COLUMN IF NOT EXISTS archived_at timestamptz;
		ALTER TABLE news DROP CONSTRAINT IF EXISTS news_status_chk;
		ALTER TABLE news ADD CONSTRAINT news_status_chk CHECK (status IN ('draft', 'published', 'featured', 'archived', 'expired'));
		CREATE INDEX IF NOT EXISTS idx_news_status_expires_published_at ON news(status, expires_at, published_at DESC);
		ALTER TABLE company_profiles ADD COLUMN IF NOT EXISTS company_size varchar(40);
		ALTER TABLE company_profiles ADD COLUMN IF NOT EXISTS institutional_description text;
		ALTER TABLE company_profiles ADD COLUMN IF NOT EXISTS logo_media_id uuid;
		ALTER TABLE company_profiles ADD COLUMN IF NOT EXISTS state varchar(2);
		ALTER TABLE company_profiles ADD COLUMN IF NOT EXISTS city varchar(120);
		ALTER TABLE company_profiles ADD COLUMN IF NOT EXISTS national_coverage boolean NOT NULL DEFAULT false;
		ALTER TABLE company_profiles ADD COLUMN IF NOT EXISTS public_profile_slug varchar(160);
		ALTER TABLE company_profiles ADD COLUMN IF NOT EXISTS is_public boolean NOT NULL DEFAULT true;
		CREATE UNIQUE INDEX IF NOT EXISTS company_profiles_public_profile_slug_key
			ON company_profiles(public_profile_slug)
			WHERE public_profile_slug IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_companies_status_deleted_at
			ON companies(status, deleted_at);
		CREATE INDEX IF NOT EXISTS idx_company_invitations_status_created_at
			ON company_invitations(status, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_company_profiles_public
			ON company_profiles(is_public, public_profile_slug);
		CREATE INDEX IF NOT EXISTS idx_users_company_status
			ON users(company_id, status)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_media_files_uploaded_by
			ON media_files(uploaded_by_user_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_posts_feed
			ON posts(status, visibility, published_at DESC, created_at DESC)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_posts_company_profile
			ON posts(company_id, status, visibility, published_at DESC, created_at DESC)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_post_comments_active
			ON post_comments(post_id, created_at)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_post_likes_post_created_at
			ON post_likes(post_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_post_favorites_user_created_at
			ON post_favorites(user_id, created_at DESC);
		ALTER TABLE tenders DROP CONSTRAINT IF EXISTS tenders_status_chk;
		ALTER TABLE tenders ADD CONSTRAINT tenders_status_chk CHECK (status IN ('draft', 'published', 'under_review', 'suspended', 'challenged', 'occurred', 'closed', 'cancelled'));
		UPDATE tenders
		SET status = 'occurred', updated_at = now()
		WHERE status IN ('published', 'under_review', 'suspended', 'challenged')
		  AND opening_date IS NOT NULL
		  AND opening_date < now()
		  AND deleted_at IS NULL;
		UPDATE tender_requirement_types
		SET name = 'Requisito operacional',
			description = 'Acervo, atestados, experiencia da empresa e pontuacao operacional exigidos pelo edital.'
		WHERE key = 'operational_qualification';
		UPDATE tender_requirement_types
		SET name = 'Requisitos profissionais',
			description = 'Equipe, responsaveis tecnicos, curriculos, CATs, disponibilidade e pontuacao profissional.'
		WHERE key = 'professional_qualification';
		CREATE UNIQUE INDEX IF NOT EXISTS partnership_ads_active_interest_uk
			ON partnership_ads(tender_interest_id)
			WHERE tender_interest_id IS NOT NULL AND deleted_at IS NULL;
		ALTER TABLE partnership_ads ADD COLUMN IF NOT EXISTS consortium_intention_id uuid;
		ALTER TABLE partnership_ads ADD COLUMN IF NOT EXISTS ad_type varchar(40) NOT NULL DEFAULT 'company';
		ALTER TABLE partnership_ads DROP CONSTRAINT IF EXISTS partnership_ads_type_chk;
		ALTER TABLE partnership_ads ADD CONSTRAINT partnership_ads_type_chk CHECK (ad_type IN ('company', 'consortium'));
		CREATE INDEX IF NOT EXISTS idx_partnership_ads_consortium
			ON partnership_ads(consortium_intention_id, status)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tender_interests_tender_status
			ON tender_interests(tender_id, status, visibility)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_partnership_ads_showcase
			ON partnership_ads(status, published_at DESC)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_matches_company_a_status
			ON matches(company_a_id, status, matched_at DESC);
		CREATE INDEX IF NOT EXISTS idx_matches_company_b_status
			ON matches(company_b_id, status, matched_at DESC);
		CREATE INDEX IF NOT EXISTS idx_match_contacts_match_company
			ON match_contacts(match_id, company_id);
		CREATE TABLE IF NOT EXISTS chat_threads (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
			partnership_ad_id uuid NOT NULL REFERENCES partnership_ads(id) ON DELETE CASCADE,
			company_a_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
			company_b_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
			created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
			status varchar(40) NOT NULL DEFAULT 'open',
			last_message_at timestamptz,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now(),
			deleted_at timestamptz,
			CONSTRAINT chat_threads_not_same_company_chk CHECK (company_a_id <> company_b_id),
			CONSTRAINT chat_threads_status_chk CHECK (status IN ('open', 'archived', 'converted_to_match')),
			CONSTRAINT chat_threads_unique_uk UNIQUE (partnership_ad_id, company_a_id, company_b_id)
		);
		CREATE TABLE IF NOT EXISTS chat_messages (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			thread_id uuid NOT NULL REFERENCES chat_threads(id) ON DELETE CASCADE,
			sender_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
			sender_company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
			content text NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now(),
			deleted_at timestamptz
		);
		CREATE TABLE IF NOT EXISTS chat_thread_reads (
			thread_id uuid NOT NULL REFERENCES chat_threads(id) ON DELETE CASCADE,
			user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			last_read_at timestamptz NOT NULL DEFAULT now(),
			PRIMARY KEY (thread_id, user_id)
		);
		CREATE INDEX IF NOT EXISTS idx_chat_threads_company_a
			ON chat_threads(company_a_id, updated_at DESC)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_chat_threads_company_b
			ON chat_threads(company_b_id, updated_at DESC)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_chat_messages_thread_created
			ON chat_messages(thread_id, created_at)
			WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_consortium_intentions_match
			ON consortium_intentions(match_id);
		CREATE INDEX IF NOT EXISTS idx_consortium_intentions_tender
			ON consortium_intentions(tender_id);
		ALTER TABLE consortium_members ADD COLUMN IF NOT EXISTS status varchar(40) NOT NULL DEFAULT 'active';
		ALTER TABLE consortium_members ADD COLUMN IF NOT EXISTS withdrawn_at timestamptz;
		ALTER TABLE consortium_members ADD COLUMN IF NOT EXISTS withdrawn_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL;
		ALTER TABLE consortium_members DROP CONSTRAINT IF EXISTS consortium_members_status_chk;
		ALTER TABLE consortium_members ADD CONSTRAINT consortium_members_status_chk CHECK (status IN ('active', 'withdrawn'));
		CREATE INDEX IF NOT EXISTS idx_consortium_members_active
			ON consortium_members(consortium_intention_id, company_id)
			WHERE status = 'active';
		CREATE TABLE IF NOT EXISTS consortium_applications (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			consortium_intention_id uuid NOT NULL REFERENCES consortium_intentions(id) ON DELETE CASCADE,
			partnership_ad_id uuid NOT NULL REFERENCES partnership_ads(id) ON DELETE CASCADE,
			applicant_company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
			applicant_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
			leader_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
			status varchar(40) NOT NULL DEFAULT 'interested',
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now(),
			CONSTRAINT consortium_applications_status_chk CHECK (status IN ('interested', 'accepted', 'rejected', 'withdrawn')),
			CONSTRAINT consortium_applications_uk UNIQUE (consortium_intention_id, applicant_company_id)
		);
		CREATE INDEX IF NOT EXISTS idx_consortium_applications_intention
			ON consortium_applications(consortium_intention_id, status);
		CREATE INDEX IF NOT EXISTS idx_consortium_applications_ad
			ON consortium_applications(partnership_ad_id, status);
		CREATE INDEX IF NOT EXISTS idx_notifications_related
			ON notifications(related_entity_type, related_entity_id, created_at DESC);
		UPDATE partnership_ads pa
		SET status = 'closed', updated_at = now()
		FROM matches m
		WHERE m.status = 'active'
		  AND pa.tender_id = m.tender_id
		  AND pa.company_id IN (m.company_a_id, m.company_b_id)
		  AND pa.deleted_at IS NULL
		  AND COALESCE(pa.ad_type, 'company') <> 'consortium'
		  AND pa.status = 'published';
		UPDATE chat_threads ct
		SET
			status = CASE
				WHEN ct.company_a_id IN (m.company_a_id, m.company_b_id)
				 AND ct.company_b_id IN (m.company_a_id, m.company_b_id)
				THEN 'converted_to_match'
				ELSE 'archived'
			END,
			updated_at = now()
		FROM matches m
		WHERE m.status = 'active'
		  AND ct.tender_id = m.tender_id
		  AND ct.deleted_at IS NULL
		  AND ct.status = 'open'
		  AND (ct.company_a_id IN (m.company_a_id, m.company_b_id) OR ct.company_b_id IN (m.company_a_id, m.company_b_id));
		WITH reciprocal_candidates AS (
			SELECT ca.id, ca.consortium_intention_id, ca.applicant_company_id
			FROM consortium_applications ca
			JOIN consortium_intentions ci ON ci.id = ca.consortium_intention_id
			WHERE ca.status = 'interested'
			  AND EXISTS (
				SELECT 1
				FROM partner_evaluations pe
				WHERE pe.tender_id = ci.tender_id
				  AND pe.evaluator_company_id = ci.lead_company_id
				  AND pe.evaluated_company_id = ca.applicant_company_id
				  AND pe.decision = 'liked'
			  )
		), accepted_applications AS (
			UPDATE consortium_applications ca
			SET status = 'accepted', updated_at = now()
			FROM reciprocal_candidates rc
			WHERE ca.id = rc.id
			RETURNING ca.consortium_intention_id, ca.applicant_company_id
		), closed_consortium_ads AS (
			UPDATE partnership_ads pa
			SET status = 'closed', updated_at = now()
			FROM accepted_applications aa
			JOIN consortium_intentions ci ON ci.id = aa.consortium_intention_id
			WHERE pa.tender_id = ci.tender_id
			  AND (
				pa.company_id = aa.applicant_company_id
				OR EXISTS (
					SELECT 1 FROM consortium_members cm
					WHERE cm.consortium_intention_id = aa.consortium_intention_id
					  AND cm.company_id = pa.company_id
				)
			  )
			  AND pa.status = 'published'
			RETURNING pa.id
		)
		INSERT INTO consortium_members (
			consortium_intention_id, company_id, role, responsibility_description
		)
		SELECT
			consortium_intention_id,
			applicant_company_id,
			'Nova consorciada',
			'Empresa incluida automaticamente apos aceite reciproco entre candidata e lider.'
		FROM accepted_applications
		ON CONFLICT (consortium_intention_id, company_id) DO NOTHING;
		UPDATE partnership_ads pa
		SET status = 'closed', updated_at = now()
		FROM consortium_intentions ci
		WHERE pa.tender_id = ci.tender_id
		  AND pa.status = 'published'
		  AND (
			pa.consortium_intention_id = ci.id
			OR EXISTS (
				SELECT 1 FROM consortium_members cm
				WHERE cm.consortium_intention_id = ci.id
				  AND cm.company_id = pa.company_id
			)
		  )
		  AND (
			SELECT count(*) FROM consortium_members cm_count
			WHERE cm_count.consortium_intention_id = ci.id
		  ) >= 3;
		UPDATE users u
		SET status = 'pending_invite'
		FROM companies c
		WHERE u.company_id = c.id
		  AND c.status = 'pending_review'
		  AND u.status = 'active'
		  AND u.deleted_at IS NULL;
	`, bootstrapAdminPasswordSQL))
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
		"-q",
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
		inserted_photo AS (
			INSERT INTO media_files (
				company_id, media_type, file_name, file_url, mime_type, source
			)
			SELECT
				c.id,
				'profile_photo',
				NULLIF(%s, ''),
				%s,
				NULLIF(%s, ''),
				'upload'
			FROM inserted_company c
			WHERE %s <> ''
			RETURNING id, company_id
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
				profile_photo_media_id,
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
				(SELECT id FROM inserted_photo LIMIT 1),
				'pending_invite'
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
		),
		created_admin_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, type, title, message, destination_screen,
				related_entity_type, related_entity_id
			)
			SELECT
				admin_user.id,
				'system',
				'Empresa aguardando análise',
				c.trade_name || ' concluiu o cadastro e aguarda aprovação.',
				'company-review',
				'company',
				c.id
			FROM inserted_company c
			JOIN users admin_user ON admin_user.company_id IS NULL
			JOIN access_profiles ap ON ap.id = admin_user.access_profile_id
			WHERE ap.key = 'platform_admin'
			  AND admin_user.status = 'active'
			  AND admin_user.deleted_at IS NULL
			RETURNING id
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
			CROSS JOIN inserted_user u
			CROSS JOIN updated_invitation i
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
		sqlQuote(req.ProfilePhotoFileName),
		sqlQuote(req.ProfilePhotoURL),
		sqlQuote(req.ProfilePhotoMimeType),
		sqlQuote(req.ProfilePhotoURL),
		sqlQuote(req.AdminFullName),
		sqlQuote(req.AdminEmail),
		sqlQuote(req.AdminPhone),
		sqlQuote(req.AdminJobTitle),
		sqlQuote(hashPassword(req.Password)),
	)
}

func buildCreateCompanyUserSQL(req createCompanyUserRequest, tokenHash string) string {
	photoCTE := "selected_photo AS (SELECT NULL::uuid AS id)"
	if req.ProfilePhotoURL != "" {
		photoCTE = fmt.Sprintf(`
			selected_photo AS (
				INSERT INTO media_files (company_id, media_type, file_name, file_url, mime_type, source)
				VALUES (%s::uuid, 'profile_photo', NULLIF(%s, ''), %s, NULLIF(%s, ''), 'upload')
				RETURNING id
			)`,
			sqlQuote(req.CompanyID),
			sqlQuote(req.ProfilePhotoFileName),
			sqlQuote(req.ProfilePhotoURL),
			sqlQuote(req.ProfilePhotoMimeType),
		)
	}
	return fmt.Sprintf(`
		WITH selected_profile AS (
			SELECT id FROM access_profiles WHERE key = %s LIMIT 1
		),
		%s,
		inserted_user AS (
			INSERT INTO users (
				company_id,
				access_profile_id,
				full_name,
				email,
				phone,
				job_title,
				status,
				profile_photo_media_id
			)
			SELECT
				%s::uuid,
				p.id,
				%s,
				%s,
				%s,
				NULLIF(%s, ''),
				'pending_invite',
				(SELECT id FROM selected_photo)
			FROM selected_profile p
			RETURNING *
		),
		created_token AS (
			INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
			SELECT id, %s, now() + interval '7 days'
			FROM inserted_user
			RETURNING id
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
				COALESCE(photo.file_url, '') AS "profilePhotoUrl",
				u.status,
				EXISTS (SELECT 1 FROM created_token) AS "setupTokenCreated",
				u.created_at AS "createdAt",
				u.updated_at AS "updatedAt"
			FROM inserted_user u
			JOIN access_profiles p ON p.id = u.access_profile_id
			LEFT JOIN media_files photo ON photo.id = u.profile_photo_media_id
		) item;
	`,
		sqlQuote(req.AccessProfileKey),
		photoCTE,
		sqlQuote(req.CompanyID),
		sqlQuote(req.FullName),
		sqlQuote(req.Email),
		sqlQuote(req.Phone),
		sqlQuote(req.JobTitle),
		sqlQuote(tokenHash),
	)
}

func buildUpdateCompanyUserSQL(userID string, session sessionUser, req updateCompanyUserRequest) string {
	companyFilter := ""
	if session.RoleKey != "platform_admin" {
		companyFilter = fmt.Sprintf("AND u.company_id = %s::uuid", sqlQuote(session.CompanyID))
	}
	photoCTE := "selected_photo AS (SELECT NULL::uuid AS id)"
	if req.ProfilePhotoURL != "" {
		companyForPhoto := "NULL"
		if session.CompanyID != "" {
			companyForPhoto = sqlQuote(session.CompanyID) + "::uuid"
		}
		photoCTE = fmt.Sprintf(`
			selected_photo AS (
				INSERT INTO media_files (company_id, uploaded_by_user_id, media_type, file_name, file_url, mime_type, source)
				VALUES (%s, %s::uuid, 'profile_photo', NULLIF(%s, ''), %s, NULLIF(%s, ''), 'upload')
				RETURNING id
			)`,
			companyForPhoto,
			sqlQuote(session.UserID),
			sqlQuote(req.ProfilePhotoFileName),
			sqlQuote(req.ProfilePhotoURL),
			sqlQuote(req.ProfilePhotoMimeType),
		)
	}
	return fmt.Sprintf(`
		WITH selected_profile AS (
			SELECT id FROM access_profiles WHERE key = %s LIMIT 1
		),
		%s,
		updated_user AS (
			UPDATE users u
			SET
				access_profile_id = p.id,
				full_name = %s,
				email = %s,
				phone = %s,
				job_title = NULLIF(%s, ''),
				profile_photo_media_id = COALESCE((SELECT id FROM selected_photo), u.profile_photo_media_id),
				status = %s,
				blocked_at = CASE WHEN %s = 'blocked' THEN COALESCE(u.blocked_at, now()) ELSE NULL END,
				removed_at = CASE WHEN %s = 'removed' THEN COALESCE(u.removed_at, now()) ELSE NULL END
			FROM selected_profile p
			WHERE u.id = %s::uuid
			  %s
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
				COALESCE(photo.file_url, '') AS "profilePhotoUrl",
				u.status,
				u.created_at AS "createdAt",
				u.updated_at AS "updatedAt"
			FROM updated_user u
			JOIN access_profiles p ON p.id = u.access_profile_id
			LEFT JOIN media_files photo ON photo.id = u.profile_photo_media_id
		) item;
	`,
		sqlQuote(req.AccessProfileKey),
		photoCTE,
		sqlQuote(req.FullName),
		sqlQuote(req.Email),
		sqlQuote(req.Phone),
		sqlQuote(req.JobTitle),
		sqlQuote(req.Status),
		sqlQuote(req.Status),
		sqlQuote(req.Status),
		sqlQuote(userID),
		companyFilter,
	)
}

func companyProfileSQL() string {
	return `
		SELECT row_to_json(item)
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
				COALESCE(p.company_size, '') AS "companySize",
				COALESCE(p.institutional_description, '') AS "institutionalDescription",
				COALESCE(p.national_coverage, false) AS "nationalCoverage",
				COALESCE(p.public_profile_slug, '') AS "publicProfileSlug",
				COALESCE(logo.file_url, '') AS "logoUrl",
				c.created_at AS "createdAt",
				c.updated_at AS "updatedAt"
			FROM companies c
			LEFT JOIN company_profiles p ON p.company_id = c.id
			LEFT JOIN media_files logo ON logo.id = p.logo_media_id
			WHERE c.id = %s::uuid
			  AND c.deleted_at IS NULL
			LIMIT 1
		) item;
	`
}

func companyPublicProfileSQL() string {
	return `
		SELECT row_to_json(item)
		FROM (
			SELECT
				c.id::text AS id,
				c.trade_name AS "tradeName",
				c.status,
				COALESCE(c.state, '') AS state,
				COALESCE(c.city, '') AS city,
				COALESCE(p.website, '') AS website,
				COALESCE(p.company_size, '') AS "companySize",
				COALESCE(p.institutional_description, '') AS "institutionalDescription",
				COALESCE(p.national_coverage, false) AS "nationalCoverage",
				COALESCE(logo.file_url, '') AS "logoUrl",
				COALESCE((
					SELECT json_agg(row_to_json(professional) ORDER BY professional."fullName")
					FROM (
						SELECT
							u.id::text AS id,
							u.full_name AS "fullName",
							COALESCE(u.job_title, '') AS "jobTitle",
							COALESCE(u.email, '') AS email,
							COALESCE(u.phone, '') AS phone,
							COALESCE(photo.file_url, '') AS "profilePhotoUrl"
						FROM users u
						LEFT JOIN media_files photo ON photo.id = u.profile_photo_media_id
						WHERE u.company_id = c.id
						  AND u.deleted_at IS NULL
						  AND u.status = 'active'
						ORDER BY u.full_name
					) professional
				), '[]'::json) AS professionals,
				COALESCE((
					SELECT json_agg(row_to_json(post_item) ORDER BY post_item."publishedAt" DESC NULLS LAST, post_item."createdAt" DESC)
					FROM (
						SELECT
							post.id::text AS id,
							COALESCE(pc.name, 'Publicacao') AS category,
							COALESCE(pc.slug, '') AS "categorySlug",
							c.trade_name AS company,
							c.id::text AS "companyId",
							COALESCE(p.state, c.state, 'BR') AS region,
							COALESCE(logo.file_url, '') AS "companyLogoUrl",
							COALESCE(post.title, '') AS title,
							post.content AS text,
							COALESCE(media.file_url, '') AS "imageUrl",
							post.visibility,
							post.status,
							post.published_at AS "publishedAt",
							post.created_at AS "createdAt",
							(SELECT count(*) FROM post_likes pl WHERE pl.post_id = post.id)::int AS "likeCount",
							(SELECT count(*) FROM post_comments pct WHERE pct.post_id = post.id AND pct.deleted_at IS NULL)::int AS "commentCount",
							COALESCE((
								SELECT json_agg(name ORDER BY liked_at DESC)
								FROM (
									SELECT COALESCE(like_company.trade_name, like_user.full_name, 'Usuario') AS name, pl.created_at AS liked_at
									FROM post_likes pl
									JOIN users like_user ON like_user.id = pl.user_id
									LEFT JOIN companies like_company ON like_company.id = pl.company_id
									WHERE pl.post_id = post.id
									ORDER BY pl.created_at DESC
									LIMIT 20
								) liked_names
							), '[]'::json) AS likes,
							COALESCE((
								SELECT json_agg(row_to_json(comment_item) ORDER BY comment_item."createdAt")
								FROM (
									SELECT
										pct.id::text AS id,
										COALESCE(comment_company.trade_name, comment_user.full_name, 'Usuario') AS company,
										comment_user.full_name AS "userName",
										pct.content AS text,
										(pct.user_id = %s::uuid) AS "canEdit",
										(pct.user_id = %s::uuid OR post.company_id = NULLIF(%s, '')::uuid) AS "canDelete",
										pct.created_at AS "createdAt"
									FROM post_comments pct
									JOIN users comment_user ON comment_user.id = pct.user_id
									LEFT JOIN companies comment_company ON comment_company.id = pct.company_id
									WHERE pct.post_id = post.id
									  AND pct.deleted_at IS NULL
									ORDER BY pct.created_at ASC
								) comment_item
							), '[]'::json) AS comments
						FROM posts post
						LEFT JOIN post_categories pc ON pc.id = post.category_id
						LEFT JOIN media_files media ON media.id = post.main_image_media_id
						WHERE post.company_id = c.id
						  AND post.deleted_at IS NULL
						  AND post.status = 'published'
						  AND post.visibility IN ('profile', 'both')
						ORDER BY post.published_at DESC NULLS LAST, post.created_at DESC
					) post_item
				), '[]'::json) AS posts
			FROM companies c
			LEFT JOIN company_profiles p ON p.company_id = c.id
			LEFT JOIN media_files logo ON logo.id = p.logo_media_id
			WHERE c.id = %s::uuid
			  AND c.deleted_at IS NULL
			LIMIT 1
		) item;
	`
}

func buildUpdateCompanyProfileSQL(companyID string, userID string, req updateCompanyProfileRequest, logoURL string) string {
	logoMediaCTE := "selected_logo AS (SELECT NULL::uuid AS id)"
	if logoURL != "" {
		logoMediaCTE = fmt.Sprintf(`
			selected_logo AS (
				INSERT INTO media_files (
					company_id, uploaded_by_user_id, media_type, file_name, file_url, mime_type, source
				)
				VALUES (
					%s::uuid, %s::uuid, 'company_logo', NULLIF(%s, ''), %s, NULLIF(%s, ''), 'upload'
				)
				RETURNING id
			)`,
			sqlQuote(companyID),
			sqlQuote(userID),
			sqlQuote(req.LogoFileName),
			sqlQuote(logoURL),
			sqlQuote(req.LogoMimeType),
		)
	}

	return fmt.Sprintf(`
		WITH
		%s,
		updated_company AS (
			UPDATE companies
			SET
				state = NULLIF(%s, ''),
				city = NULLIF(%s, ''),
				updated_at = now()
			WHERE id = %s::uuid
			  AND deleted_at IS NULL
			RETURNING id
		),
		upsert_profile AS (
			INSERT INTO company_profiles (
				company_id, website, company_size, institutional_description,
				state, city, national_coverage, logo_media_id
			)
			SELECT
				id, NULLIF(%s, ''), NULLIF(%s, ''), NULLIF(%s, ''),
				NULLIF(%s, ''), NULLIF(%s, ''), %t,
				(SELECT id FROM selected_logo)
			FROM updated_company
			ON CONFLICT (company_id) DO UPDATE SET
				website = EXCLUDED.website,
				company_size = EXCLUDED.company_size,
				institutional_description = EXCLUDED.institutional_description,
				state = EXCLUDED.state,
				city = EXCLUDED.city,
				national_coverage = EXCLUDED.national_coverage,
				logo_media_id = COALESCE(EXCLUDED.logo_media_id, company_profiles.logo_media_id),
				updated_at = now()
			RETURNING company_id
		)
		%s
	`,
		logoMediaCTE,
		sqlQuote(req.State),
		sqlQuote(req.City),
		sqlQuote(companyID),
		sqlQuote(req.Website),
		sqlQuote(req.CompanySize),
		sqlQuote(req.InstitutionalDescription),
		sqlQuote(req.State),
		sqlQuote(req.City),
		req.NationalCoverage,
		fmt.Sprintf(companyProfileSQL(), sqlQuote(companyID)),
	)
}

func buildCreateTenderInterestSQL(tenderID string, session sessionUser, req createTenderInterestRequest, requirementsJSON string) string {
	return fmt.Sprintf(`
		WITH selected_tender AS (
			SELECT id, number, object FROM tenders
			WHERE id = %s::uuid AND deleted_at IS NULL AND status IN ('published', 'under_review')
			LIMIT 1
		),
		ensured_requirements AS (
			INSERT INTO tender_requirements (tender_id, requirement_type_id, description, order_index)
			SELECT
				st.id,
				rt.id,
				rt.description,
				CASE rt.key
					WHEN 'operational_qualification' THEN 1
					WHEN 'professional_qualification' THEN 2
					WHEN 'technical_proposal' THEN 3
					WHEN 'certifications' THEN 4
					ELSE 9
				END
			FROM selected_tender st
			JOIN tender_requirement_types rt ON rt.key IN (
				'operational_qualification', 'professional_qualification', 'technical_proposal', 'certifications'
			)
			ON CONFLICT (tender_id, requirement_type_id) DO NOTHING
			RETURNING id
		),
		updated_interest AS (
			UPDATE tender_interests
			SET
				general_position = %s,
				desired_role = %s,
				public_summary = %s,
				internal_note = NULLIF(%s, ''),
				visibility = 'public_showcase',
				status = 'published',
				updated_at = now()
			WHERE tender_id = (SELECT id FROM selected_tender)
			  AND company_id = %s::uuid
			  AND deleted_at IS NULL
			  AND status <> 'withdrawn'
			RETURNING *
		),
		inserted_interest AS (
			INSERT INTO tender_interests (
				tender_id, company_id, created_by_user_id, general_position,
				desired_role, public_summary, internal_note, visibility, status
			)
			SELECT
				st.id,
				%s::uuid,
				%s::uuid,
				%s,
				%s,
				%s,
				NULLIF(%s, ''),
				'public_showcase',
				'published'
			FROM selected_tender st
			WHERE NOT EXISTS (SELECT 1 FROM updated_interest)
			RETURNING *
		),
		selected_interest AS (
			SELECT * FROM updated_interest
			UNION ALL
			SELECT * FROM inserted_interest
			LIMIT 1
		),
		clear_requirements AS (
			DELETE FROM tender_interest_requirements
			WHERE tender_interest_id = (SELECT id FROM selected_interest)
			RETURNING id
		),
		incoming AS (
			SELECT *
			FROM jsonb_to_recordset(%s::jsonb) AS x(
				"requirementKey" text,
				"statusKey" text,
				"whatWeHave" text,
				"whatWeSeek" text
			)
		),
		inserted_requirements AS (
			INSERT INTO tender_interest_requirements (
				tender_interest_id, tender_requirement_id, status_key, what_we_have, what_we_seek
			)
			SELECT
				si.id,
				tr.id,
				COALESCE(NULLIF(i."statusKey", ''), 'under_review'),
				NULLIF(i."whatWeHave", ''),
				NULLIF(i."whatWeSeek", '')
			FROM selected_interest si
			JOIN incoming i ON true
			JOIN tender_requirement_types rt ON rt.key = i."requirementKey"
			JOIN tender_requirements tr ON tr.tender_id = si.tender_id AND tr.requirement_type_id = rt.id
			RETURNING id
		),
		updated_ad AS (
			UPDATE partnership_ads
			SET
				title = (SELECT c.trade_name || ' disponível para parceria' FROM companies c WHERE c.id = %s::uuid),
				offer_summary = %s,
				seek_summary = COALESCE((
					SELECT string_agg(NULLIF(i."whatWeSeek", ''), ' | ')
					FROM incoming i
					WHERE NULLIF(i."whatWeSeek", '') IS NOT NULL
				), ''),
				status = 'published',
				published_at = COALESCE(published_at, now()),
				updated_at = now(),
				deleted_at = NULL
			WHERE tender_interest_id = (SELECT id FROM selected_interest)
			  AND deleted_at IS NULL
			RETURNING *
		),
		inserted_ad AS (
			INSERT INTO partnership_ads (
				tender_id, company_id, tender_interest_id, title,
				offer_summary, seek_summary, status, published_at
			)
			SELECT
				si.tender_id,
				si.company_id,
				si.id,
				c.trade_name || ' disponível para parceria',
				%s,
				COALESCE((
					SELECT string_agg(NULLIF(i."whatWeSeek", ''), ' | ')
					FROM incoming i
					WHERE NULLIF(i."whatWeSeek", '') IS NOT NULL
				), ''),
				'published',
				now()
			FROM selected_interest si
			JOIN companies c ON c.id = si.company_id
			WHERE NOT EXISTS (SELECT 1 FROM updated_ad)
			RETURNING *
		),
		selected_ad AS (
			SELECT * FROM updated_ad
			UNION ALL
			SELECT * FROM inserted_ad
			LIMIT 1
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT DISTINCT
				u.id,
				u.company_id,
				'company_interested',
				'Nova empresa interessada',
				c_new.trade_name || ' registrou interesse no mesmo edital.',
				'tender-interest-list',
				'tender',
				si.tender_id
			FROM selected_interest si
			JOIN companies c_new ON c_new.id = si.company_id
			JOIN tender_interests other_interest ON other_interest.tender_id = si.tender_id
				AND other_interest.company_id <> si.company_id
				AND other_interest.deleted_at IS NULL
				AND other_interest.status = 'published'
			JOIN users u ON u.company_id = other_interest.company_id
			WHERE u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				si.id::text AS "interestId",
				sa.id::text AS "adId",
				si.tender_id::text AS "tenderId",
				si.company_id::text AS "companyId",
				si.public_summary AS "publicSummary",
				sa.status AS "adStatus"
			FROM selected_interest si
			JOIN selected_ad sa ON sa.tender_interest_id = si.id
		) item;
	`,
		sqlQuote(tenderID),
		sqlQuote(req.GeneralPosition),
		sqlQuote(req.DesiredRole),
		sqlQuote(req.PublicSummary),
		sqlQuote(req.InternalNote),
		sqlQuote(session.CompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.UserID),
		sqlQuote(req.GeneralPosition),
		sqlQuote(req.DesiredRole),
		sqlQuote(req.PublicSummary),
		sqlQuote(req.InternalNote),
		sqlQuote(requirementsJSON),
		sqlQuote(session.CompanyID),
		sqlQuote(req.PublicSummary),
		sqlQuote(req.PublicSummary),
	)
}

func partnershipAdsSQL(extraWhere string, session sessionUser) string {
	return fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."publishedAt" DESC), '[]'::json)
		FROM (
			SELECT
				pa.id::text AS id,
				pa.tender_id::text AS "tenderId",
				pa.company_id::text AS "companyId",
				pa.tender_interest_id::text AS "interestId",
				COALESCE(pa.ad_type, 'company') AS "adType",
				COALESCE(pa.consortium_intention_id::text, '') AS "consortiumIntentionId",
				pa.title,
				COALESCE(pa.offer_summary, '') AS "offerSummary",
				COALESCE(pa.seek_summary, '') AS "seekSummary",
				pa.status,
				pa.published_at AS "publishedAt",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				COALESCE(t.judgment_criterion, '') AS "judgmentCriterion",
				COALESCE(t.state, '') AS state,
				COALESCE(t.city, '') AS city,
				c.trade_name AS "companyName",
				CASE WHEN COALESCE(pa.ad_type, 'company') = 'consortium' THEN c.trade_name ELSE '' END AS "leaderCompanyName",
				COALESCE(c.main_contact_phone, '') AS "companyPhone",
				COALESCE(logo.file_url, '') AS "companyLogoUrl",
				CASE
					WHEN COALESCE(pa.ad_type, 'company') = 'consortium' AND EXISTS (
						SELECT 1 FROM consortium_applications ca
						WHERE ca.partnership_ad_id = pa.id
						  AND ca.applicant_company_id = %s::uuid
						  AND ca.status IN ('interested', 'accepted')
					) THEN 'liked'
					ELSE COALESCE((
						SELECT pe.decision
						FROM partner_evaluations pe
						WHERE pe.tender_id = pa.tender_id
						  AND pe.evaluator_company_id = %s::uuid
						  AND pe.evaluated_company_id = pa.company_id
						ORDER BY pe.updated_at DESC
						LIMIT 1
					), '')
				END AS "currentEvaluationDecision",
				COALESCE((
					SELECT json_agg(row_to_json(member) ORDER BY member."companyName")
					FROM (
						SELECT
							cm.company_id::text AS "companyId",
							member_company.trade_name AS "companyName",
							COALESCE(cm.role, '') AS role
						FROM consortium_members cm
						JOIN companies member_company ON member_company.id = cm.company_id
						WHERE cm.consortium_intention_id = pa.consortium_intention_id
					) member
				), '[]'::json) AS "consortiumMembers",
				COALESCE((
					SELECT json_agg(row_to_json(req) ORDER BY req."orderIndex")
					FROM (
						SELECT
							rt.key AS "requirementKey",
							rt.name,
							tr.order_index AS "orderIndex",
							tir.status_key AS "statusKey",
							COALESCE(tir.what_we_have, '') AS "whatWeHave",
							COALESCE(tir.what_we_seek, '') AS "whatWeSeek"
						FROM tender_interest_requirements tir
						JOIN tender_requirements tr ON tr.id = tir.tender_requirement_id
						JOIN tender_requirement_types rt ON rt.id = tr.requirement_type_id
						WHERE tir.tender_interest_id = pa.tender_interest_id
					) req
				), '[]'::json) AS requirements
			FROM partnership_ads pa
			JOIN tenders t ON t.id = pa.tender_id
			JOIN companies c ON c.id = pa.company_id
			LEFT JOIN company_profiles cp ON cp.company_id = c.id
			LEFT JOIN media_files logo ON logo.id = cp.logo_media_id
			WHERE pa.deleted_at IS NULL
			  AND pa.status = 'published'
			  AND t.deleted_at IS NULL
			  %s
		) item;
	`, sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), extraWhere)
}

func partnershipAdDetailSQL(adID string) string {
	return fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				pa.id::text AS id,
				pa.tender_id::text AS "tenderId",
				pa.company_id::text AS "companyId",
				pa.tender_interest_id::text AS "interestId",
				COALESCE(pa.ad_type, 'company') AS "adType",
				COALESCE(pa.consortium_intention_id::text, '') AS "consortiumIntentionId",
				pa.title,
				COALESCE(pa.offer_summary, '') AS "offerSummary",
				COALESCE(pa.seek_summary, '') AS "seekSummary",
				pa.status,
				pa.published_at AS "publishedAt",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				COALESCE(t.judgment_criterion, '') AS "judgmentCriterion",
				COALESCE(t.state, '') AS state,
				COALESCE(t.city, '') AS city,
				c.trade_name AS "companyName",
				CASE WHEN COALESCE(pa.ad_type, 'company') = 'consortium' THEN c.trade_name ELSE '' END AS "leaderCompanyName",
				COALESCE(c.main_contact_phone, '') AS "companyPhone",
				COALESCE(logo.file_url, '') AS "companyLogoUrl",
				COALESCE((
					SELECT json_agg(row_to_json(member) ORDER BY member."companyName")
					FROM (
						SELECT
							cm.company_id::text AS "companyId",
							member_company.trade_name AS "companyName",
							COALESCE(cm.role, '') AS role
						FROM consortium_members cm
						JOIN companies member_company ON member_company.id = cm.company_id
						WHERE cm.consortium_intention_id = pa.consortium_intention_id
					) member
				), '[]'::json) AS "consortiumMembers",
				COALESCE((
					SELECT json_agg(row_to_json(req) ORDER BY req."orderIndex")
					FROM (
						SELECT
							rt.key AS "requirementKey",
							rt.name,
							tr.order_index AS "orderIndex",
							tir.status_key AS "statusKey",
							COALESCE(tir.what_we_have, '') AS "whatWeHave",
							COALESCE(tir.what_we_seek, '') AS "whatWeSeek"
						FROM tender_interest_requirements tir
						JOIN tender_requirements tr ON tr.id = tir.tender_requirement_id
						JOIN tender_requirement_types rt ON rt.id = tr.requirement_type_id
						WHERE tir.tender_interest_id = pa.tender_interest_id
					) req
				), '[]'::json) AS requirements
			FROM partnership_ads pa
			JOIN tenders t ON t.id = pa.tender_id
			JOIN companies c ON c.id = pa.company_id
			LEFT JOIN company_profiles cp ON cp.company_id = c.id
			LEFT JOIN media_files logo ON logo.id = cp.logo_media_id
			WHERE pa.id = %s::uuid
			  AND pa.deleted_at IS NULL
			  AND t.deleted_at IS NULL
			LIMIT 1
		) item;
	`, sqlQuote(adID))
}

func partnershipAdKindSQL(adID string) string {
	return fmt.Sprintf(`
		SELECT row_to_json(item)
		FROM (
			SELECT
				pa.id::text AS id,
				COALESCE(pa.ad_type, 'company') AS "adType",
				pa.company_id::text AS "companyId",
				COALESCE(pa.consortium_intention_id::text, '') AS "consortiumIntentionId"
			FROM partnership_ads pa
			WHERE pa.id = %s::uuid
			  AND pa.deleted_at IS NULL
			LIMIT 1
		) item;
	`, sqlQuote(adID))
}

func chatThreadsSQL(session sessionUser, extraWhere string) string {
	return fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."lastActivityAt" DESC), '[]'::json)
		FROM (
			SELECT
				ct.id::text AS id,
				ct.tender_id::text AS "tenderId",
				ct.partnership_ad_id::text AS "adId",
				ct.company_a_id::text AS "companyAId",
				ct.company_b_id::text AS "companyBId",
				other_company.id::text AS "otherCompanyId",
				other_company.trade_name AS "otherCompanyName",
				COALESCE(other_logo.file_url, '') AS "otherCompanyLogoUrl",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				ct.status,
				CASE
					WHEN ct.status = 'archived' THEN true
					ELSE false
				END AS "isClosed",
				CASE
					WHEN ct.status = 'archived' THEN 'Conversa encerrada porque uma das empresas avancou com outro parceiro neste edital.'
					WHEN ct.status = 'converted_to_match' THEN 'Conversa mantida apos match para alinhamento do consorcio.'
					ELSE ''
				END AS "closedReason",
				CASE
					WHEN ct.status IN ('open', 'converted_to_match') THEN true
					ELSE false
				END AS "canReply",
				COALESCE((
					SELECT other_ad.id::text
					FROM partnership_ads other_ad
					WHERE other_ad.tender_id = ct.tender_id
					  AND other_ad.company_id = other_company.id
					  AND other_ad.deleted_at IS NULL
					ORDER BY CASE WHEN other_ad.status = 'published' THEN 0 ELSE 1 END, other_ad.created_at DESC
					LIMIT 1
				), ct.partnership_ad_id::text) AS "evaluationAdId",
				COALESCE(ct.last_message_at, ct.updated_at, ct.created_at) AS "lastActivityAt",
				COALESCE((
					SELECT cm.content
					FROM chat_messages cm
					WHERE cm.thread_id = ct.id
					  AND cm.deleted_at IS NULL
					ORDER BY cm.created_at DESC
					LIMIT 1
				), '') AS "lastMessage",
				(
					SELECT count(*)
					FROM chat_messages cm
					WHERE cm.thread_id = ct.id
					  AND cm.deleted_at IS NULL
					  AND cm.sender_company_id <> %s::uuid
					  AND cm.created_at > COALESCE(ctr.last_read_at, 'epoch'::timestamptz)
				)::int AS "unreadCount"
			FROM chat_threads ct
			JOIN tenders t ON t.id = ct.tender_id
			JOIN companies other_company ON other_company.id = CASE
				WHEN ct.company_a_id = %s::uuid THEN ct.company_b_id
				ELSE ct.company_a_id
			END
			LEFT JOIN company_profiles other_profile ON other_profile.company_id = other_company.id
			LEFT JOIN media_files other_logo ON other_logo.id = other_profile.logo_media_id
			LEFT JOIN chat_thread_reads ctr ON ctr.thread_id = ct.id AND ctr.user_id = %s::uuid
			WHERE ct.deleted_at IS NULL
			  AND (%s::uuid IN (ct.company_a_id, ct.company_b_id))
			  %s
		) item;
	`, sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), extraWhere)
}

func buildStartChatSQL(session sessionUser, adID string) string {
	return fmt.Sprintf(`
		WITH selected_ad AS (
			SELECT pa.id, pa.tender_id, pa.company_id
			FROM partnership_ads pa
			WHERE pa.id = %s::uuid
			  AND pa.deleted_at IS NULL
			  AND pa.status = 'published'
			  AND pa.company_id <> %s::uuid
			  AND (COALESCE(pa.ad_type, 'company') = 'consortium' OR NOT EXISTS (
				SELECT 1
				FROM matches m
				WHERE m.tender_id = pa.tender_id
				  AND m.status = 'active'
				  AND (%s::uuid IN (m.company_a_id, m.company_b_id) OR pa.company_id IN (m.company_a_id, m.company_b_id))
			  ))
			LIMIT 1
		),
		upserted AS (
			INSERT INTO chat_threads (
				tender_id, partnership_ad_id, company_a_id, company_b_id, created_by_user_id, last_message_at
			)
			SELECT
				sa.tender_id,
				sa.id,
				CASE WHEN %s::uuid::text < sa.company_id::text THEN %s::uuid ELSE sa.company_id END,
				CASE WHEN %s::uuid::text < sa.company_id::text THEN sa.company_id ELSE %s::uuid END,
				%s::uuid,
				now()
			FROM selected_ad sa
			ON CONFLICT ON CONSTRAINT chat_threads_unique_uk
			DO UPDATE SET updated_at = now()
			RETURNING *
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				'chat-thread' AS "eventType",
				ct.id::text AS id,
				ct.tender_id::text AS "tenderId",
				ct.partnership_ad_id::text AS "adId",
				ct.company_a_id::text AS "companyAId",
				ct.company_b_id::text AS "companyBId",
				other_company.id::text AS "otherCompanyId",
				other_company.trade_name AS "otherCompanyName",
				COALESCE(other_logo.file_url, '') AS "otherCompanyLogoUrl",
				t.number AS "tenderNumber",
				t.agency,
				t.object AS "tenderObject",
				ct.status,
				false AS "isClosed",
				'' AS "closedReason",
				true AS "canReply",
				ct.partnership_ad_id::text AS "evaluationAdId",
				COALESCE(ct.last_message_at, ct.updated_at, ct.created_at) AS "lastActivityAt",
				'' AS "lastMessage",
				0 AS "unreadCount"
			FROM upserted ct
			JOIN tenders t ON t.id = ct.tender_id
			JOIN companies other_company ON other_company.id = CASE
				WHEN ct.company_a_id = %s::uuid THEN ct.company_b_id
				ELSE ct.company_a_id
			END
			LEFT JOIN company_profiles other_profile ON other_profile.company_id = other_company.id
			LEFT JOIN media_files other_logo ON other_logo.id = other_profile.logo_media_id
			LIMIT 1
		) item;
	`, sqlQuote(adID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(session.CompanyID))
}

func chatMessagesSQL(session sessionUser, threadID string) string {
	return fmt.Sprintf(`
		WITH selected_thread AS (
			SELECT *
			FROM chat_threads
			WHERE id = %s::uuid
			  AND deleted_at IS NULL
			  AND %s::uuid IN (company_a_id, company_b_id)
			LIMIT 1
		),
		read_marker AS (
			INSERT INTO chat_thread_reads (thread_id, user_id, last_read_at)
			SELECT id, %s::uuid, now()
			FROM selected_thread
			ON CONFLICT (thread_id, user_id)
			DO UPDATE SET last_read_at = now()
			RETURNING *
		)
		SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."createdAt"), '[]'::json)
		FROM (
			SELECT
				cm.id::text AS id,
				cm.thread_id::text AS "threadId",
				st.company_a_id::text AS "companyAId",
				st.company_b_id::text AS "companyBId",
				st.status,
				(st.status = 'archived') AS "isClosed",
				cm.sender_user_id::text AS "senderUserId",
				cm.sender_company_id::text AS "senderCompanyId",
				COALESCE(u.full_name, 'Usuario') AS "senderName",
				COALESCE(u.job_title, '') AS "senderJobTitle",
				COALESCE(user_photo.file_url, '') AS "senderPhotoUrl",
				c.trade_name AS "senderCompanyName",
				cm.content,
				(cm.sender_user_id = %s::uuid) AS mine,
				cm.created_at AS "createdAt"
			FROM selected_thread st
			JOIN chat_messages cm ON cm.thread_id = st.id
			JOIN companies c ON c.id = cm.sender_company_id
			LEFT JOIN users u ON u.id = cm.sender_user_id
			LEFT JOIN media_files user_photo ON user_photo.id = u.profile_photo_media_id
			WHERE cm.deleted_at IS NULL
		) item;
	`, sqlQuote(threadID), sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(session.UserID))
}

func buildCreateChatMessageSQL(session sessionUser, threadID string, content string) string {
	return fmt.Sprintf(`
		WITH selected_thread AS (
			SELECT *
			FROM chat_threads
			WHERE id = %s::uuid
			  AND deleted_at IS NULL
			  AND status IN ('open', 'converted_to_match')
			  AND %s::uuid IN (company_a_id, company_b_id)
			LIMIT 1
		),
		inserted_message AS (
			INSERT INTO chat_messages (thread_id, sender_user_id, sender_company_id, content)
			SELECT id, %s::uuid, %s::uuid, %s
			FROM selected_thread
			RETURNING *
		),
		updated_thread AS (
			UPDATE chat_threads
			SET last_message_at = now(), updated_at = now()
			WHERE id = (SELECT thread_id FROM inserted_message)
			RETURNING *
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				'chat-message' AS "eventType",
				im.id::text AS id,
				im.thread_id::text AS "threadId",
				ut.company_a_id::text AS "companyAId",
				ut.company_b_id::text AS "companyBId",
				im.sender_user_id::text AS "senderUserId",
				im.sender_company_id::text AS "senderCompanyId",
				COALESCE(u.full_name, 'Usuario') AS "senderName",
				COALESCE(u.job_title, '') AS "senderJobTitle",
				COALESCE(user_photo.file_url, '') AS "senderPhotoUrl",
				c.trade_name AS "senderCompanyName",
				im.content,
				im.created_at AS "createdAt"
			FROM inserted_message im
			JOIN updated_thread ut ON ut.id = im.thread_id
			JOIN companies c ON c.id = im.sender_company_id
			LEFT JOIN users u ON u.id = im.sender_user_id
			LEFT JOIN media_files user_photo ON user_photo.id = u.profile_photo_media_id
			LIMIT 1
		) item;
	`, sqlQuote(threadID), sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(content))
}

func buildEvaluateConsortiumAdSQL(adID string, session sessionUser, decision string) string {
	status := "interested"
	if decision == "rejected" {
		status = "withdrawn"
	}
	if decision == "later" {
		status = "interested"
	}
	return fmt.Sprintf(`
		WITH selected_ad AS (
			SELECT
				pa.*, ci.lead_company_id, ci.match_id, t.number AS tender_number,
				EXISTS (
					SELECT 1
					FROM partner_evaluations pe
					WHERE pe.tender_id = pa.tender_id
					  AND pe.evaluator_company_id = ci.lead_company_id
					  AND pe.evaluated_company_id = %s::uuid
					  AND pe.decision = 'liked'
				) AS leader_liked_candidate
			FROM partnership_ads pa
			JOIN consortium_intentions ci ON ci.id = pa.consortium_intention_id
			JOIN tenders t ON t.id = pa.tender_id
			WHERE pa.id = %s::uuid
			  AND pa.deleted_at IS NULL
			  AND pa.status = 'published'
			  AND pa.ad_type = 'consortium'
			  AND pa.company_id <> %s::uuid
			  AND NOT EXISTS (
				SELECT 1
				FROM consortium_members cm
				WHERE cm.consortium_intention_id = pa.consortium_intention_id
				  AND cm.company_id = %s::uuid
			  )
			LIMIT 1
		),
		updated_application AS (
			UPDATE consortium_applications ca
			SET
				status = CASE
					WHEN %s = 'liked' AND sa.leader_liked_candidate THEN 'accepted'
					ELSE %s
				END,
				applicant_user_id = %s::uuid,
				updated_at = now()
			FROM selected_ad sa
			WHERE ca.consortium_intention_id = sa.consortium_intention_id
			  AND ca.applicant_company_id = %s::uuid
			RETURNING ca.*
		),
		inserted_application AS (
			INSERT INTO consortium_applications (
				consortium_intention_id, partnership_ad_id, applicant_company_id,
				applicant_user_id, status
			)
			SELECT
				sa.consortium_intention_id,
				sa.id,
				%s::uuid,
				%s::uuid,
				CASE
					WHEN %s = 'liked' AND sa.leader_liked_candidate THEN 'accepted'
					ELSE %s
				END
			FROM selected_ad sa
			WHERE NOT EXISTS (SELECT 1 FROM updated_application)
			RETURNING *
		),
		selected_application AS (
			SELECT * FROM updated_application
			UNION ALL
			SELECT * FROM inserted_application
			LIMIT 1
		),
		inserted_member AS (
			INSERT INTO consortium_members (
				consortium_intention_id, company_id, role, responsibility_description
			)
			SELECT
				app.consortium_intention_id,
				app.applicant_company_id,
				'Nova consorciada',
				'Empresa incluida automaticamente apos aceite reciproco entre candidata e lider.'
			FROM selected_application app
			WHERE app.status = 'accepted'
			ON CONFLICT (consortium_intention_id, company_id) DO NOTHING
			RETURNING id
		),
		closed_consortium_ad AS (
			UPDATE partnership_ads pa
			SET status = 'closed', updated_at = now()
			FROM selected_ad sa
			JOIN selected_application app ON true
			WHERE pa.tender_id = sa.tender_id
			  AND app.status = 'accepted'
			  AND (
				pa.company_id = app.applicant_company_id
				OR EXISTS (
					SELECT 1 FROM consortium_members cm
					WHERE cm.consortium_intention_id = app.consortium_intention_id
					  AND cm.company_id = pa.company_id
				)
			  )
			  AND pa.status = 'published'
			RETURNING pa.id
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				sa.lead_company_id,
				'match',
				'Nova candidata ao consorcio',
				c_app.trade_name || ' demonstrou interesse em entrar no consorcio do edital ' || sa.tender_number || '.',
				'match-list',
				'consortium_application',
				app.id
			FROM selected_application app
			JOIN selected_ad sa ON true
			JOIN companies c_app ON c_app.id = app.applicant_company_id
			JOIN users u ON u.company_id = sa.lead_company_id
			WHERE app.status = 'interested'
			  AND u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		),
		accepted_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				app.applicant_company_id,
				'match',
				'Entrada no consorcio confirmada',
				'Os dois lados deram aceite. Sua empresa agora compoe este consorcio.',
				'match-list',
				'consortium_application',
				app.id
			FROM selected_application app
			JOIN users u ON u.company_id = app.applicant_company_id
			WHERE app.status = 'accepted'
			  AND u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				app.id::text AS "applicationId",
				app.status,
				false AS "matchCreated",
				'' AS "matchId",
				true AS "applicationCreated",
				(app.status = 'accepted') AS "applicationAccepted"
			FROM selected_application app
		) item;
	`,
		sqlQuote(session.CompanyID),
		sqlQuote(adID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(decision),
		sqlQuote(status),
		sqlQuote(session.UserID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.UserID),
		sqlQuote(decision),
		sqlQuote(status),
	)
}

func buildEvaluatePartnershipAdSQL(adID string, session sessionUser, decision string) string {
	return fmt.Sprintf(`
		WITH selected_ad AS (
			SELECT pa.*
			FROM partnership_ads pa
			WHERE pa.id = %s::uuid
			  AND pa.deleted_at IS NULL
			  AND pa.status = 'published'
			  AND pa.company_id <> %s::uuid
			LIMIT 1
		),
		updated_eval AS (
			UPDATE partner_evaluations pe
			SET
				decision = %s,
				evaluated_ad_id = (SELECT id FROM selected_ad),
				created_by_user_id = %s::uuid,
				updated_at = now()
			WHERE pe.tender_id = (SELECT tender_id FROM selected_ad)
			  AND pe.evaluator_company_id = %s::uuid
			  AND pe.evaluated_company_id = (SELECT company_id FROM selected_ad)
			RETURNING *
		),
		inserted_eval AS (
			INSERT INTO partner_evaluations (
				tender_id, evaluator_company_id, evaluated_company_id,
				evaluated_ad_id, decision, created_by_user_id
			)
			SELECT
				sa.tender_id,
				%s::uuid,
				sa.company_id,
				sa.id,
				%s,
				%s::uuid
			FROM selected_ad sa
			WHERE NOT EXISTS (SELECT 1 FROM updated_eval)
			RETURNING *
		),
		selected_eval AS (
			SELECT * FROM updated_eval
			UNION ALL
			SELECT * FROM inserted_eval
			LIMIT 1
		),
		reverse_eval AS (
			SELECT pe.*
			FROM partner_evaluations pe
			JOIN selected_eval se ON pe.tender_id = se.tender_id
			WHERE pe.evaluator_company_id = se.evaluated_company_id
			  AND pe.evaluated_company_id = se.evaluator_company_id
			  AND pe.decision = 'liked'
			LIMIT 1
		),
		inserted_match AS (
			INSERT INTO matches (
				tender_id, company_a_id, company_b_id, company_a_evaluation_id, company_b_evaluation_id, status
			)
			SELECT
				se.tender_id,
				CASE WHEN se.evaluator_company_id < se.evaluated_company_id THEN se.evaluator_company_id ELSE se.evaluated_company_id END,
				CASE WHEN se.evaluator_company_id < se.evaluated_company_id THEN se.evaluated_company_id ELSE se.evaluator_company_id END,
				CASE WHEN se.evaluator_company_id < se.evaluated_company_id THEN se.id ELSE re.id END,
				CASE WHEN se.evaluator_company_id < se.evaluated_company_id THEN re.id ELSE se.id END,
				'active'
			FROM selected_eval se
			JOIN reverse_eval re ON true
			WHERE se.decision = 'liked'
			ON CONFLICT (tender_id, company_a_id, company_b_id) DO UPDATE
			SET status = 'active', updated_at = now()
			RETURNING *
		),
		accepted_consortium_application AS (
			UPDATE consortium_applications ca
			SET
				status = 'accepted',
				leader_user_id = %s::uuid,
				updated_at = now()
			FROM selected_eval se
			JOIN consortium_intentions ci
				ON ci.tender_id = se.tender_id
			JOIN matches base_match
				ON base_match.id = ci.match_id AND base_match.status = 'active'
			WHERE se.decision = 'liked'
			  AND ci.lead_company_id = se.evaluator_company_id
			  AND ca.consortium_intention_id = ci.id
			  AND ca.applicant_company_id = se.evaluated_company_id
			  AND ca.status = 'interested'
			RETURNING ca.*
		),
		inserted_consortium_member AS (
			INSERT INTO consortium_members (
				consortium_intention_id, company_id, role, responsibility_description
			)
			SELECT
				ca.consortium_intention_id,
				ca.applicant_company_id,
				'Nova consorciada',
				'Empresa incluida automaticamente apos aceite reciproco entre candidata e lider.'
			FROM accepted_consortium_application ca
			ON CONFLICT (consortium_intention_id, company_id) DO NOTHING
			RETURNING id
		),
		closed_consortium_ads AS (
			UPDATE partnership_ads pa
			SET status = 'closed', updated_at = now()
			FROM accepted_consortium_application ca
			JOIN consortium_intentions ci ON ci.id = ca.consortium_intention_id
			WHERE pa.tender_id = ci.tender_id
			  AND (
				pa.company_id = ca.applicant_company_id
				OR EXISTS (
					SELECT 1 FROM consortium_members cm
					WHERE cm.consortium_intention_id = ca.consortium_intention_id
					  AND cm.company_id = pa.company_id
				)
			  )
			  AND pa.status = 'published'
			RETURNING pa.id
		),
		closed_ads AS (
			UPDATE partnership_ads pa
			SET status = 'closed', updated_at = now()
			FROM inserted_match m
			WHERE pa.tender_id = m.tender_id
			  AND pa.company_id IN (m.company_a_id, m.company_b_id)
			  AND pa.deleted_at IS NULL
			  AND COALESCE(pa.ad_type, 'company') <> 'consortium'
			RETURNING pa.id
		),
		updated_chat_threads AS (
			UPDATE chat_threads ct
			SET
				status = CASE
					WHEN ct.company_a_id IN (m.company_a_id, m.company_b_id)
					 AND ct.company_b_id IN (m.company_a_id, m.company_b_id)
					THEN 'converted_to_match'
					ELSE 'archived'
				END,
				updated_at = now()
			FROM inserted_match m
			WHERE ct.tender_id = m.tender_id
			  AND ct.deleted_at IS NULL
			  AND ct.status = 'open'
			  AND (ct.company_a_id IN (m.company_a_id, m.company_b_id) OR ct.company_b_id IN (m.company_a_id, m.company_b_id))
			RETURNING ct.id
		),
		inserted_contacts AS (
			INSERT INTO match_contacts (match_id, company_id, contact_name, phone, whatsapp_url, message_template)
			SELECT
				m.id,
				c.id,
				COALESCE(NULLIF(c.main_contact_name, ''), c.trade_name),
				COALESCE(NULLIF(c.main_contact_phone, ''), 'Telefone não informado'),
				CASE
					WHEN regexp_replace(COALESCE(c.main_contact_phone, ''), '[^0-9]', '', 'g') <> ''
					THEN 'https://wa.me/55' || regexp_replace(c.main_contact_phone, '[^0-9]', '', 'g')
					ELSE ''
				END,
				'Contato gerado pelo match da LicitaHub.'
			FROM inserted_match m
			JOIN companies c ON c.id IN (m.company_a_id, m.company_b_id)
			ON CONFLICT DO NOTHING
			RETURNING id
		),
		liked_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				se.evaluated_company_id,
				'match',
				'Empresa aprovou seu anuncio',
				c_eval.trade_name || ' aprovou sua empresa como possivel parceira.',
				'match-profile',
				'partnership_ad',
				se.evaluated_ad_id
			FROM selected_eval se
			JOIN companies c_eval ON c_eval.id = se.evaluator_company_id
			JOIN users u ON u.company_id = se.evaluated_company_id
			WHERE se.decision = 'liked'
			  AND NOT EXISTS (SELECT 1 FROM inserted_match)
			  AND NOT EXISTS (SELECT 1 FROM accepted_consortium_application)
			  AND u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		),
		consortium_accept_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				ca.applicant_company_id,
				'match',
				'Entrada no consorcio confirmada',
				'Os dois lados deram aceite. Sua empresa agora compoe este consorcio.',
				'match-list',
				'consortium_application',
				ca.id
			FROM accepted_consortium_application ca
			JOIN users u ON u.company_id = ca.applicant_company_id
			WHERE u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		),
		match_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				u.company_id,
				'match',
				'Match e consorcio realizado',
				'Houve aceite reciproco para formacao de consorcio.',
				'match-success',
				'match',
				m.id
			FROM inserted_match m
			JOIN users u ON u.company_id IN (m.company_a_id, m.company_b_id)
			WHERE u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				se.id::text AS "evaluationId",
				se.decision,
				EXISTS (SELECT 1 FROM inserted_match) AS "matchCreated",
				(SELECT id::text FROM inserted_match LIMIT 1) AS "matchId",
				EXISTS (SELECT 1 FROM accepted_consortium_application) AS "applicationAccepted"
			FROM selected_eval se
		) item;
	`,
		sqlQuote(adID),
		sqlQuote(session.CompanyID),
		sqlQuote(decision),
		sqlQuote(session.UserID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(decision),
		sqlQuote(session.UserID),
		sqlQuote(session.UserID),
	)
}

func buildUpdateConsortiumLeaderSQL(matchID string, session sessionUser, req updateConsortiumLeaderRequest) string {
	return fmt.Sprintf(`
		WITH selected_match AS (
			SELECT *
			FROM matches
			WHERE id = %s::uuid
			  AND status = 'active'
			  AND %s::uuid IN (company_a_id, company_b_id)
			  AND %s::uuid IN (company_a_id, company_b_id)
			LIMIT 1
		),
		updated_intention AS (
			UPDATE consortium_intentions ci
			SET
				lead_company_id = %s::uuid,
				status = 'intention_registered',
				notes = NULLIF(%s, ''),
				updated_at = now()
			FROM selected_match sm
			WHERE ci.match_id = sm.id
			RETURNING ci.*
		),
		inserted_intention AS (
			INSERT INTO consortium_intentions (
				match_id, tender_id, lead_company_id, status, notes
			)
			SELECT
				sm.id,
				sm.tender_id,
				%s::uuid,
				'intention_registered',
				NULLIF(%s, '')
			FROM selected_match sm
			WHERE NOT EXISTS (SELECT 1 FROM updated_intention)
			RETURNING *
		),
		selected_intention AS (
			SELECT * FROM updated_intention
			UNION ALL
			SELECT * FROM inserted_intention
			LIMIT 1
		),
		inserted_members AS (
			INSERT INTO consortium_members (
				consortium_intention_id, company_id, role, responsibility_description
			)
			SELECT
				si.id,
				company_id,
				CASE WHEN company_id = %s::uuid THEN 'Lider do consorcio' ELSE 'Parceira consorciada' END,
				NULL
			FROM selected_intention si
			CROSS JOIN LATERAL (
				SELECT sm.company_a_id AS company_id FROM selected_match sm
				UNION
				SELECT sm.company_b_id AS company_id FROM selected_match sm
			) members
			ON CONFLICT (consortium_intention_id, company_id) DO UPDATE
			SET role = EXCLUDED.role
			RETURNING id
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				u.company_id,
				'match',
				'Lider do consorcio definido',
				c_lead.trade_name || ' foi definida como lider do consorcio.',
				'match-list',
				'match',
				sm.id
			FROM selected_match sm
			JOIN companies c_lead ON c_lead.id = %s::uuid
			JOIN users u ON u.company_id IN (sm.company_a_id, sm.company_b_id)
			WHERE u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				si.id::text AS id,
				si.match_id::text AS "matchId",
				si.tender_id::text AS "tenderId",
				si.lead_company_id::text AS "leadCompanyId",
				c.trade_name AS "leadCompanyName",
				si.status,
				COALESCE(si.notes, '') AS notes
			FROM selected_intention si
			JOIN companies c ON c.id = si.lead_company_id
		) item;
	`,
		sqlQuote(matchID),
		sqlQuote(session.CompanyID),
		sqlQuote(req.LeadCompanyID),
		sqlQuote(req.LeadCompanyID),
		sqlQuote(req.Notes),
		sqlQuote(req.LeadCompanyID),
		sqlQuote(req.Notes),
		sqlQuote(req.LeadCompanyID),
		sqlQuote(req.LeadCompanyID),
	)
}

func buildCreateConsortiumAdSQL(matchID string, session sessionUser, req createConsortiumAdRequest) string {
	return fmt.Sprintf(`
		WITH selected_intention AS (
			SELECT ci.*, m.tender_id AS match_tender_id, t.number, t.object, t.agency, lead.trade_name AS lead_name
			FROM consortium_intentions ci
			JOIN matches m ON m.id = ci.match_id
			JOIN tenders t ON t.id = m.tender_id
			JOIN companies lead ON lead.id = ci.lead_company_id
			WHERE ci.match_id = %s::uuid
			  AND ci.lead_company_id = %s::uuid
			  AND m.status = 'active'
			LIMIT 1
		),
		member_names AS (
			SELECT
				si.id AS consortium_intention_id,
				string_agg(c.trade_name, ', ' ORDER BY c.trade_name) AS names
			FROM selected_intention si
			JOIN consortium_members cm ON cm.consortium_intention_id = si.id
			JOIN companies c ON c.id = cm.company_id
			GROUP BY si.id
		),
		updated_ad AS (
			UPDATE partnership_ads pa
			SET
				title = 'Consorcio em formacao busca nova consorciada',
				offer_summary = 'Consorcio formado por: ' || COALESCE((SELECT names FROM member_names), si.lead_name),
				seek_summary = %s,
				status = 'published',
				published_at = COALESCE(pa.published_at, now()),
				updated_at = now(),
				ad_type = 'consortium',
				company_id = si.lead_company_id,
				tender_id = si.match_tender_id
			FROM selected_intention si
			WHERE pa.consortium_intention_id = si.id
			  AND pa.ad_type = 'consortium'
			  AND pa.deleted_at IS NULL
			RETURNING pa.*
		),
		inserted_ad AS (
			INSERT INTO partnership_ads (
				tender_id, company_id, consortium_intention_id, ad_type, title,
				offer_summary, seek_summary, status, published_at
			)
			SELECT
				si.match_tender_id,
				si.lead_company_id,
				si.id,
				'consortium',
				'Consorcio em formacao busca nova consorciada',
				'Consorcio formado por: ' || COALESCE((SELECT names FROM member_names), si.lead_name),
				%s,
				'published',
				now()
			FROM selected_intention si
			WHERE NOT EXISTS (SELECT 1 FROM updated_ad)
			RETURNING *
		),
		selected_ad AS (
			SELECT * FROM updated_ad
			UNION ALL
			SELECT * FROM inserted_ad
			LIMIT 1
		),
		updated_notes AS (
			UPDATE consortium_intentions ci
			SET notes = COALESCE(NULLIF(%s, ''), notes), updated_at = now()
			WHERE ci.id = (SELECT consortium_intention_id FROM selected_ad)
			RETURNING ci.id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				sa.id::text AS id,
				sa.tender_id::text AS "tenderId",
				sa.company_id::text AS "companyId",
				sa.consortium_intention_id::text AS "consortiumIntentionId",
				sa.ad_type AS "adType",
				sa.title,
				COALESCE(sa.offer_summary, '') AS "offerSummary",
				COALESCE(sa.seek_summary, '') AS "seekSummary",
				sa.status,
				sa.published_at AS "publishedAt"
			FROM selected_ad sa
		) item;
	`,
		sqlQuote(matchID),
		sqlQuote(session.CompanyID),
		sqlQuote(req.NeedSummary),
		sqlQuote(req.NeedSummary),
		sqlQuote(req.Notes),
	)
}

func buildAcceptConsortiumApplicationSQL(matchID string, session sessionUser, req acceptConsortiumApplicationRequest) string {
	return fmt.Sprintf(`
		WITH selected_application AS (
			SELECT ca.*, ci.lead_company_id, ci.match_id, ci.tender_id
			FROM consortium_applications ca
			JOIN consortium_intentions ci ON ci.id = ca.consortium_intention_id
			JOIN matches m ON m.id = ci.match_id
			WHERE ca.id = %s::uuid
			  AND ci.match_id = %s::uuid
			  AND ci.lead_company_id = %s::uuid
			  AND ca.status = 'interested'
			  AND m.status = 'active'
			LIMIT 1
		),
		updated_application AS (
			UPDATE consortium_applications ca
			SET status = 'accepted',
				leader_user_id = %s::uuid,
				updated_at = now()
			FROM selected_application sa
			WHERE ca.id = sa.id
			RETURNING ca.*
		),
		inserted_member AS (
			INSERT INTO consortium_members (
				consortium_intention_id, company_id, role, responsibility_description
			)
			SELECT
				sa.consortium_intention_id,
				sa.applicant_company_id,
				'Nova consorciada',
				'Empresa aceita pela lider para composicao complementar do consorcio.'
			FROM selected_application sa
			ON CONFLICT (consortium_intention_id, company_id) DO UPDATE
			SET role = EXCLUDED.role,
				responsibility_description = EXCLUDED.responsibility_description
			RETURNING *
		),
		closed_consortium_ad AS (
			UPDATE partnership_ads pa
			SET status = 'closed', updated_at = now()
			FROM selected_application sa
			JOIN consortium_intentions ci ON ci.id = sa.consortium_intention_id
			WHERE pa.tender_id = ci.tender_id
			  AND (
				pa.company_id = sa.applicant_company_id
				OR EXISTS (
					SELECT 1 FROM consortium_members cm
					WHERE cm.consortium_intention_id = sa.consortium_intention_id
					  AND cm.company_id = pa.company_id
				)
			  )
			  AND pa.status = 'published'
			RETURNING pa.id
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				sa.applicant_company_id,
				'match',
				'Entrada no consorcio aprovada',
				c_lead.trade_name || ' aprovou sua entrada no consorcio.',
				'match-list',
				'consortium_application',
				sa.id
			FROM selected_application sa
			JOIN companies c_lead ON c_lead.id = sa.lead_company_id
			JOIN users u ON u.company_id = sa.applicant_company_id
			WHERE u.status = 'active'
			  AND u.deleted_at IS NULL
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				ua.id::text AS id,
				ua.status,
				ua.applicant_company_id::text AS "companyId",
				c.trade_name AS "companyName"
			FROM updated_application ua
			JOIN companies c ON c.id = ua.applicant_company_id
		) item;
	`,
		sqlQuote(req.ApplicationID),
		sqlQuote(matchID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.UserID),
	)
}

func buildWithdrawFromConsortiumSQL(matchID string, session sessionUser, req withdrawFromConsortiumRequest) string {
	return fmt.Sprintf(`
		WITH selected_consortium AS (
			SELECT
				m.id AS match_id,
				ci.id AS consortium_intention_id,
				ci.tender_id,
				ci.lead_company_id,
				(
					SELECT count(*)
					FROM consortium_members cm_count
					WHERE cm_count.consortium_intention_id = ci.id
					  AND cm_count.status = 'active'
				) AS active_member_count
			FROM matches m
			JOIN consortium_intentions ci ON ci.match_id = m.id
			WHERE m.id = %s::uuid
			  AND m.status = 'active'
			  AND EXISTS (
				SELECT 1
				FROM consortium_members cm_current
				WHERE cm_current.consortium_intention_id = ci.id
				  AND cm_current.company_id = %s::uuid
				  AND cm_current.status = 'active'
			  )
			  AND NOT (
				ci.lead_company_id = %s::uuid
				AND (
					SELECT count(*) FROM consortium_members cm_count
					WHERE cm_count.consortium_intention_id = ci.id
					  AND cm_count.status = 'active'
				) - 1 >= 2
				AND NOT EXISTS (
					SELECT 1
					FROM consortium_members cm_successor
					WHERE cm_successor.consortium_intention_id = ci.id
					  AND cm_successor.company_id = NULLIF(%s, '')::uuid
					  AND cm_successor.company_id <> %s::uuid
					  AND cm_successor.status = 'active'
				)
			  )
			LIMIT 1
		), withdrawn_member AS (
			UPDATE consortium_members cm
			SET
				status = 'withdrawn',
				withdrawn_at = now(),
				withdrawn_by_user_id = %s::uuid,
				role = 'Retirada do consorcio'
			FROM selected_consortium sc
			WHERE cm.consortium_intention_id = sc.consortium_intention_id
			  AND cm.company_id = %s::uuid
			  AND cm.status = 'active'
			RETURNING sc.*
		), updated_consortium AS (
			UPDATE consortium_intentions ci
			SET
				lead_company_id = CASE
					WHEN ci.lead_company_id = %s::uuid AND wm.active_member_count - 1 >= 2
						THEN NULLIF(%s, '')::uuid
					ELSE ci.lead_company_id
				END,
				status = CASE WHEN wm.active_member_count - 1 < 2 THEN 'withdrawn' ELSE 'intention_registered' END,
				updated_at = now()
			FROM withdrawn_member wm
			WHERE ci.id = wm.consortium_intention_id
			RETURNING ci.*, wm.match_id, wm.active_member_count
		), cancelled_small_consortium AS (
			UPDATE matches m
			SET status = 'cancelled', updated_at = now()
			FROM updated_consortium uc
			WHERE m.id = uc.match_id
			  AND uc.active_member_count - 1 < 2
			RETURNING m.id
		), closed_ads AS (
			UPDATE partnership_ads pa
			SET status = 'closed', updated_at = now()
			FROM updated_consortium uc
			WHERE pa.tender_id = uc.tender_id
			  AND pa.status = 'published'
			  AND (
				pa.consortium_intention_id = uc.id
				OR pa.company_id = %s::uuid
			  )
			RETURNING pa.id
		), created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				u.company_id,
				'system',
				'Empresa desistiu do consorcio',
				c_withdrawn.trade_name || ' deixou a composicao deste consorcio.',
				'match-list',
				'match',
				uc.match_id
			FROM updated_consortium uc
			JOIN consortium_members cm ON cm.consortium_intention_id = uc.id AND cm.status = 'active'
			JOIN users u ON u.company_id = cm.company_id AND u.status = 'active' AND u.deleted_at IS NULL
			JOIN companies c_withdrawn ON c_withdrawn.id = %s::uuid
			RETURNING id
		), created_audit AS (
			INSERT INTO audit_logs (actor_user_id, company_id, module, action, entity_type, entity_id, description, metadata)
			SELECT
				%s::uuid,
				%s::uuid,
				'consorcios',
				'desistencia',
				'consortium_intention',
				uc.id,
				'Empresa desistiu do consorcio.',
				jsonb_build_object('matchId', uc.match_id, 'successorCompanyId', NULLIF(%s, ''))
			FROM updated_consortium uc
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				uc.match_id::text AS "matchId",
				uc.id::text AS "consortiumIntentionId",
				CASE WHEN uc.active_member_count - 1 < 2 THEN true ELSE false END AS "consortiumClosed"
			FROM updated_consortium uc
		) item;
	`,
		sqlQuote(matchID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(req.SuccessorCompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.UserID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(req.SuccessorCompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.CompanyID),
		sqlQuote(session.UserID),
		sqlQuote(session.CompanyID),
		sqlQuote(req.SuccessorCompanyID),
	)
}

func matchesSQL(session sessionUser, extraWhere string) string {
	companyFilter := ""
	if session.CompanyID != "" {
		companyFilter = fmt.Sprintf(`AND (
			(NOT EXISTS (SELECT 1 FROM consortium_intentions ci_existing WHERE ci_existing.match_id = m.id)
			 AND %s::uuid IN (m.company_a_id, m.company_b_id))
			OR EXISTS (
				SELECT 1
				FROM consortium_intentions ci_filter
				JOIN consortium_members cm_filter ON cm_filter.consortium_intention_id = ci_filter.id
				WHERE ci_filter.match_id = m.id
				  AND cm_filter.company_id = %s::uuid
				  AND cm_filter.status = 'active'
			)
		)`, sqlQuote(session.CompanyID), sqlQuote(session.CompanyID))
	}
	return fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."matchedAt" DESC), '[]'::json)
		FROM (
			SELECT
				m.id::text AS id,
				m.tender_id::text AS "tenderId",
				t.agency,
				t.number AS "tenderNumber",
				t.object AS "tenderObject",
				m.status,
				m.matched_at AS "matchedAt",
				COALESCE(ci.lead_company_id::text, '') AS "leadCompanyId",
				COALESCE(ci.id::text, '') AS "consortiumIntentionId",
				COALESCE(lead.trade_name, '') AS "leadCompanyName",
				COALESCE(ci.status, '') AS "consortiumStatus",
				COALESCE(ci.notes, '') AS "consortiumNotes",
				COALESCE(cons_ad.id::text, '') AS "consortiumAdId",
				COALESCE(cons_ad.seek_summary, '') AS "consortiumAdSeekSummary",
				a.trade_name AS "companyAName",
				b.trade_name AS "companyBName",
				COALESCE((
					SELECT json_agg(row_to_json(member) ORDER BY member."companyName")
					FROM (
						SELECT
							cm.company_id::text AS "companyId",
							member_company.trade_name AS "companyName",
							COALESCE(cm.role, '') AS role
						FROM consortium_members cm
						JOIN companies member_company ON member_company.id = cm.company_id
						WHERE cm.consortium_intention_id = ci.id
						  AND cm.status = 'active'
					) member
				), '[]'::json) AS "consortiumMembers",
				COALESCE((
					SELECT json_agg(row_to_json(app_item) ORDER BY app_item."createdAt" DESC)
					FROM (
						SELECT
							ca.id::text AS id,
							ca.applicant_company_id::text AS "companyId",
							app_company.trade_name AS "companyName",
							ca.status,
							ca.created_at AS "createdAt"
						FROM consortium_applications ca
						JOIN companies app_company ON app_company.id = ca.applicant_company_id
						WHERE ca.consortium_intention_id = ci.id
					) app_item
				), '[]'::json) AS applications,
				COALESCE((
					SELECT json_agg(row_to_json(contact))
					FROM (
						SELECT
							mc.company_id::text AS "companyId",
							c.trade_name AS "companyName",
							mc.contact_name AS "contactName",
							mc.phone,
							COALESCE(mc.whatsapp_url, '') AS "whatsappUrl"
						FROM match_contacts mc
						JOIN companies c ON c.id = mc.company_id
						WHERE mc.match_id = m.id
					) contact
				), '[]'::json) AS contacts
			FROM matches m
			JOIN tenders t ON t.id = m.tender_id
			JOIN companies a ON a.id = m.company_a_id
			JOIN companies b ON b.id = m.company_b_id
			LEFT JOIN consortium_intentions ci ON ci.match_id = m.id
			LEFT JOIN companies lead ON lead.id = ci.lead_company_id
			LEFT JOIN partnership_ads cons_ad ON cons_ad.consortium_intention_id = ci.id
				AND cons_ad.ad_type = 'consortium'
				AND cons_ad.deleted_at IS NULL
				AND cons_ad.status = 'published'
			WHERE m.status = 'active'
			  %s
			  %s
		) item;
	`, companyFilter, extraWhere)
}

func myUserProfileSQL() string {
	return `
		SELECT row_to_json(item)
		FROM (
			SELECT
				u.id::text AS id,
				COALESCE(u.company_id::text, '') AS "companyId",
				COALESCE(c.trade_name, 'LicitaHub') AS "companyName",
				u.full_name AS "fullName",
				u.email,
				COALESCE(u.phone, '') AS phone,
				COALESCE(u.job_title, '') AS "jobTitle",
				ap.key AS "roleKey",
				ap.name AS "roleName",
				COALESCE(photo.file_url, '') AS "profilePhotoUrl",
				COALESCE(company_logo.file_url, '') AS "companyLogoUrl",
				u.updated_at AS "updatedAt"
			FROM users u
			JOIN access_profiles ap ON ap.id = u.access_profile_id
			LEFT JOIN companies c ON c.id = u.company_id
			LEFT JOIN media_files photo ON photo.id = u.profile_photo_media_id
			LEFT JOIN company_profiles cp ON cp.company_id = c.id
			LEFT JOIN media_files company_logo ON company_logo.id = cp.logo_media_id
			WHERE u.id = %s::uuid
			  AND u.deleted_at IS NULL
			LIMIT 1
		) item;
	`
}

func buildUpdateMyUserProfileSQL(session sessionUser, req updateMyUserProfileRequest) string {
	photoCTE := "selected_photo AS (SELECT NULL::uuid AS id)"
	if req.ProfilePhotoURL != "" {
		companyForPhoto := "NULL"
		if session.CompanyID != "" {
			companyForPhoto = sqlQuote(session.CompanyID) + "::uuid"
		}
		photoCTE = fmt.Sprintf(`
			selected_photo AS (
				INSERT INTO media_files (company_id, uploaded_by_user_id, media_type, file_name, file_url, mime_type, source)
				VALUES (%s, %s::uuid, 'profile_photo', NULLIF(%s, ''), %s, NULLIF(%s, ''), 'upload')
				RETURNING id
			)`,
			companyForPhoto,
			sqlQuote(session.UserID),
			sqlQuote(req.ProfilePhotoFileName),
			sqlQuote(req.ProfilePhotoURL),
			sqlQuote(req.ProfilePhotoMimeType),
		)
	}
	return fmt.Sprintf(`
		WITH
		%s,
		updated_user AS (
			UPDATE users
			SET
				full_name = %s,
				email = %s,
				phone = %s,
				profile_photo_media_id = COALESCE((SELECT id FROM selected_photo), profile_photo_media_id),
				updated_at = now()
			WHERE id = %s::uuid
			  AND deleted_at IS NULL
			RETURNING id
		)
		%s
	`,
		photoCTE,
		sqlQuote(req.FullName),
		sqlQuote(req.Email),
		sqlQuote(req.Phone),
		sqlQuote(session.UserID),
		fmt.Sprintf(myUserProfileSQL(), sqlQuote(session.UserID)),
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
				published_at,
				expires_at,
				archived_at
			)
			SELECT
				selected_category.id,
				%s,
				%s,
				NULLIF(%s, ''),
				NULLIF(%s, ''),
				(SELECT id FROM inserted_media),
				CASE WHEN %s IN ('draft', 'archived', 'expired') THEN NULL ELSE now() END,
				CASE
					WHEN NULLIF(%s, '') IS NULL THEN NULL
					ELSE (NULLIF(%s, '')::date + interval '1 day' - interval '1 second')
				END,
				CASE WHEN %s IN ('archived', 'expired') THEN now() ELSE NULL END
			FROM selected_category
			RETURNING *
		),
		created_notifications AS (
			INSERT INTO notifications (
				recipient_user_id, recipient_company_id, type, title, message,
				destination_screen, related_entity_type, related_entity_id
			)
			SELECT
				u.id,
				u.company_id,
				'news',
				'Nova noticia no Radar',
				%s || ' foi publicada no Radar LicitaHub.',
				'radar-home',
				'news',
				n.id
			FROM inserted_news n
			JOIN users u ON u.status = 'active'
			WHERE n.status IN ('published', 'featured')
			  AND u.deleted_at IS NULL
			  AND u.company_id IS NOT NULL
			RETURNING id
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
				n.expires_at AS "expiresAt",
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
		sqlQuote(req.ExpiresAt),
		sqlQuote(req.ExpiresAt),
		sqlQuote(status),
		sqlQuote(req.Title),
	)
}

func communityPostsSQL(session sessionUser, whereExtra string) string {
	return fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json)
		FROM (
			SELECT
				p.id::text AS id,
				COALESCE(pc.name, 'Publicacao') AS category,
				COALESCE(pc.slug, '') AS "categorySlug",
				c.trade_name AS company,
				c.id::text AS "companyId",
				COALESCE(cp.state, c.state, 'BR') AS region,
				COALESCE(logo.file_url, '') AS "companyLogoUrl",
				COALESCE(u.full_name, '') AS "authorName",
				COALESCE(p.title, '') AS title,
				p.content AS text,
				COALESCE(media.file_url, '') AS "imageUrl",
				p.visibility,
				p.status,
				p.origin,
				p.published_at AS "publishedAt",
				p.created_at AS "createdAt",
				EXISTS (
					SELECT 1 FROM post_likes pl
					WHERE pl.post_id = p.id AND pl.user_id = %s::uuid
				) AS liked,
				EXISTS (
					SELECT 1 FROM post_favorites pf
					WHERE pf.post_id = p.id AND pf.user_id = %s::uuid
				) AS saved,
				(SELECT count(*) FROM post_likes pl WHERE pl.post_id = p.id)::int AS "likeCount",
				(SELECT count(*) FROM post_comments pct WHERE pct.post_id = p.id AND pct.deleted_at IS NULL)::int AS "commentCount",
				COALESCE((
					SELECT json_agg(name ORDER BY liked_at DESC)
					FROM (
						SELECT COALESCE(like_company.trade_name, like_user.full_name, 'Usuario') AS name, pl.created_at AS liked_at
						FROM post_likes pl
						JOIN users like_user ON like_user.id = pl.user_id
						LEFT JOIN companies like_company ON like_company.id = pl.company_id
						WHERE pl.post_id = p.id
						ORDER BY pl.created_at DESC
						LIMIT 20
					) liked_names
				), '[]'::json) AS likes,
				COALESCE((
					SELECT json_agg(row_to_json(comment_item) ORDER BY comment_item."createdAt")
					FROM (
						SELECT
							pct.id::text AS id,
							COALESCE(comment_company.trade_name, comment_user.full_name, 'Usuario') AS company,
							comment_user.full_name AS "userName",
							pct.content AS text,
							(pct.user_id = %s::uuid) AS "canEdit",
							(pct.user_id = %s::uuid OR p.company_id = NULLIF(%s, '')::uuid) AS "canDelete",
							pct.created_at AS "createdAt"
						FROM post_comments pct
						JOIN users comment_user ON comment_user.id = pct.user_id
						LEFT JOIN companies comment_company ON comment_company.id = pct.company_id
						WHERE pct.post_id = p.id
						  AND pct.deleted_at IS NULL
						ORDER BY pct.created_at ASC
					) comment_item
				), '[]'::json) AS comments
			FROM posts p
			JOIN companies c ON c.id = p.company_id
			LEFT JOIN company_profiles cp ON cp.company_id = c.id
			LEFT JOIN media_files logo ON logo.id = cp.logo_media_id
			LEFT JOIN users u ON u.id = p.author_user_id
			LEFT JOIN post_categories pc ON pc.id = p.category_id
			LEFT JOIN media_files media ON media.id = p.main_image_media_id
			WHERE p.deleted_at IS NULL
			  AND p.status = 'published'
			  %s
			ORDER BY p.published_at DESC NULLS LAST, p.created_at DESC
		) item;
	`, sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), whereExtra)
}

func buildCreateCommunityPostSQL(session sessionUser, req createPostRequest) string {
	imageURL := nullOrQuote(req.MainImageURL)
	imageFileName := req.MainImageFileName
	if imageFileName == "" {
		imageFileName = "imagem-publicacao"
	}
	titleSQL := nullOrQuote(req.Title)

	return fmt.Sprintf(`
		WITH selected_category AS (
			SELECT id
			FROM post_categories
			WHERE is_active = true
			  AND slug IN (%s, 'noticias')
			ORDER BY CASE WHEN slug = %s THEN 0 ELSE 1 END
			LIMIT 1
		),
		inserted_media AS (
			INSERT INTO media_files (company_id, uploaded_by_user_id, media_type, file_name, file_url, mime_type, source)
			SELECT %s::uuid, %s::uuid, 'image', %s, %s, %s, 'upload'
			WHERE %s IS NOT NULL
			RETURNING id, file_url
		),
		inserted_post AS (
			INSERT INTO posts (
				company_id,
				author_user_id,
				category_id,
				title,
				content,
				main_image_media_id,
				status,
				visibility,
				origin,
				published_at
			)
			SELECT
				%s::uuid,
				%s::uuid,
				selected_category.id,
				%s,
				%s,
				(SELECT id FROM inserted_media),
				'published',
				%s,
				'manual',
				now()
			FROM selected_category
			RETURNING id
		)
		SELECT row_to_json(item)
		FROM (
			SELECT
				p.id::text AS id,
				COALESCE(pc.name, 'Publicacao') AS category,
				COALESCE(pc.slug, '') AS "categorySlug",
				c.trade_name AS company,
				c.id::text AS "companyId",
				COALESCE(cp.state, c.state, 'BR') AS region,
				COALESCE(logo.file_url, '') AS "companyLogoUrl",
				COALESCE(u.full_name, '') AS "authorName",
				COALESCE(p.title, '') AS title,
				p.content AS text,
				COALESCE(media.file_url, '') AS "imageUrl",
				p.visibility,
				p.status,
				p.origin,
				p.published_at AS "publishedAt",
				p.created_at AS "createdAt",
				false AS liked,
				false AS saved,
				0 AS "likeCount",
				0 AS "commentCount",
				'[]'::json AS likes,
				'[]'::json AS comments
			FROM inserted_post inserted
			JOIN posts p ON p.id = inserted.id
			JOIN companies c ON c.id = p.company_id
			LEFT JOIN company_profiles cp ON cp.company_id = c.id
			LEFT JOIN media_files logo ON logo.id = cp.logo_media_id
			LEFT JOIN users u ON u.id = p.author_user_id
			LEFT JOIN post_categories pc ON pc.id = p.category_id
			LEFT JOIN media_files media ON media.id = p.main_image_media_id
		) item;
	`,
		sqlQuote(req.CategorySlug),
		sqlQuote(req.CategorySlug),
		sqlQuote(session.CompanyID),
		sqlQuote(session.UserID),
		sqlQuote(imageFileName),
		imageURL,
		nullOrQuote(req.MainImageMimeType),
		imageURL,
		sqlQuote(session.CompanyID),
		sqlQuote(session.UserID),
		titleSQL,
		sqlQuote(req.Content),
		sqlQuote(req.Visibility),
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

func saveCommunityImage(req createPostRequest) (string, error) {
	header, payload, ok := strings.Cut(req.MainImageDataURL, ",")
	if !ok || !strings.HasPrefix(header, "data:image/") {
		return "", errors.New("imagem da publicacao invalida")
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

	dir := filepath.Join("uploads", "community")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", errors.New("nao foi possivel preparar a pasta de imagens da comunidade")
	}

	fileName := fmt.Sprintf("post-%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(dir, fileName)
	if err := os.WriteFile(path, bytes, 0644); err != nil {
		return "", errors.New("nao foi possivel salvar a imagem da publicacao")
	}

	return strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/") + "/uploads/community/" + fileName, nil
}

func saveProfilePhoto(dataURL string, mimeType string) (string, error) {
	header, payload, ok := strings.Cut(dataURL, ",")
	if !ok || !strings.HasPrefix(header, "data:image/") {
		return "", errors.New("foto de perfil invalida")
	}
	if mimeType == "" {
		mimeType = strings.TrimPrefix(strings.TrimSuffix(header, ";base64"), "data:")
	}
	ext := imageExtension(mimeType)
	if ext == "" {
		return "", errors.New("tipo de foto nao permitido")
	}
	bytes, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", errors.New("nao foi possivel ler a foto")
	}
	if len(bytes) > 5*1024*1024 {
		return "", errors.New("foto maior que 5MB")
	}
	dir := filepath.Join("uploads", "profiles")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", errors.New("nao foi possivel preparar a pasta de fotos")
	}
	fileName := fmt.Sprintf("profile-%d%s", time.Now().UnixNano(), ext)
	if err := os.WriteFile(filepath.Join(dir, fileName), bytes, 0644); err != nil {
		return "", errors.New("nao foi possivel salvar a foto")
	}
	return strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/") + "/uploads/profiles/" + fileName, nil
}

func saveCompanyLogo(dataURL string, mimeType string) (string, error) {
	header, payload, ok := strings.Cut(dataURL, ",")
	if !ok || !strings.HasPrefix(header, "data:image/") {
		return "", errors.New("logomarca invalida")
	}
	if mimeType == "" {
		mimeType = strings.TrimPrefix(strings.TrimSuffix(header, ";base64"), "data:")
	}
	ext := imageExtension(mimeType)
	if ext == "" {
		return "", errors.New("tipo de logomarca nao permitido")
	}
	bytes, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", errors.New("nao foi possivel ler a logomarca")
	}
	if len(bytes) > 5*1024*1024 {
		return "", errors.New("logomarca maior que 5MB")
	}
	dir := filepath.Join("uploads", "companies")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", errors.New("nao foi possivel preparar a pasta de logomarcas")
	}
	fileName := fmt.Sprintf("company-logo-%d%s", time.Now().UnixNano(), ext)
	if err := os.WriteFile(filepath.Join(dir, fileName), bytes, 0644); err != nil {
		return "", errors.New("nao foi possivel salvar a logomarca")
	}
	return strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/") + "/uploads/companies/" + fileName, nil
}

func saveTenderHTML(dataURL string, prefix string) (string, error) {
	header, payload, ok := strings.Cut(dataURL, ",")
	if !ok || (!strings.Contains(header, "text/html") && !strings.Contains(header, "application/octet-stream")) {
		return "", errors.New("arquivo de analise deve estar no formato HTML")
	}
	bytes, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", errors.New("nao foi possivel ler o arquivo HTML")
	}
	if len(bytes) > 10*1024*1024 {
		return "", errors.New("arquivo HTML maior que 10MB")
	}
	dir := filepath.Join("uploads", "tenders")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", errors.New("nao foi possivel preparar a pasta de editais")
	}
	fileName := fmt.Sprintf("%s-%d.html", prefix, time.Now().UnixNano())
	if err := os.WriteFile(filepath.Join(dir, fileName), bytes, 0644); err != nil {
		return "", errors.New("nao foi possivel salvar o arquivo HTML")
	}
	return strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/") + "/uploads/tenders/" + fileName, nil
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
	case "image/svg+xml":
		return ".svg"
	default:
		return ""
	}
}

func normalizeCompanySize(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "nao_informado", "não informado":
		return ""
	case "small", "pequena":
		return "small"
	case "medium", "media", "média":
		return "medium"
	case "large", "grande":
		return "large"
	default:
		return ""
	}
}

func normalizeNewsStatus(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "draft", "rascunho":
		return "draft", nil
	case "published", "publicado", "publicada", "disponivel", "disponível":
		return "published", nil
	case "featured", "destaque", "destaque principal":
		return "featured", nil
	case "archived", "arquivado", "arquivada", "antiga":
		return "archived", nil
	case "expired", "expirado", "expirada", "vencida":
		return "expired", nil
	default:
		return "", errors.New("status da noticia invalido")
	}
}

func normalizeTenderStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "impugnado", "impugnada":
		return "challenged"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func isValidTenderStatus(value string) bool {
	switch value {
	case "draft", "published", "under_review", "suspended", "challenged", "occurred", "closed", "cancelled":
		return true
	default:
		return false
	}
}

func normalizePostVisibility(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "profile", "perfil":
		return "profile"
	case "both", "ambos", "comunidade_e_perfil":
		return "both"
	default:
		return "community"
	}
}

func normalizeInterestGeneralPosition(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "under_evaluation", "watching", "not_interested":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "interested"
	}
}

func normalizeInterestDesiredRole(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "can_lead_consortium", "complementary_partner", "seeks_lead_company", "evaluating_role":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "seeks_partner"
	}
}

func parseNewsExpiresAt(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("informe ate quando a noticia ficara publicada")
	}

	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, errors.New("data final de publicacao invalida")
	}

	return date.Add(24*time.Hour - time.Second), nil
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

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "licitahub_session",
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "licitahub_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})
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

func firstNonEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
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
