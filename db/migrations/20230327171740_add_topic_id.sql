-- 2023-03-27 17:17:40 : add topic id

alter table create_process
add column topic_id varchar(255) not null;

alter table delete_process
add column topic_id varchar(255) not null;

alter table topic
alter column id type varchar(255);
