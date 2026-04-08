-- 15_authz_policy_request_scope.sql
-- 权限策略申请：组织作用域审批

ALTER TABLE ty_authz_policy_change_requests
    ADD COLUMN IF NOT EXISTS scope_type VARCHAR(20) NOT NULL DEFAULT 'global',
    ADD COLUMN IF NOT EXISTS target_org_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chk_authz_policy_change_requests_scope_type'
    ) THEN
        ALTER TABLE ty_authz_policy_change_requests
            ADD CONSTRAINT chk_authz_policy_change_requests_scope_type
            CHECK (scope_type IN ('global', 'org'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_authz_policy_change_requests_scope_type
    ON ty_authz_policy_change_requests(scope_type)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_authz_policy_change_requests_target_org_id
    ON ty_authz_policy_change_requests(target_org_id)
    WHERE deleted_at IS NULL;
