-- Revert subscriptions.status default
ALTER TABLE subscriptions ALTER COLUMN status DROP DEFAULT;

-- Revert payments.user_tg_id to INTEGER (with potential data loss warning)
ALTER TABLE payments ALTER COLUMN user_tg_id TYPE INTEGER;
