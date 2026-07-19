package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type academyCourseRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	CoverImageURL string `json:"coverImageUrl"`
	WorkloadHours int    `json:"workloadHours"`
	Status        string `json:"status"`
}

type academyLessonRequest struct {
	Title          string                `json:"title"`
	Description    string                `json:"description"`
	VideoURL       string                `json:"videoUrl"`
	VideoSource    string                `json:"videoSource"`
	DurationSecond int                   `json:"durationSeconds"`
	RequiresQuiz   bool                  `json:"requiresQuiz"`
	PassingScore   int                   `json:"passingScore"`
	MaxAttempts    int                   `json:"maxAttempts"`
	Questions      []academyQuizQuestion `json:"questions"`
	IsPublished    *bool                 `json:"isPublished"`
}

type academyQuizQuestion struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Correct  int      `json:"correct"`
}

type academyProgressRequest struct {
	PositionSeconds     int  `json:"positionSeconds"`
	WatchedSecondsDelta int  `json:"watchedSecondsDelta"`
	DurationSeconds     int  `json:"durationSeconds"`
	Ended               bool `json:"ended"`
}

type academyAssessmentRequest struct {
	Answers []int `json:"answers"`
}

func (a *app) ensureAcademyMigrations(ctx context.Context) error {
	_, err := a.runPSQL(ctx, `
		CREATE TABLE IF NOT EXISTS academy_courses (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			title varchar(220) NOT NULL,
			description text NOT NULL DEFAULT '',
			category varchar(120) NOT NULL DEFAULT 'Geral',
			cover_image_url text,
			workload_hours integer NOT NULL DEFAULT 0,
			status varchar(30) NOT NULL DEFAULT 'draft',
			created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now(),
			published_at timestamptz,
			CONSTRAINT academy_courses_status_chk CHECK (status IN ('draft', 'published', 'archived'))
		);
		CREATE TABLE IF NOT EXISTS academy_lessons (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			course_id uuid NOT NULL REFERENCES academy_courses(id) ON DELETE CASCADE,
			title varchar(220) NOT NULL,
			description text NOT NULL DEFAULT '',
			video_url text NOT NULL,
			video_source varchar(20) NOT NULL DEFAULT 'youtube',
			duration_seconds integer NOT NULL DEFAULT 0,
			display_order integer NOT NULL DEFAULT 1,
			requires_quiz boolean NOT NULL DEFAULT false,
			quiz_questions jsonb NOT NULL DEFAULT '[]'::jsonb,
			passing_score integer NOT NULL DEFAULT 75,
			max_attempts integer NOT NULL DEFAULT 0,
			is_published boolean NOT NULL DEFAULT true,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now(),
			CONSTRAINT academy_lessons_score_chk CHECK (passing_score BETWEEN 1 AND 100),
			CONSTRAINT academy_lessons_attempts_chk CHECK (max_attempts >= 0),
			CONSTRAINT academy_lessons_video_source_chk CHECK (video_source IN ('youtube', 'upload'))
		);
		CREATE TABLE IF NOT EXISTS academy_enrollments (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			course_id uuid NOT NULL REFERENCES academy_courses(id) ON DELETE CASCADE,
			user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			started_at timestamptz NOT NULL DEFAULT now(),
			last_accessed_at timestamptz NOT NULL DEFAULT now(),
			completed_at timestamptz,
			UNIQUE(course_id, user_id)
		);
		CREATE TABLE IF NOT EXISTS academy_lesson_progress (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			lesson_id uuid NOT NULL REFERENCES academy_lessons(id) ON DELETE CASCADE,
			user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			last_position_seconds integer NOT NULL DEFAULT 0,
			watched_seconds integer NOT NULL DEFAULT 0,
			video_completed boolean NOT NULL DEFAULT false,
			is_completed boolean NOT NULL DEFAULT false,
			completed_at timestamptz,
			updated_at timestamptz NOT NULL DEFAULT now(),
			UNIQUE(lesson_id, user_id)
		);
		ALTER TABLE academy_lesson_progress ADD COLUMN IF NOT EXISTS video_completed boolean NOT NULL DEFAULT false;
		ALTER TABLE academy_lessons ADD COLUMN IF NOT EXISTS video_source varchar(20) NOT NULL DEFAULT 'youtube';
		ALTER TABLE academy_lessons DROP CONSTRAINT IF EXISTS academy_lessons_video_source_chk;
		ALTER TABLE academy_lessons ADD CONSTRAINT academy_lessons_video_source_chk CHECK (video_source IN ('youtube', 'upload'));
		UPDATE academy_lesson_progress SET video_completed = true WHERE is_completed = true AND video_completed = false;
		ALTER TABLE academy_lessons ALTER COLUMN passing_score SET DEFAULT 75;
		UPDATE academy_lessons SET passing_score = 75 WHERE passing_score <> 75;
		CREATE TABLE IF NOT EXISTS academy_assessment_attempts (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			lesson_id uuid NOT NULL REFERENCES academy_lessons(id) ON DELETE CASCADE,
			user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			answers jsonb NOT NULL DEFAULT '[]'::jsonb,
			score numeric(5,2) NOT NULL,
			approved boolean NOT NULL DEFAULT false,
			attempted_at timestamptz NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS academy_certificates (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			course_id uuid NOT NULL REFERENCES academy_courses(id) ON DELETE CASCADE,
			user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			verification_code varchar(32) NOT NULL UNIQUE,
			issued_at timestamptz NOT NULL DEFAULT now(),
			UNIQUE(course_id, user_id)
		);
		CREATE INDEX IF NOT EXISTS idx_academy_courses_status ON academy_courses(status, published_at DESC);
		CREATE INDEX IF NOT EXISTS idx_academy_lessons_course ON academy_lessons(course_id, display_order);
		CREATE INDEX IF NOT EXISTS idx_academy_progress_user ON academy_lesson_progress(user_id, updated_at DESC);
	`)
	return err
}

