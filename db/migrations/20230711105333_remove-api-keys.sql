-- 2023-07-11 10:53:33 : remove-api-keys

ALTER TABLE cluster_access
    DROP COLUMN IF EXISTS api_key_username,
    DROP COLUMN IF EXISTS api_key_password,
    DROP COLUMN IF EXISTS schema_registry_api_key_username,
    DROP COLUMN IF EXISTS schema_registry_api_key_password;