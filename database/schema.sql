-- LicitaHub - PostgreSQL schema
-- MVP database foundation for:
-- access/admin, companies, Radar LicitaHub, community, tenders,
-- partnership showcase, match/consortium, notifications, media and audit.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =========================================================
-- Access and administration
-- =========================================================

CREATE TABLE IF NOT EXISTS companies (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  trade_name varchar(180) NOT NULL,
  cnpj varchar(18) NOT NULL,
  status varchar(40) NOT NULL DEFAULT 'invited',
  main_contact_name varchar(180),
  main_contact_email varchar(255),
  main_contact_phone varchar(40),
  state varchar(2),
  city varchar(120),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT companies_status_chk CHECK (status IN (
    'invited', 'pending_review', 'active', 'blocked', 'rejected', 'inactive'
  )),
  CONSTRAINT companies_trade_name_uk UNIQUE (trade_name),
  CONSTRAINT companies_cnpj_uk UNIQUE (cnpj)
);

CREATE TABLE IF NOT EXISTS access_profiles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  key varchar(80) NOT NULL UNIQUE,
  name varchar(120) NOT NULL,
  description text,
  is_system boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid REFERENCES companies(id),
  access_profile_id uuid REFERENCES access_profiles(id),
  full_name varchar(180) NOT NULL,
  email varchar(255) NOT NULL,
  phone varchar(40),
  job_title varchar(160),
  password_hash text,
  profile_photo_media_id uuid,
  status varchar(40) NOT NULL DEFAULT 'pending_invite',
  last_login_at timestamptz,
  blocked_at timestamptz,
  removed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT users_status_chk CHECK (status IN (
    'pending_invite', 'active', 'blocked', 'inactive', 'removed'
  ))
);

-- =========================================================
-- Anonymous company partnership perception rounds
-- =========================================================

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
  CONSTRAINT company_rating_rounds_deadline_chk CHECK (closes_at > opened_at),
  CONSTRAINT company_rating_rounds_budget_chk CHECK (company_count >= 2 AND star_budget >= 1)
);

CREATE UNIQUE INDEX IF NOT EXISTS company_rating_one_open_round
  ON company_rating_rounds ((status)) WHERE status = 'open';

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

CREATE TABLE IF NOT EXISTS company_invitations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid REFERENCES companies(id),
  trade_name varchar(180) NOT NULL,
  cnpj varchar(18) NOT NULL,
  contact_name varchar(180) NOT NULL,
  email varchar(255) NOT NULL,
  phone varchar(40) NOT NULL,
  state varchar(2),
  internal_note text,
  status varchar(40) NOT NULL DEFAULT 'sent',
  invited_by_user_id uuid REFERENCES users(id),
  invitation_token varchar(255) NOT NULL UNIQUE,
  sent_at timestamptz,
  accepted_at timestamptz,
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT company_invitations_trade_name_uk UNIQUE (trade_name),
  CONSTRAINT company_invitations_cnpj_uk UNIQUE (cnpj),
  CONSTRAINT company_invitations_status_chk CHECK (status IN (
    'sent', 'pending_review', 'accepted', 'expired', 'cancelled', 'rejected'
  ))
);

CREATE TABLE IF NOT EXISTS auth_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash varchar(128) NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash varchar(128) NOT NULL UNIQUE,
  token_value text,
  expires_at timestamptz NOT NULL,
  used_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS company_reviews (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL REFERENCES companies(id),
  reviewed_by_user_id uuid REFERENCES users(id),
  status varchar(40) NOT NULL,
  adjustment_request text,
  review_note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT company_reviews_status_chk CHECK (status IN (
    'approved', 'adjustment_requested', 'resubmitted', 'rejected'
  ))
);

-- =========================================================
-- Media and files
-- =========================================================

CREATE TABLE IF NOT EXISTS media_files (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid REFERENCES companies(id),
  uploaded_by_user_id uuid REFERENCES users(id),
  media_type varchar(60) NOT NULL,
  file_name varchar(255),
  file_url text NOT NULL,
  mime_type varchar(120),
  file_size bigint,
  source varchar(40) NOT NULL DEFAULT 'upload',
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT media_files_source_chk CHECK (source IN (
    'upload', 'google_drive', 'external_link'
  ))
);

-- =========================================================
-- Technical certificate archive
-- =========================================================

CREATE TABLE IF NOT EXISTS technical_professionals (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  full_name varchar(255) NOT NULL,
  formation varchar(255),
  complementary_education text,
  professional_registration varchar(180),
  role_title varchar(180),
  phone varchar(30),
  email varchar(255),
  state varchar(2),
  availability_status varchar(30) NOT NULL DEFAULT 'available',
  profile_summary text,
  status varchar(30) NOT NULL DEFAULT 'active',
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT technical_professionals_availability_chk CHECK (availability_status IN ('available', 'limited', 'unavailable')),
  CONSTRAINT technical_professionals_status_chk CHECK (status IN ('active', 'archived'))
);

CREATE TABLE IF NOT EXISTS technical_professional_educations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  technical_professional_id uuid NOT NULL REFERENCES technical_professionals(id) ON DELETE CASCADE,
  education_level varchar(40) NOT NULL,
  course_name varchar(255) NOT NULL,
  institution varchar(255),
  display_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT technical_professional_educations_level_chk CHECK (education_level IN ('graduation', 'specialization', 'mba', 'masters', 'doctorate', 'postdoctorate', 'extension', 'certification', 'other'))
);

CREATE TABLE IF NOT EXISTS technical_certificates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  certificate_number varchar(160),
  issuer_name varchar(255),
  issuer_document varchar(24),
  contracted_name varchar(255),
  contract_number varchar(160),
  object text NOT NULL,
  service_description text,
  state varchar(2),
  city varchar(120),
  execution_start date,
  execution_end date,
  contract_value numeric(16,2),
  technical_manager varchar(255),
  professional_registration varchar(160),
  art_cat_reference varchar(180),
  cat_number varchar(180),
  technical_professional_id uuid REFERENCES technical_professionals(id) ON DELETE SET NULL,
  cat_professional varchar(255),
  professional_role varchar(180),
  completion_status varchar(30) NOT NULL DEFAULT 'final',
  usage_scope varchar(30) NOT NULL DEFAULT 'both',
  tags text NOT NULL DEFAULT '',
  document_url text NOT NULL,
  file_name varchar(255) NOT NULL,
  mime_type varchar(120),
  file_size bigint,
  extracted_text text NOT NULL DEFAULT '',
  extracted_text_file_path text,
  extraction_status varchar(30) NOT NULL DEFAULT 'pending_ocr',
  status varchar(30) NOT NULL DEFAULT 'active',
  uploaded_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT technical_certificates_extraction_status_chk CHECK (extraction_status IN ('extracted', 'ocr_extracted', 'manual', 'pending_ocr', 'failed')),
  CONSTRAINT technical_certificates_status_chk CHECK (status IN ('active', 'archived')),
  CONSTRAINT technical_certificates_completion_status_chk CHECK (completion_status IN ('final', 'partial')),
  CONSTRAINT technical_certificates_usage_scope_chk CHECK (usage_scope IN ('company', 'professional', 'both'))
);

