-- 2023-01-07 00:46:21 : add topic table

CREATE TABLE topic
(
    id                 UUID         NOT NULL,
    capability_root_id VARCHAR(255) NOT NULL,
    cluster_id         VARCHAR(255) NOT NULL,
    name               VARCHAR(255) NOT NULL,
    partitions         INT          NOT NULL,
    retention          BIGINT       NOT NULL,
    created_at         TIMESTAMP    NOT NULL,

    CONSTRAINT topic_pk PRIMARY KEY (id)
);
