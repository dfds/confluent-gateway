-- 2023-06-23 15:25:10 : remove_unused_topic_coloumns

ALTER TABLE create_process
    DROP COLUMN has_service_account,
    DROP COLUMN has_cluster_access,
    DROP COLUMN has_api_key,
    DROP COLUMN has_api_key_in_vault;