CREATE TABLE IF NOT EXISTS technical_certificate_quantities (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  technical_certificate_id uuid NOT NULL REFERENCES technical_certificates(id) ON DELETE CASCADE,
  description text NOT NULL,
  quantity numeric(18,4),
  unit varchar(80),
  note text,
  display_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS technical_certificate_ai_analyses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  requested_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  certificate_ids jsonb NOT NULL,
  prompt text NOT NULL,
  input_snapshot jsonb,
  status varchar(30) NOT NULL DEFAULT 'queued',
  model varchar(120) NOT NULL,
  response_id varchar(180),
  result_text text,
  error_message text,
  created_at timestamptz NOT NULL DEFAULT now(),
  started_at timestamptz,
  completed_at timestamptz,
  CONSTRAINT technical_certificate_ai_analyses_status_chk CHECK (status IN ('queued', 'processing', 'completed', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_technical_certificate_ai_analyses_company_created
  ON technical_certificate_ai_analyses(company_id, created_at DESC);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'users_profile_photo_media_fk'
  ) THEN
    ALTER TABLE users
      ADD CONSTRAINT users_profile_photo_media_fk
      FOREIGN KEY (profile_photo_media_id) REFERENCES media_files(id);
  END IF;
END;
$$;

-- =========================================================
-- Company module
-- =========================================================

CREATE TABLE IF NOT EXISTS company_profiles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL UNIQUE REFERENCES companies(id),
  website text,
  company_size varchar(40),
  institutional_description text,
  logo_media_id uuid REFERENCES media_files(id),
  state varchar(2),
  city varchar(120),
  national_coverage boolean NOT NULL DEFAULT false,
  public_profile_slug varchar(160) UNIQUE,
  is_public boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT company_profiles_company_size_chk CHECK (
    company_size IS NULL OR company_size IN ('small', 'medium', 'large')
  )
);

CREATE TABLE IF NOT EXISTS technical_areas (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(140) NOT NULL,
  slug varchar(160) NOT NULL UNIQUE,
  description text,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS company_technical_areas (
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  technical_area_id uuid NOT NULL REFERENCES technical_areas(id),
  description text,
  is_primary boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (company_id, technical_area_id)
);

-- =========================================================
-- Radar LicitaHub / News
-- =========================================================

CREATE TABLE IF NOT EXISTS news_categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(120) NOT NULL,
  slug varchar(140) NOT NULL UNIQUE,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS news (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id uuid REFERENCES news_categories(id),
  created_by_user_id uuid REFERENCES users(id),
  title varchar(220) NOT NULL,
  status varchar(40) NOT NULL DEFAULT 'draft',
  summary text,
  content text,
  main_image_media_id uuid REFERENCES media_files(id),
  published_at timestamptz,
  expires_at timestamptz,
  archived_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT news_status_chk CHECK (status IN ('draft', 'published', 'featured', 'archived', 'expired'))
);

-- =========================================================
-- Community
-- =========================================================

CREATE TABLE IF NOT EXISTS post_categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(120) NOT NULL,
  slug varchar(140) NOT NULL UNIQUE,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS posts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid NOT NULL REFERENCES companies(id),
  author_user_id uuid REFERENCES users(id),
  category_id uuid REFERENCES post_categories(id),
  title varchar(220),
  content text NOT NULL,
  main_image_media_id uuid REFERENCES media_files(id),
  status varchar(40) NOT NULL DEFAULT 'draft',
  visibility varchar(40) NOT NULL DEFAULT 'community',
  origin varchar(60) NOT NULL DEFAULT 'manual',
  published_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT posts_status_chk CHECK (status IN ('draft', 'published', 'archived')),
  CONSTRAINT posts_visibility_chk CHECK (visibility IN ('community', 'profile', 'both')),
  CONSTRAINT posts_origin_chk CHECK (origin IN ('manual', 'automatic_new_professional'))
);

CREATE TABLE IF NOT EXISTS post_images (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  media_id uuid NOT NULL REFERENCES media_files(id),
  caption varchar(255),
  position integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS post_likes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id),
  company_id uuid REFERENCES companies(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT post_likes_post_user_uk UNIQUE (post_id, user_id)
);

CREATE TABLE IF NOT EXISTS post_comments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id),
  company_id uuid REFERENCES companies(id),
  content text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS post_favorites (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT post_favorites_post_user_uk UNIQUE (post_id, user_id)
);

-- =========================================================
-- Tenders
-- =========================================================

CREATE TABLE IF NOT EXISTS tenders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_by_user_id uuid REFERENCES users(id),
  agency varchar(220) NOT NULL,
  number varchar(80) NOT NULL,
  object text NOT NULL,
  modality varchar(100),
  judgment_criterion varchar(100),
  estimated_value numeric(16, 2),
  state varchar(2),
  city varchar(120),
  scope_region varchar(160),
  opening_date timestamptz,
  status varchar(40) NOT NULL DEFAULT 'draft',
  cloud_folder_url text,
  source varchar(40) NOT NULL DEFAULT 'manual',
  source_reference varchar(220),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT tenders_status_chk CHECK (status IN (
    'draft', 'published', 'under_review', 'suspended', 'challenged', 'occurred', 'closed', 'cancelled'
  ))
);

CREATE UNIQUE INDEX IF NOT EXISTS tenders_source_reference_uk
  ON tenders(source, source_reference)
  WHERE source_reference IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS tender_timeline_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  event_type varchar(50) NOT NULL,
  title varchar(180) NOT NULL,
  description text,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  actor_company_id uuid REFERENCES companies(id) ON DELETE SET NULL,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tender_timeline_tender_created
  ON tender_timeline_events(tender_id, created_at ASC);

CREATE OR REPLACE FUNCTION licitahub_register_tender_timeline()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
  next_type varchar(50);
  next_title varchar(180);
  next_description text;
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.source <> 'manual' THEN
      next_type := 'captured';
      next_title := 'Edital captado';
      next_description := 'O edital foi recebido de uma fonte oficial e preparado no LicitaHub.';
    ELSE
      next_type := 'created';
      next_title := 'Edital cadastrado';
      next_description := 'O edital foi cadastrado no LicitaHub.';
    END IF;
    INSERT INTO tender_timeline_events (tender_id, event_type, title, description, actor_user_id)
    VALUES (NEW.id, next_type, next_title, next_description, NEW.created_by_user_id);
    RETURN NEW;
  END IF;

  IF NEW.status IS DISTINCT FROM OLD.status THEN
    CASE NEW.status
      WHEN 'published' THEN
        next_type := CASE WHEN OLD.status = 'suspended' THEN 'resumed' ELSE 'published' END;
        next_title := CASE WHEN OLD.status = 'suspended' THEN 'Edital retomado' ELSE 'Edital publicado' END;
        next_description := CASE WHEN OLD.status = 'suspended' THEN 'O edital voltou a estar disponível para acompanhamento.' ELSE 'O edital foi disponibilizado para as empresas.' END;
      WHEN 'suspended' THEN next_type := 'suspended'; next_title := 'Edital suspenso'; next_description := 'O edital e suas atividades relacionadas foram suspensos.';
      WHEN 'occurred' THEN next_type := 'occurred'; next_title := 'Sessão ocorrida'; next_description := 'A data de abertura do edital foi alcançada.';
      WHEN 'closed' THEN next_type := 'closed'; next_title := 'Edital encerrado'; next_description := 'O edital foi encerrado.';
      WHEN 'cancelled' THEN next_type := 'cancelled'; next_title := 'Edital cancelado'; next_description := 'O edital foi cancelado.';
      WHEN 'under_review' THEN next_type := 'under_review'; next_title := 'Edital em análise'; next_description := 'O edital está em análise.';
      WHEN 'challenged' THEN next_type := 'challenged'; next_title := 'Edital impugnado'; next_description := 'O edital possui uma impugnação registrada.';
      ELSE next_type := 'status_changed'; next_title := 'Status atualizado'; next_description := 'O status do edital foi atualizado.';
    END CASE;
    INSERT INTO tender_timeline_events (tender_id, event_type, title, description, metadata)
    VALUES (NEW.id, next_type, next_title, next_description, jsonb_build_object('previousStatus', OLD.status, 'status', NEW.status));
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_tenders_timeline ON tenders;
CREATE TRIGGER trg_tenders_timeline
AFTER INSERT OR UPDATE OF status ON tenders
FOR EACH ROW EXECUTE FUNCTION licitahub_register_tender_timeline();

