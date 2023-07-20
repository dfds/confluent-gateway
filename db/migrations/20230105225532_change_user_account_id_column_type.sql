-- 2023-01-05 22:55:32 : change user_account_id column type

ALTER TABLE service_account
    ALTER COLUMN user_account_id DROP DEFAULT,
    ALTER COLUMN user_account_id DROP NOT NULL,
    ALTER COLUMN user_account_id TYPE VARCHAR(50) USING
        CASE WHEN user_account_id = 0 THEN NULL ELSE CONCAT('User:', user_account_id) END;

ALTER TABLE cluster_access
    ALTER COLUMN user_account_id DROP DEFAULT,
    ALTER COLUMN user_account_id DROP NOT NULL,
    ALTER COLUMN user_account_id TYPE VARCHAR(50) USING
        CASE WHEN user_account_id = 0 THEN NULL ELSE CONCAT('User:', user_account_id) END;
