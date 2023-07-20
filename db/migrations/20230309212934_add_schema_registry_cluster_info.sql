-- 2023-03-09 21:29:34 : add schema registry cluster info

ALTER TABLE cluster
    ADD COLUMN schema_registry_api_endpoint     VARCHAR(255) NULL,
    ADD COLUMN schema_registry_api_key_username VARCHAR(255) NULL,
    ADD COLUMN schema_registry_api_key_password VARCHAR(255) NULL;