INSERT INTO tender_timeline_events (tender_id, event_type, title, description, actor_user_id, created_at)
SELECT t.id,
  CASE WHEN t.source <> 'manual' THEN 'captured' ELSE 'created' END,
  CASE WHEN t.source <> 'manual' THEN 'Edital captado' ELSE 'Edital cadastrado' END,
  CASE WHEN t.source <> 'manual' THEN 'O edital foi recebido de uma fonte oficial e preparado no LicitaHub.' ELSE 'O edital foi cadastrado no LicitaHub.' END,
  t.created_by_user_id, t.created_at
FROM tenders t
WHERE NOT EXISTS (SELECT 1 FROM tender_timeline_events e WHERE e.tender_id = t.id);

INSERT INTO tender_timeline_events (tender_id, event_type, title, description, metadata, created_at)
SELECT t.id, te.event_type, te.title, te.description, jsonb_build_object('status', t.status), t.updated_at
FROM tenders t
JOIN LATERAL (SELECT CASE t.status
  WHEN 'published' THEN 'published' WHEN 'suspended' THEN 'suspended' WHEN 'occurred' THEN 'occurred'
  WHEN 'closed' THEN 'closed' WHEN 'cancelled' THEN 'cancelled' WHEN 'under_review' THEN 'under_review'
  WHEN 'challenged' THEN 'challenged' ELSE NULL END AS event_type,
  CASE t.status
  WHEN 'published' THEN 'Edital publicado' WHEN 'suspended' THEN 'Edital suspenso' WHEN 'occurred' THEN 'Sessão ocorrida'
  WHEN 'closed' THEN 'Edital encerrado' WHEN 'cancelled' THEN 'Edital cancelado' WHEN 'under_review' THEN 'Edital em análise'
  WHEN 'challenged' THEN 'Edital impugnado' ELSE NULL END AS title,
  'Status atual do edital no momento da ativacao da linha do tempo.' AS description) te ON te.event_type IS NOT NULL
WHERE NOT EXISTS (SELECT 1 FROM tender_timeline_events e WHERE e.tender_id = t.id AND e.metadata->>'status' = t.status);

CREATE TABLE IF NOT EXISTS pncp_captures (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source varchar(40) NOT NULL DEFAULT 'pncp',
  source_key varchar(220) NOT NULL UNIQUE,
  pncp_control_number varchar(160),
  agency varchar(220) NOT NULL,
  number varchar(120) NOT NULL,
  object text NOT NULL,
  modality varchar(120),
  judgment_criterion varchar(120),
  state varchar(2),
  city varchar(120),
  opening_date timestamptz,
  estimated_value numeric(16, 2),
  external_url text,
  relevance_score integer NOT NULL DEFAULT 0,
  relevance_reasons jsonb NOT NULL DEFAULT '[]'::jsonb,
  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  status varchar(40) NOT NULL DEFAULT 'captured',
  captured_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  approved_at timestamptz,
  approved_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  discarded_at timestamptz,
  discarded_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  published_tender_id uuid REFERENCES tenders(id) ON DELETE SET NULL,
  CONSTRAINT pncp_captures_status_chk CHECK (status IN ('captured', 'prepared', 'approved', 'discarded'))
);

ALTER TABLE pncp_captures ADD COLUMN IF NOT EXISTS source varchar(40) NOT NULL DEFAULT 'pncp';
ALTER TABLE pncp_captures ADD COLUMN IF NOT EXISTS judgment_criterion varchar(120);
ALTER TABLE pncp_captures ADD COLUMN IF NOT EXISTS pncp_control_number varchar(160);
CREATE INDEX IF NOT EXISTS idx_pncp_captures_control_number ON pncp_captures(pncp_control_number) WHERE pncp_control_number IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_pncp_captures_status_relevance
  ON pncp_captures(status, relevance_score DESC, opening_date ASC);

CREATE TABLE IF NOT EXISTS tender_files (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  media_id uuid REFERENCES media_files(id),
  file_type varchar(60) NOT NULL,
  title varchar(220) NOT NULL,
  file_url text,
  mime_type varchar(120),
  file_size bigint,
  is_downloadable boolean NOT NULL DEFAULT true,
  uploaded_by_user_id uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tender_ai_analysis_jobs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  requested_by_user_id uuid REFERENCES users(id),
  status varchar(30) NOT NULL DEFAULT 'queued',
  model varchar(120),
  prompt_version varchar(80) NOT NULL DEFAULT 'analise-primaria-v1',
  source_file_count integer NOT NULL DEFAULT 0,
  response_id varchar(160),
  analysis_file_id uuid REFERENCES tender_files(id) ON DELETE SET NULL,
  error_message text,
  created_at timestamptz NOT NULL DEFAULT now(),
  started_at timestamptz,
  completed_at timestamptz,
  CONSTRAINT tender_ai_analysis_jobs_status_chk CHECK (status IN ('queued', 'processing', 'completed', 'failed'))
);

CREATE TABLE IF NOT EXISTS tender_challenges (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  subject varchar(220) NOT NULL,
  rationale text NOT NULL,
  status varchar(40) NOT NULL DEFAULT 'submitted',
  business_days_before_opening integer,
  is_untimely boolean NOT NULL DEFAULT false,
  submitted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT tender_challenges_status_chk CHECK (status IN ('draft', 'submitted', 'under_review', 'accepted', 'rejected', 'withdrawn')),
  CONSTRAINT tender_challenges_tender_company_uk UNIQUE (tender_id, company_id)
);

