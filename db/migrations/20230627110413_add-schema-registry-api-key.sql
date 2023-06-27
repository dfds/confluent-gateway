-- 2023-06-27 11:04:13 : add-schema-registry-api-key

alter table cluster_access
    add column schema_registry_api_key_username varchar(255) NULL,
       add column   schema_registry_api_key_password varchar(255) NULL;
