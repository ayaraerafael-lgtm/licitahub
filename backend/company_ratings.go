package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type companyRatingRoundRequest struct {
	Title    string `json:"title"`
	ClosesAt string `json:"closesAt"`
}

type companyRatingAllocationRequest struct {
	Stars int `json:"stars"`
}

func (a *app) ensureCompanyRatingMigrations(ctx context.Context) error {
	_, err := a.runPSQL(ctx, `
		CREATE TABLE IF NOT EXISTS company_rating_rounds (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			title varchar(180) NOT NULL,
			status varchar(20) NOT NULL DEFAULT 'open',
			company_count integer NOT NULL,
			star_budget integer NOT NULL,
			created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
			opened_at timestamptz NOT NULL DEFAULT now(),
			closes_at timestamptz NOT NULL DEFAULT (now() + interval '7 days'),
			closed_at timestamptz,
			created_at timestamptz NOT NULL DEFAULT now(),
			CONSTRAINT company_rating_rounds_status_chk CHECK (status IN ('open', 'closed')),
			CONSTRAINT company_rating_rounds_budget_chk CHECK (company_count >= 2 AND star_budget >= 1)
		);
		CREATE UNIQUE INDEX IF NOT EXISTS company_rating_one_open_round
			ON company_rating_rounds ((status)) WHERE status = 'open';
		ALTER TABLE company_rating_rounds ADD COLUMN IF NOT EXISTS closes_at timestamptz;
		UPDATE company_rating_rounds
		SET closes_at = COALESCE(closes_at, closed_at, opened_at + interval '7 days')
		WHERE closes_at IS NULL;
		ALTER TABLE company_rating_rounds
			ALTER COLUMN closes_at SET DEFAULT (now() + interval '7 days');
		ALTER TABLE company_rating_rounds ALTER COLUMN closes_at SET NOT NULL;
		ALTER TABLE company_rating_rounds DROP CONSTRAINT IF EXISTS company_rating_rounds_deadline_chk;
		ALTER TABLE company_rating_rounds
			ADD CONSTRAINT company_rating_rounds_deadline_chk CHECK (closes_at > opened_at);

		CREATE TABLE IF NOT EXISTS company_rating_round_companies (
			round_id uuid NOT NULL REFERENCES company_rating_rounds(id) ON DELETE CASCADE,
			company_id uuid NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
			star_budget integer NOT NULL,
			submitted_at timestamptz,
			submitted_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
			included_at timestamptz NOT NULL DEFAULT now(),
			PRIMARY KEY (round_id, company_id),
			CONSTRAINT company_rating_round_company_budget_chk CHECK (star_budget >= 1)
		);

		CREATE TABLE IF NOT EXISTS company_rating_allocations (
			round_id uuid NOT NULL REFERENCES company_rating_rounds(id) ON DELETE CASCADE,
			evaluator_company_id uuid NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
			target_company_id uuid NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
			stars smallint NOT NULL,
			updated_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now(),
			PRIMARY KEY (round_id, evaluator_company_id, target_company_id),
			CONSTRAINT company_rating_allocations_stars_chk CHECK (stars >= 1),
			CONSTRAINT company_rating_allocations_distinct_chk CHECK (evaluator_company_id <> target_company_id),
			FOREIGN KEY (round_id, evaluator_company_id)
				REFERENCES company_rating_round_companies(round_id, company_id) ON DELETE CASCADE,
			FOREIGN KEY (round_id, target_company_id)
				REFERENCES company_rating_round_companies(round_id, company_id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_company_rating_allocations_target
			ON company_rating_allocations(round_id, target_company_id);
		CREATE INDEX IF NOT EXISTS idx_company_rating_round_companies_submission
			ON company_rating_round_companies(round_id, submitted_at);
		ALTER TABLE company_rating_allocations
			DROP CONSTRAINT IF EXISTS company_rating_allocations_stars_chk;
		ALTER TABLE company_rating_allocations
			ADD CONSTRAINT company_rating_allocations_stars_chk CHECK (stars >= 1);
		UPDATE company_rating_rounds r
		SET status = 'closed', closed_at = COALESCE(r.closed_at, now())
		WHERE r.status = 'open'
		  AND (
			r.closes_at <= now()
			OR NOT EXISTS (
				SELECT 1
				FROM company_rating_round_companies pending
				WHERE pending.round_id = r.id AND pending.submitted_at IS NULL
			)
		  );
	`)
	return err
}