CREATE TABLE IF NOT EXISTS tender_challenge_files (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_challenge_id uuid NOT NULL REFERENCES tender_challenges(id) ON DELETE CASCADE,
  title varchar(220) NOT NULL,
  file_url text NOT NULL,
  mime_type varchar(120),
  file_size bigint,
  uploaded_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tender_requirement_types (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  key varchar(80) NOT NULL UNIQUE,
  name varchar(160) NOT NULL,
  description text,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tender_requirements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  requirement_type_id uuid NOT NULL REFERENCES tender_requirement_types(id),
  is_required boolean NOT NULL DEFAULT true,
  description text,
  weight numeric(8, 2),
  order_index integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT tender_requirements_tender_type_uk UNIQUE (tender_id, requirement_type_id)
);

CREATE TABLE IF NOT EXISTS tender_interests (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id),
  created_by_user_id uuid REFERENCES users(id),
  general_position varchar(60) NOT NULL DEFAULT 'interested',
  desired_role varchar(80) NOT NULL DEFAULT 'seeks_partner',
  participation_mode varchar(40) NOT NULL DEFAULT 'seeking_partners',
  public_summary text,
  internal_note text,
  visibility varchar(60) NOT NULL DEFAULT 'visible_to_interested',
  status varchar(40) NOT NULL DEFAULT 'draft',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT tender_interests_general_position_chk CHECK (general_position IN (
    'interested', 'under_evaluation', 'watching', 'not_interested'
  )),
  CONSTRAINT tender_interests_desired_role_chk CHECK (desired_role IN (
    'seeks_partner', 'can_lead_consortium', 'complementary_partner',
    'seeks_lead_company', 'evaluating_role'
  )),
  CONSTRAINT tender_interests_participation_mode_chk CHECK (participation_mode IN (
    'individual', 'seeking_partners', 'undecided'
  )),
  CONSTRAINT tender_interests_visibility_chk CHECK (visibility IN (
    'private', 'visible_to_interested', 'public_showcase'
  )),
  CONSTRAINT tender_interests_status_chk CHECK (status IN (
    'draft', 'published', 'hidden', 'withdrawn'
  ))
);

CREATE UNIQUE INDEX IF NOT EXISTS tender_interests_active_uk
  ON tender_interests (tender_id, company_id)
  WHERE deleted_at IS NULL AND status <> 'withdrawn';

CREATE TABLE IF NOT EXISTS tender_interest_requirements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_interest_id uuid NOT NULL REFERENCES tender_interests(id) ON DELETE CASCADE,
  tender_requirement_id uuid NOT NULL REFERENCES tender_requirements(id) ON DELETE CASCADE,
  status_key varchar(100) NOT NULL,
  what_we_have text,
  what_we_seek text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT tender_interest_requirements_uk UNIQUE (tender_interest_id, tender_requirement_id)
);

-- =========================================================
-- Partnership showcase, match and consortium
-- =========================================================

CREATE TABLE IF NOT EXISTS partnership_ads (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id),
  tender_interest_id uuid REFERENCES tender_interests(id) ON DELETE SET NULL,
  consortium_intention_id uuid,
  ad_type varchar(40) NOT NULL DEFAULT 'company',
  title varchar(220),
  offer_summary text,
  seek_summary text,
  status varchar(40) NOT NULL DEFAULT 'draft',
  paused_by_tender boolean NOT NULL DEFAULT false,
  published_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT partnership_ads_type_chk CHECK (ad_type IN ('company', 'consortium')),
  CONSTRAINT partnership_ads_status_chk CHECK (status IN (
    'draft', 'published', 'paused', 'closed'
  ))
);

CREATE TABLE IF NOT EXISTS partner_evaluations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  evaluator_company_id uuid NOT NULL REFERENCES companies(id),
  evaluated_company_id uuid NOT NULL REFERENCES companies(id),
  evaluated_ad_id uuid REFERENCES partnership_ads(id) ON DELETE SET NULL,
  decision varchar(40) NOT NULL,
  created_by_user_id uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT partner_evaluations_decision_chk CHECK (decision IN ('liked', 'rejected', 'later')),
  CONSTRAINT partner_evaluations_not_self_chk CHECK (evaluator_company_id <> evaluated_company_id),
  CONSTRAINT partner_evaluations_uk UNIQUE (
    tender_id, evaluator_company_id, evaluated_company_id
  )
);

CREATE TABLE IF NOT EXISTS matches (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  company_a_id uuid NOT NULL REFERENCES companies(id),
  company_b_id uuid NOT NULL REFERENCES companies(id),
  company_a_evaluation_id uuid REFERENCES partner_evaluations(id),
  company_b_evaluation_id uuid REFERENCES partner_evaluations(id),
  status varchar(40) NOT NULL DEFAULT 'active',
  matched_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT matches_status_chk CHECK (status IN ('active', 'cancelled', 'closed')),
  CONSTRAINT matches_company_order_chk CHECK (company_a_id < company_b_id),
  CONSTRAINT matches_uk UNIQUE (tender_id, company_a_id, company_b_id)
);

CREATE TABLE IF NOT EXISTS match_contacts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  match_id uuid NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id),
  contact_user_id uuid REFERENCES users(id),
  contact_name varchar(180) NOT NULL,
  phone varchar(40) NOT NULL,
  whatsapp_url text,
  message_template text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

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

CREATE TABLE IF NOT EXISTS consortium_intentions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  match_id uuid REFERENCES matches(id) ON DELETE SET NULL,
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  lead_company_id uuid REFERENCES companies(id),
  status varchar(60) NOT NULL DEFAULT 'in_conversation',
  notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT consortium_intentions_status_chk CHECK (status IN (
    'in_conversation', 'intention_registered', 'withdrawn', 'advanced_to_consortium'
  ))
);

CREATE TABLE IF NOT EXISTS consortium_members (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  consortium_intention_id uuid NOT NULL REFERENCES consortium_intentions(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id),
  role varchar(120),
  responsibility_description text,
  status varchar(40) NOT NULL DEFAULT 'active',
  withdrawn_at timestamptz,
  withdrawn_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT consortium_members_status_chk CHECK (status IN ('active', 'withdrawn')),
  CONSTRAINT consortium_members_uk UNIQUE (consortium_intention_id, company_id)
);

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

-- =========================================================
-- Central de montagem da licitacao
-- =========================================================

CREATE TABLE IF NOT EXISTS bid_assembly_templates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id uuid REFERENCES companies(id) ON DELETE CASCADE,
  name varchar(180) NOT NULL,
  description text,
  is_system boolean NOT NULL DEFAULT false,
  is_active boolean NOT NULL DEFAULT true,
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS bid_assembly_templates_system_name_uk
  ON bid_assembly_templates(name) WHERE is_system = true;

