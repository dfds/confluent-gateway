-- 2023-06-15 12:36:04 : add_schemas_deleted_at

ALTER TABLE delete_process
    ADD COLUMN schemas_deleted_at TIMESTAMP NULL;