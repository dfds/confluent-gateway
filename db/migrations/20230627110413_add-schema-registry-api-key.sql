-- 2023-06-27 11:04:13 : add-schema-registry-api-key

ALTER TABLE cluster_access
    ADD COLUMN schema_registry_api_key_username VARCHAR(255) NULL,
       ADD COLUMN   schema_registry_api_key_password VARCHAR(255) NULL;