CREATE TABLE IF NOT EXISTS bid_assembly_template_stages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  template_id uuid NOT NULL REFERENCES bid_assembly_templates(id) ON DELETE CASCADE,
  title varchar(180) NOT NULL,
  description text,
  position integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bid_assembly_template_tasks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  template_stage_id uuid NOT NULL REFERENCES bid_assembly_template_stages(id) ON DELETE CASCADE,
  title varchar(220) NOT NULL,
  description text,
  position integer NOT NULL DEFAULT 0,
  default_weight numeric(8,2) NOT NULL DEFAULT 1,
  default_days_offset integer,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bid_assemblies (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  consortium_intention_id uuid REFERENCES consortium_intentions(id) ON DELETE RESTRICT,
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE RESTRICT,
  template_id uuid REFERENCES bid_assembly_templates(id) ON DELETE SET NULL,
  assembly_type varchar(30) NOT NULL DEFAULT 'consortium',
  owner_company_id uuid REFERENCES companies(id),
  lead_company_id uuid NOT NULL REFERENCES companies(id),
  title varchar(220) NOT NULL,
  status varchar(40) NOT NULL DEFAULT 'preparing',
  status_before_pause varchar(40),
  start_date date NOT NULL DEFAULT CURRENT_DATE,
  due_date date,
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bid_assemblies_status_chk CHECK (status IN ('preparing', 'in_progress', 'under_review', 'ready_to_submit', 'submitted', 'paused', 'cancelled')),
  CONSTRAINT bid_assemblies_type_chk CHECK (assembly_type IN ('consortium', 'individual')),
  CONSTRAINT bid_assemblies_consortium_uk UNIQUE (consortium_intention_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS bid_assemblies_individual_company_tender_uk
  ON bid_assemblies(tender_id, owner_company_id)
  WHERE assembly_type = 'individual' AND status <> 'cancelled';

CREATE TABLE IF NOT EXISTS bid_assembly_participants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  assembly_id uuid NOT NULL REFERENCES bid_assemblies(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id),
  user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  role varchar(40) NOT NULL DEFAULT 'collaborator',
  status varchar(40) NOT NULL DEFAULT 'active',
  joined_at timestamptz NOT NULL DEFAULT now(),
  removed_at timestamptz,
  CONSTRAINT bid_assembly_participants_role_chk CHECK (role IN ('coordinator', 'collaborator', 'viewer')),
  CONSTRAINT bid_assembly_participants_status_chk CHECK (status IN ('active', 'removed'))
);

CREATE UNIQUE INDEX IF NOT EXISTS bid_assembly_participants_company_uk
  ON bid_assembly_participants(assembly_id, company_id) WHERE user_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS bid_assembly_participants_user_uk
  ON bid_assembly_participants(assembly_id, user_id) WHERE user_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS bid_assembly_stages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  assembly_id uuid NOT NULL REFERENCES bid_assemblies(id) ON DELETE CASCADE,
  source_template_stage_id uuid REFERENCES bid_assembly_template_stages(id) ON DELETE SET NULL,
  title varchar(180) NOT NULL,
  description text,
  position integer NOT NULL DEFAULT 0,
  is_custom boolean NOT NULL DEFAULT false,
  is_archived boolean NOT NULL DEFAULT false,
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bid_assembly_tasks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  stage_id uuid NOT NULL REFERENCES bid_assembly_stages(id) ON DELETE CASCADE,
  source_template_task_id uuid REFERENCES bid_assembly_template_tasks(id) ON DELETE SET NULL,
  title varchar(220) NOT NULL,
  description text,
  position integer NOT NULL DEFAULT 0,
  status varchar(40) NOT NULL DEFAULT 'pending',
  priority varchar(20) NOT NULL DEFAULT 'normal',
  weight numeric(8,2) NOT NULL DEFAULT 1,
  responsible_company_id uuid REFERENCES companies(id) ON DELETE SET NULL,
  responsible_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  due_at timestamptz,
  submitted_at timestamptz,
  completed_at timestamptz,
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  is_custom boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bid_assembly_tasks_status_chk CHECK (status IN ('pending', 'in_progress', 'waiting_information', 'blocked', 'under_review', 'returned_for_adjustment', 'completed', 'not_applicable')),
  CONSTRAINT bid_assembly_tasks_priority_chk CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
  CONSTRAINT bid_assembly_tasks_weight_chk CHECK (weight > 0)
);

CREATE TABLE IF NOT EXISTS bid_assembly_task_comments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id uuid NOT NULL REFERENCES bid_assembly_tasks(id) ON DELETE CASCADE,
  user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  company_id uuid REFERENCES companies(id) ON DELETE SET NULL,
  content text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS bid_assembly_task_evidences (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id uuid NOT NULL REFERENCES bid_assembly_tasks(id) ON DELETE CASCADE,
  media_file_id uuid REFERENCES media_files(id) ON DELETE SET NULL,
  evidence_type varchar(30) NOT NULL,
  title varchar(220) NOT NULL,
  external_url text,
  note text,
  version_number integer NOT NULL DEFAULT 1,
  status varchar(30) NOT NULL DEFAULT 'current',
  uploaded_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  company_id uuid REFERENCES companies(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bid_assembly_evidence_type_chk CHECK (evidence_type IN ('file', 'link', 'note')),
  CONSTRAINT bid_assembly_evidence_status_chk CHECK (status IN ('current', 'superseded', 'approved'))
);

CREATE TABLE IF NOT EXISTS bid_assembly_deadline_alerts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id uuid NOT NULL REFERENCES bid_assembly_tasks(id) ON DELETE CASCADE,
  recipient_user_id uuid REFERENCES users(id) ON DELETE CASCADE,
  alert_type varchar(30) NOT NULL,
  alert_date date NOT NULL DEFAULT CURRENT_DATE,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT bid_assembly_deadline_alerts_type_chk CHECK (alert_type IN ('due_soon', 'due_today', 'overdue')),
  CONSTRAINT bid_assembly_deadline_alerts_uk UNIQUE (task_id, recipient_user_id, alert_type, alert_date)
);

CREATE TABLE IF NOT EXISTS bid_assembly_activity_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  assembly_id uuid NOT NULL REFERENCES bid_assemblies(id) ON DELETE CASCADE,
  task_id uuid REFERENCES bid_assembly_tasks(id) ON DELETE CASCADE,
  actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  actor_company_id uuid REFERENCES companies(id) ON DELETE SET NULL,
  action varchar(80) NOT NULL,
  description text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bid_assembly_stages_assembly_position ON bid_assembly_stages(assembly_id, position);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_tasks_stage_position ON bid_assembly_tasks(stage_id, position);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_tasks_responsible_due ON bid_assembly_tasks(responsible_user_id, due_at, status);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_comments_task ON bid_assembly_task_comments(task_id, created_at);
CREATE INDEX IF NOT EXISTS idx_bid_assembly_evidences_task ON bid_assembly_task_evidences(task_id, created_at);

INSERT INTO bid_assembly_templates (name, description, is_system, is_active)
VALUES ('Modelo LicitaHub', 'Estrutura padrao de montagem colaborativa de licitacoes em oito fases.', true, true)
ON CONFLICT DO NOTHING;

WITH template AS (
  SELECT id FROM bid_assembly_templates WHERE name = 'Modelo LicitaHub' AND is_system = true LIMIT 1
), incoming(title, description, position) AS (
  VALUES
    ('Planejamento da montagem', 'Leitura do edital, estrategia, calendario e organizacao do trabalho.', 1),
    ('Concepcao consorcial', 'Formalizacao, identidade e responsabilidades do consorcio.', 2),
    ('Montagem da peca qualitativa', 'Desenvolvimento dos conteudos tecnicos e qualitativos da proposta.', 3),
    ('Montagem do orcamento', 'Viabilidade, formacao de preco e planilhas comerciais.', 4),
    ('Montagem da equipe tecnica', 'Definicao da equipe e consolidacao de comprovacoes profissionais.', 5),
    ('Montagem das declaracoes', 'Preparacao e assinatura das declaracoes exigidas.', 6),
    ('Certificacoes e quesitos de pontuacao', 'Organizacao de certificados e comprovacoes adicionais.', 7),
    ('Revisao e consolidacao', 'Conferencia, consolidacao, assinatura e preparacao para envio.', 8)
)
INSERT INTO bid_assembly_template_stages (template_id, title, description, position)
SELECT template.id, incoming.title, incoming.description, incoming.position
FROM template CROSS JOIN incoming
WHERE NOT EXISTS (
  SELECT 1 FROM bid_assembly_template_stages existing
  WHERE existing.template_id = template.id AND existing.title = incoming.title
);

