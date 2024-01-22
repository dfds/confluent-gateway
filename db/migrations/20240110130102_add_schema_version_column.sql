-- 2024-01-10 13:01:02 : add schema version

ALTER TABLE schema_process
    ADD COLUMN schema_version int NOT NULL DEFAULT 1;