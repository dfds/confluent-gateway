-- 2023-03-27 17:17:40 : add topic id

ALTER TABLE create_process
    ADD COLUMN topic_id VARCHAR(255) NOT NULL;

ALTER TABLE delete_process
    ADD COLUMN topic_id VARCHAR(255) NOT NULL;

ALTER TABLE topic
    ALTER COLUMN id TYPE VARCHAR(255);
