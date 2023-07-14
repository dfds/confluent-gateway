-- 2023-03-09 11:06:56 : rename capability root id

ALTER TABLE service_account
    RENAME COLUMN capability_root_id TO capability_id;

ALTER TABLE create_process
    RENAME COLUMN capability_root_id TO capability_id;

ALTER TABLE delete_process
    RENAME COLUMN capability_root_id TO capability_id;

ALTER TABLE topic
    RENAME COLUMN capability_root_id TO capability_id;

