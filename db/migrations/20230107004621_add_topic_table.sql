-- 2023-01-07 00:46:21 : add topic table

CREATE TABLE topic (
    id uuid NOT NULL,
    capability_root_id varchar(255) NOT NULL,
    cluster_id varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    partitions int NOT NULL,
    retention bigint NOT NULL,
    created_at timestamp NOT NULL,

    CONSTRAINT topic_pk PRIMARY KEY(id)
);
