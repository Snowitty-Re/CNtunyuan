-- ============================================================
-- 增量迁移：为走失人员补充经纬度与案件类型
-- 执行时间：2026-03-31
-- ============================================================

ALTER TABLE ty_missing_persons
  ADD COLUMN IF NOT EXISTS case_type VARCHAR(20) NOT NULL DEFAULT 'other' COMMENT '案件类型: elderly-老人, child-儿童, adult-成人, disability-残障, other-其他' AFTER description,
  ADD COLUMN IF NOT EXISTS missing_latitude DOUBLE NOT NULL DEFAULT 0 COMMENT '走失地点纬度' AFTER address,
  ADD COLUMN IF NOT EXISTS missing_longitude DOUBLE NOT NULL DEFAULT 0 COMMENT '走失地点经度' AFTER missing_latitude;

UPDATE ty_missing_persons
SET case_type = 'other'
WHERE case_type IS NULL OR case_type = '';

-- MySQL 8.0.16+ 支持 CHECK 约束
ALTER TABLE ty_missing_persons
  ADD CONSTRAINT chk_mp_case_type_v2 CHECK (case_type IN ('elderly', 'child', 'adult', 'disability', 'other')),
  ADD CONSTRAINT chk_mp_missing_latitude_v2 CHECK (missing_latitude BETWEEN -90 AND 90),
  ADD CONSTRAINT chk_mp_missing_longitude_v2 CHECK (missing_longitude BETWEEN -180 AND 180);

CREATE INDEX idx_missing_persons_case_type ON ty_missing_persons(case_type);
CREATE INDEX idx_missing_persons_geo ON ty_missing_persons(missing_latitude, missing_longitude);
