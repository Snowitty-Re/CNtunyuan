-- 17_dialect_card_template_and_batch.sql
-- 方言卡片模板（分组/卡片）与批量录入支持

CREATE TABLE IF NOT EXISTS ty_dialect_card_groups (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by CHAR(36) NOT NULL,
    CONSTRAINT chk_dialect_card_group_status CHECK (status IN ('active', 'inactive')),
    CONSTRAINT fk_dialect_card_group_creator FOREIGN KEY (created_by) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='方言卡片分组表';

CREATE INDEX idx_dialect_card_groups_status ON ty_dialect_card_groups(status);
CREATE INDEX idx_dialect_card_groups_sort ON ty_dialect_card_groups(sort_order, created_at);
CREATE INDEX idx_dialect_card_groups_deleted_at ON ty_dialect_card_groups(deleted_at);

CREATE TABLE IF NOT EXISTS ty_dialect_cards (
    id CHAR(36) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    group_id CHAR(36) NOT NULL,
    content VARCHAR(200) NOT NULL,
    image_url VARCHAR(255) DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    required TINYINT(1) NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by CHAR(36) NOT NULL,
    CONSTRAINT chk_dialect_card_status CHECK (status IN ('active', 'inactive')),
    CONSTRAINT fk_dialect_cards_group FOREIGN KEY (group_id) REFERENCES ty_dialect_card_groups(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_dialect_cards_creator FOREIGN KEY (created_by) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='方言录入卡片表';

CREATE INDEX idx_dialect_cards_group ON ty_dialect_cards(group_id);
CREATE INDEX idx_dialect_cards_status ON ty_dialect_cards(status);
CREATE INDEX idx_dialect_cards_sort ON ty_dialect_cards(group_id, sort_order, created_at);
CREATE INDEX idx_dialect_cards_deleted_at ON ty_dialect_cards(deleted_at);

ALTER TABLE ty_dialects
    ADD COLUMN IF NOT EXISTS batch_id CHAR(36) NULL COMMENT '批次ID',
    ADD COLUMN IF NOT EXISTS card_group_id CHAR(36) NULL COMMENT '卡片分组ID',
    ADD COLUMN IF NOT EXISTS card_id CHAR(36) NULL COMMENT '卡片ID';

SET @dialect_group_fk_exists := (
    SELECT COUNT(*)
    FROM information_schema.TABLE_CONSTRAINTS
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'ty_dialects'
      AND CONSTRAINT_NAME = 'fk_dialects_card_group'
);
SET @dialect_group_fk_sql := IF(
    @dialect_group_fk_exists = 0,
    'ALTER TABLE ty_dialects ADD CONSTRAINT fk_dialects_card_group FOREIGN KEY (card_group_id) REFERENCES ty_dialect_card_groups(id) ON DELETE SET NULL ON UPDATE CASCADE',
    'SELECT 1'
);
PREPARE stmt FROM @dialect_group_fk_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @dialect_card_fk_exists := (
    SELECT COUNT(*)
    FROM information_schema.TABLE_CONSTRAINTS
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'ty_dialects'
      AND CONSTRAINT_NAME = 'fk_dialects_card'
);
SET @dialect_card_fk_sql := IF(
    @dialect_card_fk_exists = 0,
    'ALTER TABLE ty_dialects ADD CONSTRAINT fk_dialects_card FOREIGN KEY (card_id) REFERENCES ty_dialect_cards(id) ON DELETE SET NULL ON UPDATE CASCADE',
    'SELECT 1'
);
PREPARE stmt FROM @dialect_card_fk_sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

CREATE INDEX idx_dialects_batch_id ON ty_dialects(batch_id);
CREATE INDEX idx_dialects_card_group_id ON ty_dialects(card_group_id);
CREATE INDEX idx_dialects_card_id ON ty_dialects(card_id);
CREATE UNIQUE INDEX uk_dialects_batch_card ON ty_dialects(batch_id, card_id);
