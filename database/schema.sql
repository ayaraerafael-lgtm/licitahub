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
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT tenders_status_chk CHECK (status IN (
    'draft', 'published', 'under_review', 'suspended', 'challenged', 'occurred', 'closed', 'cancelled'
  ))
);

CREATE TABLE IF NOT EXISTS tender_files (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tender_id uuid NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
  media_id uuid REFERENCES media_files(id),
  file_type varchar(60) NOT NULL,
  title varchar(220) NOT NULL,
  file_url text,
  mime_type varchar(120),
  is_downloadable boolean NOT NULL DEFAULT true,
  uploaded_by_user_id uuid REFERENCES users(id),
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