func (a *app) handleAcademyCourses(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		filter := "AND course.status = 'published' AND EXISTS (SELECT 1 FROM academy_lessons available_lesson WHERE available_lesson.course_id = course.id AND available_lesson.is_published)"
		if session.canManagePlatform() {
			filter = ""
		}
		payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
			SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."publishedAt" DESC, item."createdAt" DESC), '[]'::json)
			FROM (
				SELECT course.id::text AS id, course.title, course.description, course.category,
					COALESCE(course.cover_image_url, '') AS "coverImageUrl", course.workload_hours AS "workloadHours", course.status,
					course.published_at AS "publishedAt", course.created_at AS "createdAt",
					(SELECT count(*) FROM academy_lessons lesson WHERE lesson.course_id = course.id AND lesson.is_published)::int AS "lessonCount",
					(SELECT count(*) FROM academy_enrollments enrollment_count WHERE enrollment_count.course_id = course.id)::int AS "enrollmentCount",
					EXISTS (SELECT 1 FROM academy_enrollments enrollment WHERE enrollment.course_id = course.id AND enrollment.user_id = %s::uuid) AS enrolled,
					COALESCE((SELECT round(100.0 * count(*) FILTER (WHERE progress.is_completed) / NULLIF(count(*), 0), 0)
						FROM academy_lessons lesson LEFT JOIN academy_lesson_progress progress ON progress.lesson_id = lesson.id AND progress.user_id = %s::uuid
						WHERE lesson.course_id = course.id AND lesson.is_published), 0)::int AS "progressPercent"
				FROM academy_courses course
				WHERE true %s
			) item;
		`, sqlQuote(session.UserID), sqlQuote(session.UserID), filter))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel carregar os cursos")
			return
		}
		writeRawJSON(w, http.StatusOK, payload)
	case http.MethodPost:
		if !session.canManagePlatform() {
			writeError(w, http.StatusForbidden, "somente administrador da plataforma pode criar cursos")
			return
		}
		var req academyCourseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "json invalido")
			return
		}
		req.Title, req.Description, req.Category, req.CoverImageURL = strings.TrimSpace(req.Title), strings.TrimSpace(req.Description), strings.TrimSpace(req.Category), strings.TrimSpace(req.CoverImageURL)
		if req.Title == "" {
			writeError(w, http.StatusBadRequest, "titulo do curso e obrigatorio")
			return
		}
		if req.Category == "" {
			req.Category = "Geral"
		}
		if req.WorkloadHours < 0 {
			req.WorkloadHours = 0
		}
		if req.Status != "published" {
			req.Status = "draft"
		}
		payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
			WITH inserted AS (
				INSERT INTO academy_courses (title, description, category, cover_image_url, workload_hours, status, created_by_user_id, published_at)
				VALUES (%s, %s, %s, NULLIF(%s, ''), %d, %s, %s::uuid, CASE WHEN %s = 'published' THEN now() ELSE NULL END)
				RETURNING *
			) SELECT row_to_json(item) FROM (SELECT id::text AS id, title, status FROM inserted) item;
		`, sqlQuote(req.Title), sqlQuote(req.Description), sqlQuote(req.Category), sqlQuote(req.CoverImageURL), req.WorkloadHours, sqlQuote(req.Status), sqlQuote(session.UserID), sqlQuote(req.Status)))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel criar o curso")
			return
		}
		writeRawJSON(w, http.StatusCreated, payload)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleAcademyVideoUpload(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode enviar videos")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}

	const maxVideoSize = int64(500 * 1024 * 1024)
	r.Body = http.MaxBytesReader(w, r.Body, maxVideoSize+1024*1024)
	if err := r.ParseMultipartForm(8 * 1024 * 1024); err != nil {
		writeError(w, http.StatusBadRequest, "video invalido ou maior que 500 MB")
		return
	}
	file, header, err := r.FormFile("video")
	if err != nil {
		writeError(w, http.StatusBadRequest, "selecione o arquivo do video")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]string{".mp4": "video/mp4", ".webm": "video/webm"}
	mimeType, allowedType := allowed[ext]
	if !allowedType {
		writeError(w, http.StatusBadRequest, "use um video MP4 ou WebM")
		return
	}
	dir := filepath.Join("uploads", "academy")
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel preparar a pasta de videos")
		return
	}
	token, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel preparar o video")
		return
	}
	fileName := "academy-" + token + ext
	path := filepath.Join(dir, fileName)
	output, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel salvar o video")
		return
	}
	size, copyErr := io.Copy(output, io.LimitReader(file, maxVideoSize+1))
	closeErr := output.Close()
	if copyErr != nil || closeErr != nil || size > maxVideoSize {
		_ = os.Remove(path)
		if size > maxVideoSize {
			writeError(w, http.StatusBadRequest, "o video deve ter no maximo 500 MB")
		} else {
			writeError(w, http.StatusInternalServerError, "nao foi possivel salvar o video")
		}
		return
	}
	url := strings.TrimRight(getenv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/") + "/uploads/academy/" + fileName
	writeJSON(w, http.StatusCreated, map[string]any{
		"url":      url,
		"fileName": header.Filename,
		"mimeType": mimeType,
		"fileSize": size,
	})
}

