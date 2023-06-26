-- 2023-06-19 14:55:31 : add role binding fields

alter table cluster
    add column organization_id    varchar(255) not null default '',
    add column environment_id     varchar(255) not null default '',
    add column schema_registry_id varchar(255) not null default '';