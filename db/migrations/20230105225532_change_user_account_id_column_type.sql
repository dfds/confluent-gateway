-- 2023-01-05 22:55:32 : change user_account_id column type

alter table service_account
alter column user_account_id drop default,
alter column user_account_id drop not null,
alter column user_account_id type varchar(50) using
case when user_account_id=0 then null else concat('User:', user_account_id) end;

alter table cluster_access
alter column user_account_id drop default,
alter column user_account_id drop not null,
alter column user_account_id type varchar(50) using
case when user_account_id=0 then null else concat('User:', user_account_id) end;
