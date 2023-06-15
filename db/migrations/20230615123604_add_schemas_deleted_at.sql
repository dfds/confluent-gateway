-- 2023-06-15 12:36:04 : add_schemas_deleted_at

alter table delete_process
add column schemas_deleted_at timestamp NULL;