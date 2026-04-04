-- ============================================================
-- 07. 方言表结构对齐（兼容历史库）
-- - 补齐 missing_person_id
-- - 补齐采集地址与经纬度
-- ============================================================

ALTER TABLE ty_dialects
  ADD COLUMN IF NOT EXISTS missing_person_id CHAR(36) COMMENT '关联走失人员ID',
  ADD COLUMN IF NOT EXISTS collect_address VARCHAR(255) COMMENT '采集地址',
  ADD COLUMN IF NOT EXISTS collect_latitude DECIMAL(10,7) COMMENT '采集纬度',
  ADD COLUMN IF NOT EXISTS collect_longitude DECIMAL(10,7) COMMENT '采集经度';

SET @fk_exists := (
  SELECT COUNT(*)
  FROM information_schema.TABLE_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ty_dialects'
    AND CONSTRAINT_NAME = 'fk_dialect_missing_person'
    AND CONSTRAINT_TYPE = 'FOREIGN KEY'
);

SET @fk_sql := IF(
  @fk_exists = 0,
  'ALTER TABLE ty_dialects ADD CONSTRAINT fk_dialect_missing_person FOREIGN KEY (missing_person_id) REFERENCES ty_missing_persons(id) ON DELETE SET NULL ON UPDATE CASCADE',
  'SELECT 1'
);

PREPARE stmt FROM @fk_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ty_dialects'
    AND INDEX_NAME = 'idx_dialects_missing_person_id'
);

SET @idx_sql := IF(
  @idx_exists = 0,
  'CREATE INDEX idx_dialects_missing_person_id ON ty_dialects(missing_person_id)',
  'SELECT 1'
);

PREPARE idx_stmt FROM @idx_sql;
EXECUTE idx_stmt;
DEALLOCATE PREPARE idx_stmt;