WITH template AS (
  SELECT id FROM bid_assembly_templates WHERE name = 'Modelo LicitaHub' AND is_system = true LIMIT 1
), incoming(stage_title, title, description, position, weight) AS (
  VALUES
    ('Concepcao consorcial', 'Elaborar o termo de constituicao do consorcio', 'Consolidar participantes, objeto, compromissos e regras de representacao.', 1, 2),
    ('Concepcao consorcial', 'Definir nome do consorcio', 'Registrar a denominacao que sera usada na proposta.', 2, 1),
    ('Concepcao consorcial', 'Definir identidade visual do consorcio', 'Organizar logomarca, capa e padrao visual dos documentos.', 3, 1),
    ('Concepcao consorcial', 'Definir papeis e responsabilidades das empresas', 'Registrar lideranca, entregas e pontos focais de cada consorciada.', 4, 2),
    ('Planejamento da montagem', 'Realizar leitura orientada do edital', 'Mapear entregaveis, criterios, prazos e riscos da proposta.', 1, 2),
    ('Planejamento da montagem', 'Montar calendario geral da proposta', 'Definir marcos internos anteriores a data oficial de entrega.', 2, 1),
    ('Planejamento da montagem', 'Distribuir responsabilidades por fase', 'Atribuir empresas e profissionais responsaveis.', 3, 1),
    ('Planejamento da montagem', 'Organizar repositorio compartilhado', 'Definir estrutura de pastas, nomes e controle de versoes.', 4, 1),
    ('Montagem da peca qualitativa', 'Elaborar conhecimento do objeto', 'Descrever compreensao do contexto, desafios e objetivos do contrato.', 1, 3),
    ('Montagem da peca qualitativa', 'Descrever produtos e entregaveis', 'Consolidar produtos, resultados e criterios de aceitacao.', 2, 2),
    ('Montagem da peca qualitativa', 'Desenvolver metodologia', 'Detalhar abordagem, procedimentos, ferramentas e integracao das disciplinas.', 3, 4),
    ('Montagem da peca qualitativa', 'Elaborar plano de trabalho', 'Organizar atividades, sequencia, interfaces e responsabilidades.', 4, 3),
    ('Montagem da peca qualitativa', 'Revisar aderencia aos criterios tecnicos', 'Conferir atendimento e potencial de pontuacao da peca.', 5, 2),
    ('Montagem do orcamento', 'Analisar viabilidade financeira', 'Avaliar custos, riscos, fluxo e condicoes comerciais.', 1, 2),
    ('Montagem do orcamento', 'Definir preco da proposta', 'Consolidar estrategia de preco e premissas comerciais.', 2, 2),
    ('Montagem do orcamento', 'Montar planilha orcamentaria', 'Preparar quantitativos, custos, encargos e composicoes.', 3, 3),
    ('Montagem do orcamento', 'Revisar tributos, BDI e consistencia', 'Conferir calculos, incidencias e compatibilidade com o edital.', 4, 2),
    ('Montagem da equipe tecnica', 'Analisar profissionais disponiveis', 'Verificar disponibilidade, vinculo e aderencia dos profissionais.', 1, 2),
    ('Montagem da equipe tecnica', 'Montar quadro de experiencia profissional', 'Consolidar experiencias e pontuacoes por profissional.', 2, 2),
    ('Montagem da equipe tecnica', 'Disponibilizar CATs dos profissionais', 'Anexar comprovacoes de responsabilidade e experiencia tecnica.', 3, 2),
    ('Montagem da equipe tecnica', 'Disponibilizar documentos de formacao', 'Anexar diplomas, certificados e demais comprovacoes academicas.', 4, 1),
    ('Montagem da equipe tecnica', 'Preparar declaracoes dos profissionais', 'Consolidar disponibilidade, compromisso e autorizacoes exigidas.', 5, 1),
    ('Montagem da equipe tecnica', 'Disponibilizar registros profissionais', 'Anexar registros e certidoes dos conselhos de classe.', 6, 1),
    ('Montagem da equipe tecnica', 'Revisar atendimento e pontuacao da equipe', 'Conferir lacunas, validade e potencial de pontuacao.', 7, 2),
    ('Montagem das declaracoes', 'Mapear declaracoes exigidas', 'Criar lista completa conforme edital e anexos.', 1, 1),
    ('Montagem das declaracoes', 'Elaborar declaracoes', 'Preencher os modelos e textos aplicaveis ao consorcio.', 2, 2),
    ('Montagem das declaracoes', 'Coletar assinaturas e validar poderes', 'Conferir assinantes, procuracoes e versoes finais.', 3, 2),
    ('Certificacoes e quesitos de pontuacao', 'Mapear certificacoes requeridas', 'Identificar certificados obrigatorios e pontuaveis.', 1, 1),
    ('Certificacoes e quesitos de pontuacao', 'Disponibilizar documentos comprobatorios', 'Anexar certificacoes e demais evidencias de pontuacao.', 2, 2),
    ('Certificacoes e quesitos de pontuacao', 'Conferir validade e aderencia', 'Verificar emissor, prazo, escopo e criterio atendido.', 3, 1),
    ('Revisao e consolidacao', 'Consolidar documentos da proposta', 'Reunir somente versoes atuais e aprovadas.', 1, 3),
    ('Revisao e consolidacao', 'Realizar revisao cruzada', 'Conferir coerencia entre tecnica, equipe, documentos e preco.', 2, 3),
    ('Revisao e consolidacao', 'Conferir assinaturas e formatos', 'Validar assinaturas, extensoes, limites e nomenclaturas.', 3, 2),
    ('Revisao e consolidacao', 'Validar dossie final', 'Confirmar que todas as evidencias obrigatorias estao presentes.', 4, 2),
    ('Revisao e consolidacao', 'Registrar protocolo de envio', 'Guardar comprovante, data, horario e versao submetida.', 5, 1)
)
INSERT INTO bid_assembly_template_tasks (template_stage_id, title, description, position, default_weight)
SELECT stage.id, incoming.title, incoming.description, incoming.position, incoming.weight
FROM incoming
JOIN template ON true
JOIN bid_assembly_template_stages stage ON stage.template_id = template.id AND stage.title = incoming.stage_title
WHERE NOT EXISTS (
  SELECT 1 FROM bid_assembly_template_tasks existing
  WHERE existing.template_stage_id = stage.id AND existing.title = incoming.title
);

-- =========================================================
-- Notifications and audit
-- =========================================================

