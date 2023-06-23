-- 2023-06-12 11:29:21 : remove_boolean_status_flags

alter table create_process
drop column has_service_account,
drop column has_cluster_access,
drop column has_api_key,
drop column has_api_key_in_vault;