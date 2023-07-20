-- 2023-05-04 12:42:21 : remove redundant columns from delete process

ALTER TABLE delete_process
    DROP COLUMN capability_id,
    DROP COLUMN cluster_id,
    DROP COLUMN topic_name;

