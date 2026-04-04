-- ============================================================
-- 07. 方言表结构对齐（兼容历史库）
-- - 补齐 missing_person_id
-- - 补齐采集地址与经纬度
-- ============================================================

BEGIN;

ALTER TABLE ty_dialects
  ADD COLUMN IF NOT EXISTS missing_person_id UUID,
  ADD COLUMN IF NOT EXISTS collect_address VARCHAR(255),
  ADD COLUMN IF NOT EXISTS collect_latitude DECIMAL(10,7),
  ADD COLUMN IF NOT EXISTS collect_longitude DECIMAL(10,7);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_dialect_missing_person'
  ) THEN
    ALTER TABLE ty_dialects
      ADD CONSTRAINT fk_dialect_missing_person
      FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id)
      ON DELETE SET NULL ON UPDATE CASCADE;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_dialects_missing_person_id
  ON ty_dialects(missing_person_id)
  WHERE deleted_at IS NULL AND missing_person_id IS NOT NULL;

COMMIT;
