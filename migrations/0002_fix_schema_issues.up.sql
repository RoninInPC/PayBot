-- Fix data type mismatch: payments.user_tg_id should be BIGINT not INTEGER
ALTER TABLE payments ALTER COLUMN user_tg_id TYPE BIGINT;

-- Add default value for subscriptions.status
ALTER TABLE subscriptions ALTER COLUMN status SET DEFAULT 'active';
