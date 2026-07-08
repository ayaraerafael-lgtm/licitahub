package main

import (
	"context"
	"encoding/base64"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", application.handleHealth)
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
		writeError(w, http.StatusInternalServerError, "nao foi possivel cadastrar a noticia")
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

func (a *app) queryJSON(ctx context.Context, sql string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	output, err := a.runPSQL(ctx, sql)
	if err != nil {
		return nil, err
	}

	payload := strings.TrimSpace(output)
	if payload == "" {
		payload = "null"
	}

	return []byte(payload), nil
}

func (a *app) runPSQL(ctx context.Context, sql string) (string, error) {
	cmd := exec.CommandContext(
		ctx,
		a.psqlPath,
		"-h", a.pg.Host,
		"-p", a.pg.Port,
		"-U", a.pg.User,
		"-d", a.pg.Database,
		"-tA",
		"-c", sql,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+a.pg.Password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("psql error: %s", strings.TrimSpace(string(output)))
		return "", err
	}

	return string(output), nil
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
