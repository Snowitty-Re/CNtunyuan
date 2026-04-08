-- 17_dialect_card_template_and_batch.sql
-- 方言卡片模板（分组/卡片）与批量录入支持

BEGIN;

CREATE TABLE IF NOT EXISTS ty_dialect_card_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_by UUID NOT NULL,
    CONSTRAINT fk_dialect_card_group_creator FOREIGN KEY (created_by) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_dialect_card_groups_status ON ty_dialect_card_groups(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dialect_card_groups_sort ON ty_dialect_card_groups(sort_order, created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dialect_card_groups_deleted_at ON ty_dialect_card_groups(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS ty_dialect_cards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    group_id UUID NOT NULL,
    content VARCHAR(200) NOT NULL,
    image_url VARCHAR(255) DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    required BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_by UUID NOT NULL,
    CONSTRAINT fk_dialect_cards_group FOREIGN KEY (group_id) REFERENCES ty_dialect_card_groups(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_dialect_cards_creator FOREIGN KEY (created_by) REFERENCES ty_users(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_dialect_cards_group ON ty_dialect_cards(group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dialect_cards_status ON ty_dialect_cards(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dialect_cards_sort ON ty_dialect_cards(group_id, sort_order, created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dialect_cards_deleted_at ON ty_dialect_cards(deleted_at) WHERE deleted_at IS NOT NULL;

ALTER TABLE ty_dialects
    ADD COLUMN IF NOT EXISTS batch_id UUID,
    ADD COLUMN IF NOT EXISTS card_group_id UUID,
    ADD COLUMN IF NOT EXISTS card_id UUID;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_dialects_card_group'
  ) THEN
    ALTER TABLE ty_dialects
      ADD CONSTRAINT fk_dialects_card_group
      FOREIGN KEY (card_group_id) REFERENCES ty_dialect_card_groups(id) ON DELETE SET NULL ON UPDATE CASCADE;
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_dialects_card'
  ) THEN
    ALTER TABLE ty_dialects
      ADD CONSTRAINT fk_dialects_card
      FOREIGN KEY (card_id) REFERENCES ty_dialect_cards(id) ON DELETE SET NULL ON UPDATE CASCADE;
  END IF;
END$$;

CREATE INDEX IF NOT EXISTS idx_dialects_batch_id ON ty_dialects(batch_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dialects_card_group_id ON ty_dialects(card_group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dialects_card_id ON ty_dialects(card_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_dialects_batch_card ON ty_dialects(batch_id, card_id) WHERE batch_id IS NOT NULL AND card_id IS NOT NULL AND deleted_at IS NULL;

CREATE TRIGGER update_dialect_card_groups_updated_at BEFORE UPDATE ON ty_dialect_card_groups
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dialect_cards_updated_at BEFORE UPDATE ON ty_dialect_cards
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
