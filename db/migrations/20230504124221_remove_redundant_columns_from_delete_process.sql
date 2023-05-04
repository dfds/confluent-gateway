-- 2023-05-04 12:42:21 : remove redundant columns from delete process

alter table delete_process
drop column capability_id,
drop column cluster_id,
drop column topic_name;

