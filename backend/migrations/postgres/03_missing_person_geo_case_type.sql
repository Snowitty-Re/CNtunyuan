-- ============================================================
-- 增量迁移：为走失人员补充经纬度与案件类型
-- 执行时间：2026-03-31
-- ============================================================

BEGIN;

ALTER TABLE ty_missing_persons
  ADD COLUMN IF NOT EXISTS case_type VARCHAR(20) NOT NULL DEFAULT 'other',
  ADD COLUMN IF NOT EXISTS missing_latitude DOUBLE PRECISION NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS missing_longitude DOUBLE PRECISION NOT NULL DEFAULT 0;

UPDATE ty_missing_persons
SET case_type = 'other'
WHERE case_type IS NULL OR case_type = '';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_missing_person_case_type'
  ) THEN
    ALTER TABLE ty_missing_persons
      ADD CONSTRAINT chk_missing_person_case_type
      CHECK (case_type IN ('elderly', 'child', 'adult', 'disability', 'other'));
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_missing_person_latitude'
  ) THEN
    ALTER TABLE ty_missing_persons
      ADD CONSTRAINT chk_missing_person_latitude
      CHECK (missing_latitude BETWEEN -90 AND 90);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_missing_person_longitude'
  ) THEN
    ALTER TABLE ty_missing_persons
      ADD CONSTRAINT chk_missing_person_longitude
      CHECK (missing_longitude BETWEEN -180 AND 180);
  END IF;
END $$;

COMMENT ON COLUMN ty_missing_persons.case_type IS '案件类型: elderly-老人, child-儿童, adult-成人, disability-残障, other-其他';
COMMENT ON COLUMN ty_missing_persons.missing_latitude IS '走失地点纬度';
COMMENT ON COLUMN ty_missing_persons.missing_longitude IS '走失地点经度';

CREATE INDEX IF NOT EXISTS idx_missing_persons_case_type ON ty_missing_persons(case_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_missing_persons_geo ON ty_missing_persons(missing_latitude, missing_longitude) WHERE deleted_at IS NULL;

COMMIT;