CREATE TABLE IF NOT EXISTS notifications (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  recipient_user_id uuid REFERENCES users(id),
  recipient_company_id uuid REFERENCES companies(id),
  type varchar(60) NOT NULL,
  title varchar(180) NOT NULL,
  message text,
  destination_screen varchar(120),
  related_entity_type varchar(80),
  related_entity_id uuid,
  is_read boolean NOT NULL DEFAULT false,
  read_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT notifications_type_chk CHECK (type IN (
    'post_like', 'post_comment', 'match', 'company_interested',
    'news', 'system'
  ))
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id uuid REFERENCES users(id),
  company_id uuid REFERENCES companies(id),
  module varchar(80) NOT NULL,
  action varchar(120) NOT NULL,
  entity_type varchar(80),
  entity_id uuid,
  description text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- =========================================================
-- Updated-at triggers
-- =========================================================

DROP TRIGGER IF EXISTS trg_companies_updated_at ON companies;
CREATE TRIGGER trg_companies_updated_at
BEFORE UPDATE ON companies
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_access_profiles_updated_at ON access_profiles;
CREATE TRIGGER trg_access_profiles_updated_at
BEFORE UPDATE ON access_profiles
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_company_invitations_updated_at ON company_invitations;
CREATE TRIGGER trg_company_invitations_updated_at
BEFORE UPDATE ON company_invitations
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_company_profiles_updated_at ON company_profiles;
CREATE TRIGGER trg_company_profiles_updated_at
BEFORE UPDATE ON company_profiles
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_technical_areas_updated_at ON technical_areas;
CREATE TRIGGER trg_technical_areas_updated_at
BEFORE UPDATE ON technical_areas
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_news_categories_updated_at ON news_categories;
CREATE TRIGGER trg_news_categories_updated_at
BEFORE UPDATE ON news_categories
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_news_updated_at ON news;
CREATE TRIGGER trg_news_updated_at
BEFORE UPDATE ON news
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_post_categories_updated_at ON post_categories;
CREATE TRIGGER trg_post_categories_updated_at
BEFORE UPDATE ON post_categories
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_posts_updated_at ON posts;
CREATE TRIGGER trg_posts_updated_at
BEFORE UPDATE ON posts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_post_comments_updated_at ON post_comments;
CREATE TRIGGER trg_post_comments_updated_at
BEFORE UPDATE ON post_comments
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_tenders_updated_at ON tenders;
CREATE TRIGGER trg_tenders_updated_at
BEFORE UPDATE ON tenders
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_tender_requirement_types_updated_at ON tender_requirement_types;
CREATE TRIGGER trg_tender_requirement_types_updated_at
BEFORE UPDATE ON tender_requirement_types
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_tender_requirements_updated_at ON tender_requirements;
CREATE TRIGGER trg_tender_requirements_updated_at
BEFORE UPDATE ON tender_requirements
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_tender_interests_updated_at ON tender_interests;
CREATE TRIGGER trg_tender_interests_updated_at
BEFORE UPDATE ON tender_interests
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_tender_interest_requirements_updated_at ON tender_interest_requirements;
CREATE TRIGGER trg_tender_interest_requirements_updated_at
BEFORE UPDATE ON tender_interest_requirements
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_partnership_ads_updated_at ON partnership_ads;
CREATE TRIGGER trg_partnership_ads_updated_at
BEFORE UPDATE ON partnership_ads
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_partner_evaluations_updated_at ON partner_evaluations;
CREATE TRIGGER trg_partner_evaluations_updated_at
BEFORE UPDATE ON partner_evaluations
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_matches_updated_at ON matches;
CREATE TRIGGER trg_matches_updated_at
BEFORE UPDATE ON matches
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_match_contacts_updated_at ON match_contacts;
CREATE TRIGGER trg_match_contacts_updated_at
BEFORE UPDATE ON match_contacts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_chat_threads_updated_at ON chat_threads;
CREATE TRIGGER trg_chat_threads_updated_at
BEFORE UPDATE ON chat_threads
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_consortium_intentions_updated_at ON consortium_intentions;
CREATE TRIGGER trg_consortium_intentions_updated_at
BEFORE UPDATE ON consortium_intentions
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_consortium_applications_updated_at ON consortium_applications;
CREATE TRIGGER trg_consortium_applications_updated_at
BEFORE UPDATE ON consortium_applications
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =========================================================
-- Indexes
-- =========================================================

CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_company_status ON users(company_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_companies_status_deleted_at ON companies(status, deleted_at);
CREATE INDEX IF NOT EXISTS idx_company_invitations_status_created_at ON company_invitations(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_company_profiles_public ON company_profiles(is_public, public_profile_slug);
CREATE INDEX IF NOT EXISTS idx_media_files_company_id ON media_files(company_id);
CREATE INDEX IF NOT EXISTS idx_media_files_uploaded_by ON media_files(uploaded_by_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_technical_certificates_company_status ON technical_certificates(company_id, status, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_technical_professionals_company_status ON technical_professionals(company_id, status, full_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_technical_professional_educations_professional ON technical_professional_educations(technical_professional_id, display_order);
CREATE INDEX IF NOT EXISTS idx_technical_certificate_quantities_certificate ON technical_certificate_quantities(technical_certificate_id, display_order);
CREATE INDEX IF NOT EXISTS idx_technical_certificates_search ON technical_certificates USING GIN (
  to_tsvector('portuguese', coalesce(object, '') || ' ' || coalesce(service_description, '') || ' ' || coalesce(tags, '') || ' ' || coalesce(extracted_text, ''))
);
CREATE INDEX IF NOT EXISTS idx_posts_company_id ON posts(company_id);
CREATE INDEX IF NOT EXISTS idx_posts_published_at ON posts(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_feed ON posts(status, visibility, published_at DESC, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_posts_company_profile ON posts(company_id, status, visibility, published_at DESC, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_post_comments_post_id ON post_comments(post_id);
CREATE INDEX IF NOT EXISTS idx_post_comments_active ON post_comments(post_id, created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_post_likes_post_created_at ON post_likes(post_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_post_favorites_user_created_at ON post_favorites(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_news_status_expires_published_at ON news(status, expires_at, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_token ON auth_sessions(token_hash, expires_at) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens ON password_reset_tokens(token_hash, expires_at) WHERE used_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenders_status_opening_date ON tenders(status, opening_date);
CREATE INDEX IF NOT EXISTS idx_tender_files_tender_id ON tender_files(tender_id);
CREATE INDEX IF NOT EXISTS idx_tender_files_tender_type_created ON tender_files(tender_id, file_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tender_ai_analysis_jobs_tender_created ON tender_ai_analysis_jobs(tender_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tender_challenges_tender_status ON tender_challenges(tender_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tender_challenge_files_challenge ON tender_challenge_files(tender_challenge_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tender_requirements_tender_id ON tender_requirements(tender_id);
CREATE INDEX IF NOT EXISTS idx_tender_interests_tender_id ON tender_interests(tender_id);
CREATE INDEX IF NOT EXISTS idx_tender_interests_company_id ON tender_interests(company_id);
CREATE INDEX IF NOT EXISTS idx_tender_interests_tender_status ON tender_interests(tender_id, status, visibility) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_partnership_ads_tender_status ON partnership_ads(tender_id, status);
CREATE INDEX IF NOT EXISTS idx_partnership_ads_showcase ON partnership_ads(status, published_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_partnership_ads_consortium ON partnership_ads(consortium_intention_id, status) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS partnership_ads_active_interest_uk
  ON partnership_ads(tender_interest_id)
  WHERE tender_interest_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_partner_evaluations_tender_eval ON partner_evaluations(tender_id, evaluator_company_id);
CREATE INDEX IF NOT EXISTS idx_matches_tender_id ON matches(tender_id);
CREATE INDEX IF NOT EXISTS idx_matches_company_a_status ON matches(company_a_id, status, matched_at DESC);
CREATE INDEX IF NOT EXISTS idx_matches_company_b_status ON matches(company_b_id, status, matched_at DESC);
CREATE INDEX IF NOT EXISTS idx_match_contacts_match_company ON match_contacts(match_id, company_id);
CREATE INDEX IF NOT EXISTS idx_chat_threads_company_a ON chat_threads(company_a_id, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_chat_threads_company_b ON chat_threads(company_b_id, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_chat_messages_thread_created ON chat_messages(thread_id, created_at) WHERE deleted_at IS NULL;
ALTER TABLE chat_threads ALTER COLUMN partnership_ad_id DROP NOT NULL;
ALTER TABLE chat_threads ALTER COLUMN tender_id DROP NOT NULL;
ALTER TABLE chat_threads DROP CONSTRAINT IF EXISTS chat_threads_not_same_company_chk;
ALTER TABLE chat_threads ADD COLUMN IF NOT EXISTS context_type varchar(40) NOT NULL DEFAULT 'partnership_ad';
ALTER TABLE chat_threads ADD COLUMN IF NOT EXISTS context_id uuid;
ALTER TABLE chat_threads ADD COLUMN IF NOT EXISTS context_title varchar(220);
ALTER TABLE chat_threads ADD COLUMN IF NOT EXISTS context_key varchar(160);
CREATE UNIQUE INDEX IF NOT EXISTS chat_threads_context_uk ON chat_threads(context_type, context_id) WHERE context_id IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS chat_threads_context_key_uk ON chat_threads(context_type, context_key) WHERE context_key IS NOT NULL AND deleted_at IS NULL;
CREATE TABLE IF NOT EXISTS chat_thread_participants (
  thread_id uuid NOT NULL REFERENCES chat_threads(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  company_id uuid NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  participant_role varchar(40) NOT NULL DEFAULT 'member',
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (thread_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_consortium_intentions_match ON consortium_intentions(match_id);
CREATE INDEX IF NOT EXISTS idx_consortium_intentions_tender ON consortium_intentions(tender_id);
CREATE INDEX IF NOT EXISTS idx_consortium_members_active ON consortium_members(consortium_intention_id, company_id) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_consortium_applications_intention ON consortium_applications(consortium_intention_id, status);
CREATE INDEX IF NOT EXISTS idx_consortium_applications_ad ON consortium_applications(partnership_ad_id, status);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_user ON notifications(recipient_user_id, is_read, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_company ON notifications(recipient_company_id, is_read, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_related ON notifications(related_entity_type, related_entity_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- =========================================================
-- Seed data
-- =========================================================

-- Academia LicitaHub: cursos, aulas, progresso individual, provas e certificados.
CREATE TABLE IF NOT EXISTS academy_courses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title varchar(220) NOT NULL,
  description text NOT NULL DEFAULT '',
  category varchar(120) NOT NULL DEFAULT 'Geral',
  cover_image_url text,
  workload_hours integer NOT NULL DEFAULT 0,
  status varchar(30) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
  created_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz
);

CREATE TABLE IF NOT EXISTS academy_lessons (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id uuid NOT NULL REFERENCES academy_courses(id) ON DELETE CASCADE,
  title varchar(220) NOT NULL,
  description text NOT NULL DEFAULT '',
  video_url text NOT NULL,
  video_source varchar(20) NOT NULL DEFAULT 'youtube' CHECK (video_source IN ('youtube', 'upload')),
  duration_seconds integer NOT NULL DEFAULT 0,
  display_order integer NOT NULL DEFAULT 1,
  requires_quiz boolean NOT NULL DEFAULT false,
  quiz_questions jsonb NOT NULL DEFAULT '[]'::jsonb,
  passing_score integer NOT NULL DEFAULT 75 CHECK (passing_score BETWEEN 1 AND 100),
  max_attempts integer NOT NULL DEFAULT 0 CHECK (max_attempts >= 0),
  is_published boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
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

INSERT INTO access_profiles (key, name, description, is_system)
VALUES
  ('platform_admin', 'Administrador da plataforma', 'Acesso administrativo total da LicitaHub.', true),
  ('company_admin', 'Administrador da empresa', 'Gerencia perfil, usuarios, publicacoes, editais e match da empresa.', true),
  ('commercial', 'Comercial / Relacionamento', 'Atua em publicacoes, interesses e relacionamento com parceiros.', true),
  ('technical', 'Tecnico', 'Consulta editais, requisitos e apoia analises tecnicas.', true),
  ('reader', 'Leitor', 'Acesso de leitura aos conteudos permitidos.', true)
ON CONFLICT (key) DO NOTHING;

INSERT INTO news_categories (name, slug)
VALUES
  ('Licitacoes', 'licitacoes'),
  ('Mercado', 'mercado'),
  ('Legislacao', 'legislacao'),
  ('Eventos', 'eventos'),
  ('Comunicados', 'comunicados')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO post_categories (name, slug)
VALUES
  ('Equipe comercial', 'equipe-comercial'),
  ('Noticias', 'noticias'),
  ('Atividades', 'atividades'),
  ('Eventos', 'eventos'),
  ('Conquistas', 'conquistas'),
  ('Conteudo tecnico', 'conteudo-tecnico'),
  ('Destaque', 'destaque')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tender_requirement_types (key, name, description)
VALUES
  ('operational_qualification', 'Requisito operacional', 'Acervo, atestados, experiencia da empresa e pontuacao operacional exigidos pelo edital.'),
  ('professional_qualification', 'Requisitos profissionais', 'Equipe, responsaveis tecnicos, curriculos, CATs, disponibilidade e pontuacao profissional.'),
  ('technical_proposal', 'Peca tecnica qualitativa', 'Metodologia, plano de trabalho, abordagem tecnica e proposta qualitativa.'),
  ('certifications', 'Certificacoes requeridas', 'Certificacoes, registros ou comprovacoes formais exigidas.')
ON CONFLICT (key) DO NOTHING;

INSERT INTO technical_areas (name, slug)
VALUES
  ('Engenharia consultiva', 'engenharia-consultiva'),
  ('Saneamento', 'saneamento'),
  ('Supervisao ambiental', 'supervisao-ambiental'),
  ('Arqueologia', 'arqueologia'),
  ('Projetos sociais', 'projetos-sociais'),
  ('Meio ambiente', 'meio-ambiente'),
  ('Supervisao de obras', 'supervisao-de-obras'),
  ('Gerenciamento de projetos', 'gerenciamento-de-projetos')
ON CONFLICT (slug) DO NOTHING;