func (a *app) handleAcademyCourseByPath(w http.ResponseWriter, r *http.Request) {
	courseID, action := splitResourcePath(r.URL.Path, "/api/academy/courses/")
	if courseID == "" {
		writeError(w, http.StatusNotFound, "curso nao encontrado")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if action == "enroll" && r.Method == http.MethodPost {
		a.enrollAcademyCourse(w, r, courseID, session)
		return
	}
	if action == "lessons" && r.Method == http.MethodPost {
		a.createAcademyLesson(w, r, courseID, session)
		return
	}
	if action == "certificate" && r.Method == http.MethodGet {
		a.downloadAcademyCertificate(w, r, courseID, session)
		return
	}
	if action == "" && r.Method == http.MethodGet {
		a.getAcademyCourse(w, r, courseID, session)
		return
	}
	if action == "" && r.Method == http.MethodPut {
		a.updateAcademyCourse(w, r, courseID, session)
		return
	}
	if action == "" && r.Method == http.MethodDelete {
		a.deleteAcademyCourse(w, r, courseID, session)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "acao de curso nao permitida")
}

func (a *app) updateAcademyCourse(w http.ResponseWriter, r *http.Request, courseID string, session sessionUser) {
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode editar cursos")
		return
	}
	var req academyCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Category = strings.TrimSpace(req.Category)
	req.CoverImageURL = strings.TrimSpace(req.CoverImageURL)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "titulo do curso e obrigatorio")
		return
	}
	if req.Category == "" {
		req.Category = "Geral"
	}
	if req.WorkloadHours < 0 {
		req.WorkloadHours = 0
	}
	if req.Status != "draft" && req.Status != "published" && req.Status != "archived" {
		writeError(w, http.StatusBadRequest, "situacao do curso invalida")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH updated AS (
			UPDATE academy_courses
			SET title = %s, description = %s, category = %s, cover_image_url = NULLIF(%s, ''), workload_hours = %d,
				status = %s, published_at = CASE WHEN %s = 'published' THEN COALESCE(published_at, now()) ELSE published_at END, updated_at = now()
			WHERE id = %s::uuid
			RETURNING *
		) SELECT row_to_json(item) FROM (SELECT id::text AS id, title, status FROM updated) item;
	`, sqlQuote(req.Title), sqlQuote(req.Description), sqlQuote(req.Category), sqlQuote(req.CoverImageURL), req.WorkloadHours, sqlQuote(req.Status), sqlQuote(req.Status), sqlQuote(courseID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o curso")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "curso nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) deleteAcademyCourse(w http.ResponseWriter, r *http.Request, courseID string, session sessionUser) {
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode excluir cursos")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH deleted AS (
			DELETE FROM academy_courses course
			WHERE course.id = %s::uuid
				AND NOT EXISTS (SELECT 1 FROM academy_enrollments enrollment WHERE enrollment.course_id = course.id)
			RETURNING course.id
		) SELECT row_to_json(item) FROM (SELECT id::text AS id FROM deleted) item;
	`, sqlQuote(courseID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel excluir o curso")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		existsPayload, existsErr := a.queryJSON(r.Context(), fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM academy_courses WHERE id = %s::uuid) AS exists`, sqlQuote(courseID)))
		var result struct {
			Exists bool `json:"exists"`
		}
		if existsErr == nil && json.Unmarshal(existsPayload, &result) == nil && result.Exists {
			writeError(w, http.StatusConflict, "este curso possui alunos ou historico; arquive-o para preservar os registros")
			return
		}
		writeError(w, http.StatusNotFound, "curso nao encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) getAcademyCourse(w http.ResponseWriter, r *http.Request, courseID string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item) FROM (
			SELECT course.id::text AS id, course.title, course.description, course.category, COALESCE(course.cover_image_url, '') AS "coverImageUrl",
				course.workload_hours AS "workloadHours", course.status,
				EXISTS (SELECT 1 FROM academy_enrollments enrollment WHERE enrollment.course_id = course.id AND enrollment.user_id = %s::uuid) AS enrolled,
				COALESCE((SELECT certificate.verification_code FROM academy_certificates certificate WHERE certificate.course_id = course.id AND certificate.user_id = %s::uuid), '') AS "certificateCode",
				COALESCE((SELECT json_agg(row_to_json(lesson_item) ORDER BY lesson_item."order") FROM (
					SELECT lesson.id::text AS id, lesson.title, lesson.description, lesson.video_url AS "videoUrl", lesson.video_source AS "videoSource", lesson.duration_seconds AS "durationSeconds", lesson.display_order AS "order",
						lesson.requires_quiz AS "requiresQuiz", CASE WHEN %s THEN lesson.quiz_questions ELSE COALESCE((SELECT jsonb_agg(question - 'correct') FROM jsonb_array_elements(lesson.quiz_questions) question), '[]'::jsonb) END AS "quizQuestions",
						lesson.passing_score AS "passingScore", lesson.max_attempts AS "maxAttempts", lesson.is_published AS "isPublished",
						COALESCE(progress.last_position_seconds, 0) AS "lastPositionSeconds", COALESCE(progress.video_completed, false) AS "videoCompleted", COALESCE(progress.is_completed, false) AS completed,
						NOT EXISTS (SELECT 1 FROM academy_lessons previous LEFT JOIN academy_lesson_progress previous_progress ON previous_progress.lesson_id = previous.id AND previous_progress.user_id = %s::uuid WHERE previous.course_id = course.id AND previous.is_published AND previous.display_order < lesson.display_order AND NOT COALESCE(previous_progress.is_completed, false)) AS unlocked,
						COALESCE((SELECT count(*) FROM academy_assessment_attempts attempt WHERE attempt.lesson_id = lesson.id AND attempt.user_id = %s::uuid), 0)::int AS "attemptCount"
					FROM academy_lessons lesson LEFT JOIN academy_lesson_progress progress ON progress.lesson_id = lesson.id AND progress.user_id = %s::uuid
					WHERE lesson.course_id = course.id AND (lesson.is_published OR %s)
				) lesson_item), '[]'::json) AS lessons
			FROM academy_courses course
			WHERE course.id = %s::uuid AND (course.status = 'published' OR %s)
		) item;
	`, sqlQuote(session.UserID), sqlQuote(session.UserID), sqlBool(session.canManagePlatform()), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlBool(session.canManagePlatform()), sqlQuote(courseID), sqlBool(session.canManagePlatform())))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o curso")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "curso nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) enrollAcademyCourse(w http.ResponseWriter, r *http.Request, courseID string, session sessionUser) {
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH course AS (SELECT id FROM academy_courses WHERE id = %s::uuid AND status = 'published' AND EXISTS (SELECT 1 FROM academy_lessons lesson WHERE lesson.course_id = academy_courses.id AND lesson.is_published)), enrollment AS (
			INSERT INTO academy_enrollments (course_id, user_id) SELECT id, %s::uuid FROM course
			ON CONFLICT (course_id, user_id) DO UPDATE SET last_accessed_at = now()
			RETURNING id
		) SELECT row_to_json(item) FROM (SELECT id::text AS id FROM enrollment) item;
	`, sqlQuote(courseID), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel iniciar o curso")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "curso indisponivel")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) createAcademyLesson(w http.ResponseWriter, r *http.Request, courseID string, session sessionUser) {
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode criar aulas")
		return
	}
	var req academyLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Title, req.Description, req.VideoURL = strings.TrimSpace(req.Title), strings.TrimSpace(req.Description), strings.TrimSpace(req.VideoURL)
	req.VideoSource = strings.ToLower(strings.TrimSpace(req.VideoSource))
	if req.VideoSource == "" {
		req.VideoSource = "youtube"
	}
	if req.VideoSource != "youtube" && req.VideoSource != "upload" {
		writeError(w, http.StatusBadRequest, "origem do video invalida")
		return
	}
	if req.Title == "" || req.VideoURL == "" {
		writeError(w, http.StatusBadRequest, "titulo e video sao obrigatorios")
		return
	}
	if req.VideoSource == "upload" && !strings.Contains(req.VideoURL, "/uploads/academy/") {
		writeError(w, http.StatusBadRequest, "envie o arquivo do video antes de salvar a aula")
		return
	}
	if req.DurationSecond < 0 {
		req.DurationSecond = 0
	}
	req.PassingScore = 75
	if req.MaxAttempts < 0 {
		req.MaxAttempts = 0
	}
	if req.RequiresQuiz && len(req.Questions) == 0 {
		writeError(w, http.StatusBadRequest, "inclua ao menos uma questao para a prova")
		return
	}
	for _, question := range req.Questions {
		if req.RequiresQuiz && (strings.TrimSpace(question.Question) == "" || len(question.Options) < 2 || question.Correct < 0 || question.Correct >= len(question.Options)) {
			writeError(w, http.StatusBadRequest, "revise as questoes, alternativas e respostas corretas")
			return
		}
	}
	isPublished := true
	if req.IsPublished != nil {
		isPublished = *req.IsPublished
	}
	questions, _ := json.Marshal(req.Questions)
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH course AS (SELECT id FROM academy_courses WHERE id = %s::uuid), inserted AS (
			INSERT INTO academy_lessons (course_id, title, description, video_url, video_source, duration_seconds, display_order, requires_quiz, quiz_questions, passing_score, max_attempts, is_published)
			SELECT id, %s, %s, %s, %s, %d, COALESCE((SELECT max(display_order) + 1 FROM academy_lessons WHERE course_id = course.id), 1), %s, %s::jsonb, %d, %d, %s FROM course RETURNING *
		) SELECT row_to_json(item) FROM (SELECT id::text AS id, title FROM inserted) item;
	`, sqlQuote(courseID), sqlQuote(req.Title), sqlQuote(req.Description), sqlQuote(req.VideoURL), sqlQuote(req.VideoSource), req.DurationSecond, sqlBool(req.RequiresQuiz), sqlQuote(string(questions)), req.PassingScore, req.MaxAttempts, sqlBool(isPublished)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel criar a aula")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "curso nao encontrado")
		return
	}
	writeRawJSON(w, http.StatusCreated, payload)
}

func (a *app) handleAcademyLessonByPath(w http.ResponseWriter, r *http.Request) {
	lessonID, action := splitResourcePath(r.URL.Path, "/api/academy/lessons/")
	if lessonID == "" {
		writeError(w, http.StatusNotFound, "aula nao encontrada")
		return
	}
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if action == "progress" && r.Method == http.MethodPut {
		a.updateAcademyProgress(w, r, lessonID, session)
		return
	}
	if action == "assessment" && r.Method == http.MethodPost {
		a.submitAcademyAssessment(w, r, lessonID, session)
		return
	}
	if action == "" && r.Method == http.MethodPut {
		a.updateAcademyLesson(w, r, lessonID, session)
		return
	}
	if action == "" && r.Method == http.MethodDelete {
		a.deleteAcademyLesson(w, r, lessonID, session)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "acao de aula nao permitida")
}

func (a *app) updateAcademyLesson(w http.ResponseWriter, r *http.Request, lessonID string, session sessionUser) {
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode editar aulas")
		return
	}
	var req academyLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.VideoURL = strings.TrimSpace(req.VideoURL)
	req.VideoSource = strings.ToLower(strings.TrimSpace(req.VideoSource))
	if req.VideoSource == "" {
		req.VideoSource = "youtube"
	}
	if req.VideoSource != "youtube" && req.VideoSource != "upload" {
		writeError(w, http.StatusBadRequest, "origem do video invalida")
		return
	}
	if req.Title == "" || req.VideoURL == "" {
		writeError(w, http.StatusBadRequest, "titulo e video sao obrigatorios")
		return
	}
	if req.VideoSource == "upload" && !strings.Contains(req.VideoURL, "/uploads/academy/") {
		writeError(w, http.StatusBadRequest, "envie o arquivo do video antes de salvar a aula")
		return
	}
	if req.DurationSecond < 0 {
		req.DurationSecond = 0
	}
	req.PassingScore = 75
	if req.MaxAttempts < 0 {
		req.MaxAttempts = 0
	}
	if req.RequiresQuiz && len(req.Questions) == 0 {
		writeError(w, http.StatusBadRequest, "inclua ao menos uma questao para a prova")
		return
	}
	for _, question := range req.Questions {
		if req.RequiresQuiz && (strings.TrimSpace(question.Question) == "" || len(question.Options) < 2 || question.Correct < 0 || question.Correct >= len(question.Options)) {
			writeError(w, http.StatusBadRequest, "revise as questoes, alternativas e respostas corretas")
			return
		}
	}
	isPublished := true
	if req.IsPublished != nil {
		isPublished = *req.IsPublished
	}
	questions, _ := json.Marshal(req.Questions)
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH updated AS (
			UPDATE academy_lessons SET title = %s, description = %s, video_url = %s, video_source = %s, duration_seconds = %d,
				requires_quiz = %s, quiz_questions = %s::jsonb, passing_score = 75, max_attempts = %d,
				is_published = %s, updated_at = now()
			WHERE id = %s::uuid RETURNING id, title
		) SELECT row_to_json(item) FROM (SELECT id::text AS id, title FROM updated) item;
	`, sqlQuote(req.Title), sqlQuote(req.Description), sqlQuote(req.VideoURL), sqlQuote(req.VideoSource), req.DurationSecond, sqlBool(req.RequiresQuiz), sqlQuote(string(questions)), req.MaxAttempts, sqlBool(isPublished), sqlQuote(lessonID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar a aula")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "aula nao encontrada")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) deleteAcademyLesson(w http.ResponseWriter, r *http.Request, lessonID string, session sessionUser) {
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode remover aulas")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH removed AS (
			DELETE FROM academy_lessons lesson WHERE lesson.id = %s::uuid
				AND NOT EXISTS (SELECT 1 FROM academy_lesson_progress progress WHERE progress.lesson_id = lesson.id)
			RETURNING id
		), archived AS (
			UPDATE academy_lessons SET is_published = false, updated_at = now()
			WHERE id = %s::uuid AND NOT EXISTS (SELECT 1 FROM removed)
			RETURNING id
		) SELECT row_to_json(item) FROM (
			SELECT id::text AS id, 'deleted'::text AS action FROM removed
			UNION ALL SELECT id::text AS id, 'archived'::text AS action FROM archived
		) item;
	`, sqlQuote(lessonID), sqlQuote(lessonID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel remover a aula")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusNotFound, "aula nao encontrada")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) updateAcademyProgress(w http.ResponseWriter, r *http.Request, lessonID string, session sessionUser) {
	var req academyProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	if req.PositionSeconds < 0 {
		req.PositionSeconds = 0
	}
	if req.WatchedSecondsDelta < 0 {
		req.WatchedSecondsDelta = 0
	}
	if req.WatchedSecondsDelta > 30 {
		req.WatchedSecondsDelta = 30
	}
	if req.DurationSeconds < 0 {
		req.DurationSeconds = 0
	}
	var payload []byte
	var err error
	if req.Ended && req.DurationSeconds > 0 && req.PositionSeconds >= req.DurationSeconds-5 {
		payload, err = a.queryJSON(r.Context(), fmt.Sprintf(`
			WITH lesson AS (
				SELECT lesson.* FROM academy_lessons lesson
				JOIN academy_courses course ON course.id = lesson.course_id
				JOIN academy_enrollments enrollment ON enrollment.course_id = course.id AND enrollment.user_id = %s::uuid
				WHERE lesson.id = %s::uuid AND lesson.is_published AND course.status = 'published'
			), updated AS (
				INSERT INTO academy_lesson_progress (lesson_id, user_id, last_position_seconds, watched_seconds, video_completed, is_completed, completed_at)
				SELECT id, %s::uuid, %d, %d, true, NOT requires_quiz, CASE WHEN NOT requires_quiz THEN now() ELSE NULL END
				FROM lesson
				ON CONFLICT (lesson_id, user_id) DO UPDATE SET
					last_position_seconds = %d,
					watched_seconds = GREATEST(academy_lesson_progress.watched_seconds, %d),
					video_completed = true,
					is_completed = academy_lesson_progress.is_completed OR NOT (SELECT requires_quiz FROM lesson),
					completed_at = CASE
						WHEN academy_lesson_progress.completed_at IS NOT NULL THEN academy_lesson_progress.completed_at
						WHEN NOT (SELECT requires_quiz FROM lesson) THEN now()
						ELSE NULL
					END,
					updated_at = now()
				RETURNING *
			) SELECT row_to_json(item) FROM (
				SELECT id::text AS id, last_position_seconds AS "lastPositionSeconds", watched_seconds AS "watchedSeconds",
					video_completed AS "videoCompleted", is_completed AS completed
				FROM updated
			) item;
		`, sqlQuote(session.UserID), sqlQuote(lessonID), sqlQuote(session.UserID),
			req.DurationSeconds, req.DurationSeconds, req.DurationSeconds, req.DurationSeconds))
	} else {
		payload, err = a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH lesson AS (
			SELECT lesson.* FROM academy_lessons lesson
			JOIN academy_courses course ON course.id = lesson.course_id
			JOIN academy_enrollments enrollment ON enrollment.course_id = course.id AND enrollment.user_id = %s::uuid
			WHERE lesson.id = %s::uuid AND lesson.is_published AND course.status = 'published'
		), updated AS (
			INSERT INTO academy_lesson_progress (lesson_id, user_id, last_position_seconds, watched_seconds, video_completed, is_completed, completed_at)
			SELECT id, %s::uuid,
				CASE WHEN COALESCE(NULLIF(%d, 0), duration_seconds) > 0 THEN LEAST(GREATEST(0, %d), COALESCE(NULLIF(%d, 0), duration_seconds)) ELSE GREATEST(0, %d) END,
				LEAST(COALESCE(NULLIF(%d, 0), duration_seconds), %d),
				%d > 0 AND %d >= CEIL(COALESCE(NULLIF(%d, 0), duration_seconds) * 0.98),
				NOT requires_quiz AND %d > 0 AND %d >= CEIL(COALESCE(NULLIF(%d, 0), duration_seconds) * 0.98),
				CASE WHEN NOT requires_quiz AND %d > 0 AND %d >= CEIL(COALESCE(NULLIF(%d, 0), duration_seconds) * 0.98) THEN now() ELSE NULL END
			FROM lesson
			ON CONFLICT (lesson_id, user_id) DO UPDATE SET
				last_position_seconds = EXCLUDED.last_position_seconds,
				watched_seconds = LEAST(COALESCE(NULLIF(%d, 0), (SELECT duration_seconds FROM lesson)), academy_lesson_progress.watched_seconds + %d),
				video_completed = academy_lesson_progress.video_completed OR (COALESCE(NULLIF(%d, 0), (SELECT duration_seconds FROM lesson)) > 0 AND academy_lesson_progress.watched_seconds + %d >= CEIL(COALESCE(NULLIF(%d, 0), (SELECT duration_seconds FROM lesson)) * 0.98)),
				is_completed = academy_lesson_progress.is_completed OR (NOT (SELECT requires_quiz FROM lesson) AND COALESCE(NULLIF(%d, 0), (SELECT duration_seconds FROM lesson)) > 0 AND academy_lesson_progress.watched_seconds + %d >= CEIL(COALESCE(NULLIF(%d, 0), (SELECT duration_seconds FROM lesson)) * 0.98)),
				completed_at = CASE WHEN academy_lesson_progress.completed_at IS NOT NULL THEN academy_lesson_progress.completed_at WHEN NOT (SELECT requires_quiz FROM lesson) AND COALESCE(NULLIF(%d, 0), (SELECT duration_seconds FROM lesson)) > 0 AND academy_lesson_progress.watched_seconds + %d >= CEIL(COALESCE(NULLIF(%d, 0), (SELECT duration_seconds FROM lesson)) * 0.98) THEN now() ELSE NULL END,
				updated_at = now()
			RETURNING *
		) SELECT row_to_json(item) FROM (SELECT id::text AS id, last_position_seconds AS "lastPositionSeconds", watched_seconds AS "watchedSeconds", video_completed AS "videoCompleted", is_completed AS completed FROM updated) item;
	`, sqlQuote(session.UserID), sqlQuote(lessonID), sqlQuote(session.UserID),
			req.DurationSeconds, req.PositionSeconds, req.DurationSeconds, req.PositionSeconds,
			req.DurationSeconds, req.WatchedSecondsDelta,
			req.DurationSeconds, req.WatchedSecondsDelta, req.DurationSeconds,
			req.DurationSeconds, req.WatchedSecondsDelta, req.DurationSeconds,
			req.DurationSeconds, req.WatchedSecondsDelta, req.DurationSeconds,
			req.DurationSeconds, req.WatchedSecondsDelta,
			req.DurationSeconds, req.WatchedSecondsDelta, req.DurationSeconds,
			req.DurationSeconds, req.WatchedSecondsDelta, req.DurationSeconds,
			req.DurationSeconds, req.WatchedSecondsDelta, req.DurationSeconds))
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel salvar o progresso")
		return
	}
	if strings.TrimSpace(string(payload)) == "null" {
		writeError(w, http.StatusForbidden, "inicie o curso antes de registrar o progresso")
		return
	}
	var progress struct {
		Completed bool `json:"completed"`
	}
	if json.Unmarshal(payload, &progress) == nil && progress.Completed {
		coursePayload, courseErr := a.queryJSON(r.Context(), fmt.Sprintf(`SELECT course_id::text AS "courseId" FROM academy_lessons WHERE id = %s::uuid`, sqlQuote(lessonID)))
		var course struct {
			CourseID string `json:"courseId"`
		}
		if courseErr == nil && json.Unmarshal(coursePayload, &course) == nil && course.CourseID != "" {
			_, _ = a.issueAcademyCertificate(r.Context(), course.CourseID, session.UserID)
		}
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) submitAcademyAssessment(w http.ResponseWriter, r *http.Request, lessonID string, session sessionUser) {
	var req academyAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return
	}
	lessonPayload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT row_to_json(item) FROM (
			SELECT lesson.course_id::text AS "courseId", lesson.requires_quiz AS "requiresQuiz", lesson.quiz_questions AS questions, 75 AS "passingScore", lesson.max_attempts AS "maxAttempts", progress.video_completed AS "videoCompleted",
				(SELECT count(*) FROM academy_assessment_attempts attempt WHERE attempt.lesson_id = lesson.id AND attempt.user_id = %s::uuid)::int AS "attemptCount"
			FROM academy_lessons lesson JOIN academy_enrollments enrollment ON enrollment.course_id = lesson.course_id AND enrollment.user_id = %s::uuid
			JOIN academy_lesson_progress progress ON progress.lesson_id = lesson.id AND progress.user_id = %s::uuid
			WHERE lesson.id = %s::uuid AND lesson.is_published
		) item;
	`, sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(lessonID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar a prova")
		return
	}
	var lesson struct {
		CourseID       string                `json:"courseId"`
		RequiresQuiz   bool                  `json:"requiresQuiz"`
		Questions      []academyQuizQuestion `json:"questions"`
		PassingScore   int                   `json:"passingScore"`
		MaxAttempts    int                   `json:"maxAttempts"`
		AttemptCount   int                   `json:"attemptCount"`
		VideoCompleted bool                  `json:"videoCompleted"`
	}
	if strings.TrimSpace(string(lessonPayload)) == "null" || json.Unmarshal(lessonPayload, &lesson) != nil || !lesson.RequiresQuiz {
		writeError(w, http.StatusNotFound, "prova indisponivel para esta aula")
		return
	}
	if !lesson.VideoCompleted {
		writeError(w, http.StatusForbidden, "assista ao video completo antes de responder o questionario")
		return
	}
	if lesson.MaxAttempts > 0 && lesson.AttemptCount >= lesson.MaxAttempts {
		writeError(w, http.StatusForbidden, "numero maximo de tentativas atingido")
		return
	}
	correct := 0
	wrongQuestionIndexes := make([]int, 0)
	for index, question := range lesson.Questions {
		if index < len(req.Answers) && req.Answers[index] == question.Correct {
			correct++
		} else {
			wrongQuestionIndexes = append(wrongQuestionIndexes, index)
		}
	}
	score := 0.0
	if len(lesson.Questions) > 0 {
		score = float64(correct) * 100 / float64(len(lesson.Questions))
	}
	approved := score >= float64(lesson.PassingScore)
	answers, _ := json.Marshal(req.Answers)
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH attempt AS (
			INSERT INTO academy_assessment_attempts (lesson_id, user_id, answers, score, approved) VALUES (%s::uuid, %s::uuid, %s::jsonb, %.2f, %s) RETURNING *
		), progress AS (
			INSERT INTO academy_lesson_progress (lesson_id, user_id, is_completed, completed_at) SELECT %s::uuid, %s::uuid, %s, CASE WHEN %s THEN now() ELSE NULL END
			ON CONFLICT (lesson_id, user_id) DO UPDATE SET is_completed = academy_lesson_progress.is_completed OR EXCLUDED.is_completed, completed_at = COALESCE(academy_lesson_progress.completed_at, EXCLUDED.completed_at), updated_at = now()
			RETURNING *
		) SELECT row_to_json(item) FROM (SELECT attempt.score, attempt.approved, progress.is_completed AS completed FROM attempt CROSS JOIN progress) item;
	`, sqlQuote(lessonID), sqlQuote(session.UserID), sqlQuote(string(answers)), score, sqlBool(approved), sqlQuote(lessonID), sqlQuote(session.UserID), sqlBool(approved), sqlBool(approved)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel registrar a prova")
		return
	}
	if approved {
		_, _ = a.issueAcademyCertificate(r.Context(), lesson.CourseID, session.UserID)
	}
	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel concluir a prova")
		return
	}
	result["wrongQuestionIndexes"] = wrongQuestionIndexes
	writeJSON(w, http.StatusOK, result)
}

