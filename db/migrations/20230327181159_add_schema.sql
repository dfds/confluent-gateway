-- 2023-03-27 18:11:59 : add schema

CREATE TABLE schema_process
(
    id                  UUID         NOT NULL,
    cluster_id          VARCHAR(255) NOT NULL,
    message_contract_id VARCHAR(255) NOT NULL,
    topic_id            VARCHAR(255) NOT NULL,
    message_type        VARCHAR(255) NOT NULL,
    description         VARCHAR(255) NOT NULL,
    subject             VARCHAR(255) NOT NULL,
    schema              TEXT         NOT NULL,
    created_at          TIMESTAMP    NOT NULL,
    completed_at        TIMESTAMP    NULL,

    CONSTRAINT schema_process_pk PRIMARY KEY (id)
);