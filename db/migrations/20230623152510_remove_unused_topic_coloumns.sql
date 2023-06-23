-- 2023-06-23 15:25:10 : remove_unused_topic_coloumns

alter table create_process
drop column has_service_account,
drop column has_cluster_access,
drop column has_api_key,
drop column has_api_key_in_vault;