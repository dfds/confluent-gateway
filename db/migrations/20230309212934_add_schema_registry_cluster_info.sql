-- 2023-03-09 21:29:34 : add schema registry cluster info

alter table cluster
    add column schema_registry_api_endpoint varchar(255) null,
    add column schema_registry_api_key_username varchar(255) null,
    add column schema_registry_api_key_password varchar(255) null;
