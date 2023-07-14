-- 2023-06-12 11:29:21 : remove_boolean_status_flags

ALTER TABLE create_process
    DROP COLUMN has_service_account,
    DROP COLUMN has_cluster_access,
    DROP COLUMN has_api_key,
    DROP COLUMN has_api_key_in_vault;