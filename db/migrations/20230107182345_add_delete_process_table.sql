-- 2023-01-07 18:23:45 : add delete process table

CREATE TABLE delete_process
(
    id                 UUID         NOT NULL,
    capability_root_id VARCHAR(255) NOT NULL,
    cluster_id         VARCHAR(255) NOT NULL,
    topic_name         VARCHAR(255) NOT NULL,
    created_at         TIMESTAMP    NULL,
    completed_at       TIMESTAMP    NULL,

    CONSTRAINT delete_process_pk PRIMARY KEY (id)
);

