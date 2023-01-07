-- 2023-01-07 18:23:45 : add delete process table

CREATE TABLE delete_process (
    id uuid NOT NULL,
    capability_root_id varchar(255) NOT NULL,
    cluster_id varchar(255) NOT NULL,
    topic_name varchar(255) NOT NULL,
    created_at timestamp NULL,
    completed_at timestamp NULL,

    CONSTRAINT delete_process_pk PRIMARY KEY(id)
);

