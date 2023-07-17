-- 2023-06-19 14:55:31 : add role binding fields

ALTER TABLE cluster
    ADD COLUMN organization_id    VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN environment_id     VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN schema_registry_id VARCHAR(255) NOT NULL DEFAULT '';