func (a *app) closeDueCompanyRatingRounds(ctx context.Context) error {
	_, err := a.runPSQL(ctx, `
		UPDATE company_rating_rounds r
		SET status = 'closed', closed_at = COALESCE(r.closed_at, now())
		WHERE r.status = 'open'
		  AND (
			r.closes_at <= now()
			OR NOT EXISTS (
				SELECT 1
				FROM company_rating_round_companies pending
				WHERE pending.round_id = r.id AND pending.submitted_at IS NULL
			)
		  );
	`)
	return err
}

func (a *app) runCompanyRatingDeadlineWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.closeDueCompanyRatingRounds(ctx); err != nil {
				log.Printf("event=company_rating_deadline_error error=%q", err.Error())
			}
		}
	}
}

func (a *app) handleCompanyRatings(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "esta funcionalidade e exclusiva das empresas associadas")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if err := a.closeDueCompanyRatingRounds(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o prazo da rodada")
		return
	}

	canManage := "false"
	if session.canManageCompany() {
		canManage = "true"
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH open_round AS (
			SELECT r.*
			FROM company_rating_rounds r
			WHERE r.status = 'open'
			ORDER BY r.opened_at DESC
			LIMIT 1
		),
		membership AS (
			SELECT rc.*
			FROM company_rating_round_companies rc
			JOIN open_round r ON r.id = rc.round_id
			WHERE rc.company_id = %s::uuid
		)
		SELECT json_build_object(
			'round', (
				SELECT json_build_object(
					'id', r.id::text,
					'title', r.title,
					'status', r.status,
					'companyCount', r.company_count,
					'starBudget', r.star_budget,
					'openedAt', r.opened_at,
					'closesAt', r.closes_at
				) FROM open_round r
			),
			'included', EXISTS (SELECT 1 FROM membership),
			'canManage', %s,
			'submitted', COALESCE((SELECT submitted_at IS NOT NULL FROM membership), false),
			'usedStars', COALESCE((
				SELECT sum(a.stars)::int
				FROM company_rating_allocations a
				JOIN open_round r ON r.id = a.round_id
				WHERE a.evaluator_company_id = %s::uuid
			), 0),
			'companies', COALESCE((
				SELECT json_agg(json_build_object(
					'id', c.id::text,
					'name', c.trade_name,
					'logoUrl', COALESCE(media.file_url, ''),
					'isOwn', c.id = %s::uuid,
					'allocatedStars', COALESCE(allocation.stars, 0)
				) ORDER BY c.trade_name)
				FROM open_round r
				JOIN company_rating_round_companies rc ON rc.round_id = r.id
				JOIN companies c ON c.id = rc.company_id
				LEFT JOIN company_profiles profile ON profile.company_id = c.id
				LEFT JOIN media_files media ON media.id = profile.logo_media_id
				LEFT JOIN company_rating_allocations allocation
					ON allocation.round_id = r.id
					AND allocation.evaluator_company_id = %s::uuid
					AND allocation.target_company_id = c.id
			), '[]'::json)
		);
	`, sqlQuote(session.CompanyID), canManage, sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar a rodada de avaliacao")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleCompanyRatingAllocation(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManageCompany() {
		writeError(w, http.StatusForbidden, "somente o administrador da empresa pode distribuir estrelas")
		return
	}
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if err := a.closeDueCompanyRatingRounds(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o prazo da rodada")
		return
	}
	targetCompanyID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/company-ratings/allocations/"), "/")
	if targetCompanyID == "" || targetCompanyID == session.CompanyID {
		writeError(w, http.StatusBadRequest, "selecione outra empresa para avaliar")
		return
	}
	var req companyRatingAllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Stars < 0 {
		writeError(w, http.StatusBadRequest, "informe uma quantidade valida de estrelas")
		return
	}

	mutation := fmt.Sprintf(`
		INSERT INTO company_rating_allocations (
			round_id, evaluator_company_id, target_company_id, stars, updated_by_user_id
		)
		SELECT r.id, %s::uuid, %s::uuid, %d, %s::uuid
		FROM company_rating_rounds r
		JOIN company_rating_round_companies evaluator
			ON evaluator.round_id = r.id AND evaluator.company_id = %s::uuid
		JOIN company_rating_round_companies target
			ON target.round_id = r.id AND target.company_id = %s::uuid
		WHERE r.status = 'open'
		  AND evaluator.submitted_at IS NULL
		  AND (
			COALESCE((
				SELECT sum(existing.stars)
				FROM company_rating_allocations existing
				WHERE existing.round_id = r.id
				  AND existing.evaluator_company_id = %s::uuid
				  AND existing.target_company_id <> %s::uuid
			), 0) + %d
		  ) <= evaluator.star_budget
		ON CONFLICT (round_id, evaluator_company_id, target_company_id)
		DO UPDATE SET stars = EXCLUDED.stars, updated_by_user_id = EXCLUDED.updated_by_user_id, updated_at = now()
		RETURNING stars
	`, sqlQuote(session.CompanyID), sqlQuote(targetCompanyID), req.Stars, sqlQuote(session.UserID),
		sqlQuote(session.CompanyID), sqlQuote(targetCompanyID), sqlQuote(session.CompanyID), sqlQuote(targetCompanyID), req.Stars)
	if req.Stars == 0 {
		mutation = fmt.Sprintf(`
			DELETE FROM company_rating_allocations allocation
			USING company_rating_rounds r, company_rating_round_companies evaluator
			WHERE allocation.round_id = r.id
			  AND evaluator.round_id = r.id
			  AND evaluator.company_id = %s::uuid
			  AND evaluator.submitted_at IS NULL
			  AND r.status = 'open'
			  AND allocation.evaluator_company_id = %s::uuid
			  AND allocation.target_company_id = %s::uuid
			RETURNING 0 AS stars
		`, sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(targetCompanyID))
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH changed AS (%s)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM changed) THEN json_build_object(
				'ok', true,
				'usedStars', COALESCE((
					SELECT sum(a.stars)::int
					FROM company_rating_allocations a
					JOIN company_rating_rounds r ON r.id = a.round_id AND r.status = 'open'
					WHERE a.evaluator_company_id = %s::uuid
				), 0)
			)
			ELSE json_build_object('ok', false)
		END;
	`, mutation, sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel salvar esta avaliacao")
		return
	}
	if strings.Contains(string(payload), `"ok":false`) || strings.Contains(string(payload), `"ok": false`) {
		writeError(w, http.StatusConflict, "a avaliacao nao pode ser salva; confira o saldo de estrelas e a situacao da rodada")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleCompanyRatingSubmit(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManageCompany() {
		writeError(w, http.StatusForbidden, "somente o administrador da empresa pode concluir a avaliacao")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if err := a.closeDueCompanyRatingRounds(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o prazo da rodada")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH evaluation_round AS (
			SELECT r.id
			FROM company_rating_rounds r
			JOIN company_rating_round_companies rc
				ON rc.round_id = r.id AND rc.company_id = %s::uuid
			WHERE r.status = 'open' OR rc.submitted_at IS NOT NULL
			ORDER BY (r.status = 'open') DESC, r.opened_at DESC
			LIMIT 1
		),
		submitted AS (
			UPDATE company_rating_round_companies rc
			SET submitted_at = now(), submitted_by_user_id = %s::uuid
			FROM company_rating_rounds r
			WHERE rc.round_id = r.id
			  AND rc.round_id IN (SELECT id FROM evaluation_round)
			  AND rc.company_id = %s::uuid
			  AND rc.submitted_at IS NULL
			  AND r.status = 'open'
			  AND rc.star_budget = (
				SELECT COALESCE(sum(a.stars), 0)
				FROM company_rating_allocations a
				WHERE a.round_id = r.id AND a.evaluator_company_id = rc.company_id
			  )
			RETURNING rc.round_id
		),
		closed AS (
			UPDATE company_rating_rounds r
			SET status = 'closed', closed_at = now()
			WHERE r.id IN (SELECT round_id FROM submitted)
			  AND NOT EXISTS (
				SELECT 1 FROM company_rating_round_companies pending
				WHERE pending.round_id = r.id
				  AND pending.submitted_at IS NULL
				  AND pending.company_id <> %s::uuid
			  )
			RETURNING r.id
		)
		SELECT json_build_object(
			'submitted', EXISTS (SELECT 1 FROM submitted),
			'alreadySubmitted', EXISTS (
				SELECT 1
				FROM company_rating_round_companies rc
				JOIN evaluation_round selected ON selected.id = rc.round_id
				WHERE rc.company_id = %s::uuid
				  AND rc.submitted_at IS NOT NULL
			),
			'roundClosed', EXISTS (SELECT 1 FROM closed)
		);
	`, sqlQuote(session.CompanyID), sqlQuote(session.UserID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel concluir a avaliacao")
		return
	}
	var result struct {
		Submitted        bool `json:"submitted"`
		AlreadySubmitted bool `json:"alreadySubmitted"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel confirmar a conclusao da avaliacao")
		return
	}
	if !result.Submitted && !result.AlreadySubmitted {
		writeError(w, http.StatusConflict, "distribua todas as estrelas antes de concluir")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleCompanyRatingResults(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if session.CompanyID == "" {
		writeError(w, http.StatusForbidden, "esta funcionalidade e exclusiva das empresas associadas")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
		return
	}
	if err := a.closeDueCompanyRatingRounds(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o prazo da rodada")
		return
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH round_scores AS (
			SELECT r.id, r.title, r.status, r.star_budget, r.company_count, r.opened_at, r.closes_at, r.closed_at,
				self_rc.submitted_at, target_rc.company_id,
				COALESCE((
					SELECT sum(a.stars)::int
					FROM company_rating_allocations a
					JOIN company_rating_round_companies voter
					  ON voter.round_id = a.round_id
					 AND voter.company_id = a.evaluator_company_id
					 AND voter.submitted_at IS NOT NULL
					WHERE a.round_id = r.id AND a.target_company_id = target_rc.company_id
				), 0) AS received,
				(
					SELECT count(*)::int
					FROM company_rating_round_companies submitted_voter
					WHERE submitted_voter.round_id = r.id AND submitted_voter.submitted_at IS NOT NULL
				) AS submitted_voters
			FROM company_rating_rounds r
			JOIN company_rating_round_companies target_rc ON target_rc.round_id = r.id
			JOIN company_rating_round_companies self_rc
			  ON self_rc.round_id = r.id AND self_rc.company_id = %s::uuid
		),
		ranked_scores AS (
			SELECT *,
				rank() OVER (PARTITION BY id ORDER BY received DESC)::int AS position,
				(submitted_voters * star_budget::numeric / NULLIF(company_count, 0)) AS market_average
			FROM round_scores
		),
		series AS (
			SELECT id, title, status, star_budget, opened_at, closes_at, closed_at, submitted_at, received,
				position, round(market_average, 1) AS market_average,
				round(received::numeric * 100 / NULLIF(market_average, 0), 1) AS relative_index
			FROM ranked_scores
			WHERE company_id = %s::uuid
			  AND (status = 'closed' OR submitted_at IS NOT NULL)
		),
		latest AS (
			SELECT * FROM series ORDER BY opened_at DESC LIMIT 1
		),
		history AS (
			SELECT
				COALESCE(round(avg(relative_index), 1), 0) AS relative_index,
				count(*)::int AS session_count
			FROM series WHERE status = 'closed'
		)
		SELECT json_build_object(
			'latest', (SELECT json_build_object(
				'id', id::text, 'title', title, 'status', status, 'starBudget', star_budget,
				'receivedStars', received, 'relativeIndex', relative_index,
				'marketAverage', market_average, 'position', position,
				'openedAt', opened_at, 'closesAt', closes_at, 'closedAt', closed_at
			) FROM latest),
			'history', (SELECT json_build_object(
				'relativeIndex', COALESCE(relative_index, 0),
				'sessionCount', session_count
			) FROM history),
			'series', COALESCE((
				SELECT json_agg(json_build_object(
					'id', id::text, 'title', title, 'status', status, 'starBudget', star_budget,
					'receivedStars', received, 'relativeIndex', relative_index,
					'marketAverage', market_average, 'position', position,
					'openedAt', opened_at, 'closesAt', closes_at, 'closedAt', closed_at
				) ORDER BY opened_at)
				FROM series
			), '[]'::json)
		);
	`, sqlQuote(session.CompanyID), sqlQuote(session.CompanyID)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar os resultados")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func (a *app) handleAdminCompanyRatingRounds(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode controlar as rodadas")
		return
	}
	if err := a.closeDueCompanyRatingRounds(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o prazo das rodadas")
		return
	}
	switch r.Method {
	case http.MethodGet:
		payload, err := a.queryJSON(r.Context(), `
			SELECT COALESCE(json_agg(row_to_json(item) ORDER BY item."openedAt" DESC), '[]'::json)
			FROM (
				SELECT r.id::text AS id, r.title, r.status, r.company_count AS "companyCount",
					r.star_budget AS "starBudget", r.opened_at AS "openedAt", r.closes_at AS "closesAt", r.closed_at AS "closedAt",
					count(rc.*) FILTER (WHERE rc.submitted_at IS NOT NULL)::int AS "submittedCount",
					count(rc.*) FILTER (WHERE rc.submitted_at IS NULL)::int AS "pendingCount"
				FROM company_rating_rounds r
				LEFT JOIN company_rating_round_companies rc ON rc.round_id = r.id
				GROUP BY r.id
			) item;
		`)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel carregar as rodadas")
			return
		}
		writeRawJSON(w, http.StatusOK, payload)
	case http.MethodPost:
		var req companyRatingRoundRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "json invalido")
			return
		}
		req.Title = strings.TrimSpace(req.Title)
		req.ClosesAt = strings.TrimSpace(req.ClosesAt)
		if req.Title == "" {
			writeError(w, http.StatusBadRequest, "informe o nome da rodada")
			return
		}
		closesAt, err := time.Parse(time.RFC3339, req.ClosesAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "informe um prazo valido para a rodada")
			return
		}
		if !closesAt.After(time.Now().Add(time.Minute)) {
			writeError(w, http.StatusBadRequest, "o prazo da rodada deve ser posterior ao horario atual")
			return
		}
		payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
			WITH eligible AS (
				SELECT c.id
				FROM companies c
				WHERE c.status = 'active' AND c.deleted_at IS NULL
			),
			created AS (
				INSERT INTO company_rating_rounds (
					title, company_count, star_budget, created_by_user_id, closes_at
				)
				SELECT %s, count(*)::int, CEIL(count(*) * 0.30)::int, %s::uuid, %s::timestamptz
				FROM eligible
				HAVING count(*) >= 2
				   AND NOT EXISTS (SELECT 1 FROM company_rating_rounds WHERE status = 'open')
				RETURNING *
			),
			snapshot AS (
				INSERT INTO company_rating_round_companies (round_id, company_id, star_budget)
				SELECT created.id, eligible.id, created.star_budget
				FROM created CROSS JOIN eligible
				RETURNING round_id, company_id
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
					'Nova rodada de avaliacao aberta',
					'A rodada ' || created.title || ' esta disponivel ate ' ||
						to_char(created.closes_at AT TIME ZONE 'America/Sao_Paulo', 'DD/MM/YYYY HH24:MI') ||
						'. Acesse a avaliacao de parcerias para participar.',
					'company-rating',
					'company_rating_round',
					created.id
				FROM created
				JOIN snapshot ON snapshot.round_id = created.id
				JOIN users u ON u.company_id = snapshot.company_id
				WHERE u.status = 'active' AND u.deleted_at IS NULL
				RETURNING id
			)
			SELECT COALESCE((
				SELECT json_build_object(
					'id', id::text, 'title', title, 'status', status,
					'companyCount', company_count, 'starBudget', star_budget, 'openedAt', opened_at,
					'closesAt', closes_at,
					'notificationCount', (SELECT count(*) FROM created_notifications)
				) FROM created
			), 'null'::json);
		`, sqlQuote(req.Title), sqlQuote(session.UserID), sqlQuote(closesAt.UTC().Format(time.RFC3339))))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel abrir a rodada")
			return
		}
		if string(payload) == "null" {
			writeError(w, http.StatusConflict, "ja existe uma rodada aberta ou faltam empresas ativas")
			return
		}
		writeRawJSON(w, http.StatusCreated, payload)
	case http.MethodDelete:
		roundID := strings.TrimSpace(r.URL.Query().Get("roundId"))
		if roundID == "" {
			writeError(w, http.StatusBadRequest, "informe a rodada que sera excluida")
			return
		}
		payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
			WITH target_round AS (
				SELECT id, title
				FROM company_rating_rounds
				WHERE id::text = %s
				  AND status = 'closed'
			),
			deleted_notifications AS (
				DELETE FROM notifications notification
				USING target_round target
				WHERE notification.related_entity_type = 'company_rating_round'
				  AND notification.related_entity_id = target.id
				RETURNING notification.id
			),
			deleted_round AS (
				DELETE FROM company_rating_rounds round_item
				USING target_round target
				WHERE round_item.id = target.id
				RETURNING round_item.id, round_item.title
			)
			SELECT COALESCE((
				SELECT json_build_object(
					'id', id::text,
					'title', title,
					'deleted', true,
					'deletedNotifications', (SELECT count(*) FROM deleted_notifications)
				)
				FROM deleted_round
			), 'null'::json);
		`, sqlQuote(roundID)))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "nao foi possivel excluir a rodada")
			return
		}
		if strings.TrimSpace(string(payload)) == "null" {
			writeError(w, http.StatusConflict, "somente uma rodada encerrada pode ser excluida")
			return
		}
		log.Printf("event=company_rating_round_deleted round_id=%s user_id=%s", roundID, session.UserID)
		writeRawJSON(w, http.StatusOK, payload)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metodo nao permitido")
	}
}

func (a *app) handleAdminCompanyRatingResults(w http.ResponseWriter, r *http.Request) {
	session, ok := a.currentSessionUser(w, r)
	if !ok {
		return
	}
	if !session.canManagePlatform() {
		writeError(w, http.StatusForbidden, "somente administrador da plataforma pode consultar este resultado")
		return
	}
	if err := a.closeDueCompanyRatingRounds(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel atualizar o prazo das rodadas")
		return
	}
	roundID := strings.TrimSpace(r.URL.Query().Get("roundId"))
	roundFilter := ""
	if roundID != "" {
		roundFilter = fmt.Sprintf("WHERE r.id = %s::uuid", sqlQuote(roundID))
	}
	payload, err := a.queryJSON(r.Context(), fmt.Sprintf(`
		WITH selected_round AS (
			SELECT r.* FROM company_rating_rounds r
			%s
			ORDER BY r.opened_at DESC LIMIT 1
		),
		participation AS (
			SELECT c.id, c.trade_name, rc.submitted_at
			FROM selected_round r
			JOIN company_rating_round_companies rc ON rc.round_id = r.id
			JOIN companies c ON c.id = rc.company_id
		),
		session_scores AS (
			SELECT c.id, c.trade_name,
				COALESCE(sum(a.stars), 0)::int AS stars,
				(
					(SELECT count(*) FROM company_rating_round_companies submitted
					 WHERE submitted.round_id = r.id AND submitted.submitted_at IS NOT NULL)
					* r.star_budget::numeric / NULLIF(r.company_count, 0)
				) AS market_average
			FROM selected_round r
			JOIN company_rating_round_companies rc ON rc.round_id = r.id
			JOIN companies c ON c.id = rc.company_id
			LEFT JOIN company_rating_allocations a ON a.round_id = r.id AND a.target_company_id = c.id
			WHERE r.status = 'closed'
			GROUP BY c.id, c.trade_name, r.id, r.star_budget, r.company_count
		),
		session_ranking AS (
			SELECT id, trade_name, stars,
				COALESCE(round(stars::numeric * 100 / NULLIF(market_average, 0), 1), 0) AS relative_index
			FROM session_scores
		),
		closed_round_scores AS (
			SELECT c.id, c.trade_name,
				r.id AS round_id,
				COALESCE(sum(a.stars), 0)::numeric AS stars,
				(
					(SELECT count(*) FROM company_rating_round_companies submitted
					 WHERE submitted.round_id = r.id AND submitted.submitted_at IS NOT NULL)
					* r.star_budget::numeric / NULLIF(r.company_count, 0)
				) AS market_average
			FROM companies c
			JOIN company_rating_round_companies rc ON rc.company_id = c.id
			JOIN company_rating_rounds r ON r.id = rc.round_id AND r.status = 'closed'
			LEFT JOIN company_rating_allocations a ON a.round_id = r.id AND a.target_company_id = c.id
			GROUP BY c.id, c.trade_name, r.id, r.star_budget, r.company_count
		),
		closed_round_indices AS (
			SELECT id, trade_name, round(stars, 1) AS stars,
				COALESCE(round(stars * 100 / NULLIF(market_average, 0), 1), 0) AS relative_index
			FROM closed_round_scores
		),
		historical AS (
			SELECT id, trade_name,
				round(avg(stars), 1) AS stars,
				round(avg(relative_index), 1) AS relative_index,
				count(*)::int AS session_count
			FROM closed_round_indices
			GROUP BY id, trade_name
		)
		SELECT json_build_object(
			'round', (SELECT json_build_object(
				'id', id::text, 'title', title, 'status', status, 'companyCount', company_count,
				'starBudget', star_budget, 'openedAt', opened_at, 'closesAt', closes_at, 'closedAt', closed_at
			) FROM selected_round),
			'participation', COALESCE((
				SELECT json_agg(json_build_object(
					'id', id::text, 'name', trade_name, 'submitted', submitted_at IS NOT NULL,
					'submittedAt', submitted_at
				) ORDER BY submitted_at IS NULL DESC, trade_name)
				FROM participation
			), '[]'::json),
			'sessionRanking', COALESCE((
				SELECT json_agg(json_build_object(
					'id', id::text, 'name', trade_name, 'stars', stars, 'relativeIndex', relative_index
				) ORDER BY relative_index DESC, trade_name)
				FROM session_ranking
			), '[]'::json),
			'historicalRanking', COALESCE((
				SELECT json_agg(json_build_object(
					'id', id::text, 'name', trade_name, 'stars', stars, 'relativeIndex', relative_index,
					'sessionCount', session_count
				) ORDER BY relative_index DESC, trade_name)
				FROM historical
			), '[]'::json)
		);
	`, roundFilter))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "nao foi possivel carregar o acompanhamento da rodada")
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}