func (a *app) issueAcademyCertificate(ctx context.Context, courseID, userID string) ([]byte, error) {
	return a.queryJSON(ctx, fmt.Sprintf(`
		WITH completed AS (
			SELECT NOT EXISTS (
				SELECT 1 FROM academy_lessons lesson LEFT JOIN academy_lesson_progress progress ON progress.lesson_id = lesson.id AND progress.user_id = %s::uuid
				WHERE lesson.course_id = %s::uuid AND lesson.is_published AND NOT COALESCE(progress.is_completed, false)
			) AS ready
		), certificate AS (
			INSERT INTO academy_certificates (course_id, user_id, verification_code)
			SELECT %s::uuid, %s::uuid, upper(substr(replace(gen_random_uuid()::text, '-', ''), 1, 12)) FROM completed WHERE ready
			ON CONFLICT (course_id, user_id) DO UPDATE SET issued_at = academy_certificates.issued_at
			RETURNING verification_code
		), completed_enrollment AS (
			UPDATE academy_enrollments SET completed_at = COALESCE(completed_at, now()), last_accessed_at = now()
			WHERE course_id = %s::uuid AND user_id = %s::uuid AND EXISTS (SELECT 1 FROM certificate)
			RETURNING id
		) SELECT COALESCE(json_agg(row_to_json(item)), '[]'::json) FROM (SELECT verification_code AS "verificationCode" FROM certificate) item;
	`, sqlQuote(userID), sqlQuote(courseID), sqlQuote(courseID), sqlQuote(userID), sqlQuote(courseID), sqlQuote(userID)))
}

func (a *app) handleAcademyMyLearning(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."lastAccessedAt" DESC), '[]'::json) FROM (
			SELECT course.id::text AS id, course.title, course.category, COALESCE(course.cover_image_url, '') AS "coverImageUrl", enrollment.last_accessed_at AS "lastAccessedAt",
				COALESCE((SELECT round(100.0 * count(*) FILTER (WHERE progress.is_completed) / NULLIF(count(*),0),0) FROM academy_lessons lesson LEFT JOIN academy_lesson_progress progress ON progress.lesson_id=lesson.id AND progress.user_id=%s::uuid WHERE lesson.course_id=course.id AND lesson.is_published),0)::int AS "progressPercent",
				COALESCE((SELECT certificate.verification_code FROM academy_certificates certificate WHERE certificate.course_id=course.id AND certificate.user_id=%s::uuid),'') AS "certificateCode"
			FROM academy_enrollments enrollment JOIN academy_courses course ON course.id=enrollment.course_id WHERE enrollment.user_id=%s::uuid
		) item;
	`, sqlQuote(session.UserID), sqlQuote(session.UserID), sqlQuote(session.UserID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar seus cursos")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}
