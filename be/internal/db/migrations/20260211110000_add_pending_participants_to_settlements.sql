-- +goose Up
-- +goose StatementBegin
ALTER TABLE settlements
    ALTER COLUMN payer_id DROP NOT NULL,
    ALTER COLUMN payee_id DROP NOT NULL,
    ADD COLUMN payer_pending_user_id UUID REFERENCES pending_users(id),
    ADD COLUMN payee_pending_user_id UUID REFERENCES pending_users(id);

ALTER TABLE settlements
    ADD CONSTRAINT settlements_payer_participant_check
        CHECK ((payer_id IS NOT NULL) <> (payer_pending_user_id IS NOT NULL)),
    ADD CONSTRAINT settlements_payee_participant_check
        CHECK ((payee_id IS NOT NULL) <> (payee_pending_user_id IS NOT NULL)),
    ADD CONSTRAINT settlements_distinct_participants_check
        CHECK (
            COALESCE('u:' || payer_id::text, 'p:' || payer_pending_user_id::text) <>
            COALESCE('u:' || payee_id::text, 'p:' || payee_pending_user_id::text)
        );

CREATE INDEX idx_settlements_payer_pending_user_id
    ON settlements(payer_pending_user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_settlements_payee_pending_user_id
    ON settlements(payee_pending_user_id)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_settlements_payee_pending_user_id;
DROP INDEX IF EXISTS idx_settlements_payer_pending_user_id;

ALTER TABLE settlements
    DROP CONSTRAINT IF EXISTS settlements_distinct_participants_check,
    DROP CONSTRAINT IF EXISTS settlements_payee_participant_check,
    DROP CONSTRAINT IF EXISTS settlements_payer_participant_check,
    DROP COLUMN IF EXISTS payee_pending_user_id,
    DROP COLUMN IF EXISTS payer_pending_user_id,
    ALTER COLUMN payee_id SET NOT NULL,
    ALTER COLUMN payer_id SET NOT NULL;
-- +goose StatementEnd
