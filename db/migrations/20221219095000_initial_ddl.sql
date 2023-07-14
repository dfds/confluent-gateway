-- 2022-12-19 09:50:00 : initial ddl

CREATE TABLE cluster
(
    id                     VARCHAR(255) NOT NULL,
    name                   VARCHAR(255) NOT NULL,
    admin_api_endpoint     VARCHAR(255) NOT NULL,
    admin_api_key_username VARCHAR(255) NOT NULL,
    admin_api_key_password VARCHAR(255) NOT NULL,
    bootstrap_endpoint     VARCHAR(255) NOT NULL,

    CONSTRAINT cluster_pk PRIMARY KEY (id)
);

CREATE TABLE service_account
(
    id                 VARCHAR(255) NOT NULL,
    capability_root_id VARCHAR(255) NOT NULL,
    created_at         TIMESTAMP    NOT NULL,

    CONSTRAINT service_account_pk PRIMARY KEY (id)
);

CREATE TABLE cluster_access
(
    id                 UUID         NOT NULL,
    service_account_id VARCHAR(255) NULL,
    cluster_id         VARCHAR(255) NOT NULL,
    api_key_username   VARCHAR(255) NULL,
    api_key_password   VARCHAR(255) NULL,
    created_at         TIMESTAMP    NOT NULL,

    CONSTRAINT cluster_access_pk PRIMARY KEY (id),
    CONSTRAINT fk_service_account FOREIGN KEY (service_account_id) REFERENCES service_account (id)
);

CREATE TABLE acl
(
    id                UUID         NOT NULL,
    cluster_access_id UUID         NOT NULL,
    resource_type     VARCHAR(255) NOT NULL,
    resource_name     VARCHAR(255) NOT NULL,
    pattern_type      VARCHAR(255) NOT NULL,
    operation_type    VARCHAR(255) NOT NULL,
    permission_type   VARCHAR(255) NOT NULL,
    created_at        TIMESTAMP    NULL,

    CONSTRAINT acl_pk PRIMARY KEY (id),
    CONSTRAINT fk_cluster_access FOREIGN KEY (cluster_access_id) REFERENCES cluster_access (id)
);


CREATE TABLE process
(
    id                   UUID         NOT NULL,
    capability_root_id   VARCHAR(255) NOT NULL,
    cluster_id           VARCHAR(255) NOT NULL,
    topic_name           VARCHAR(255) NOT NULL,
    topic_partitions     INT          NOT NULL,
    topic_retention      INT          NOT NULL,
    has_service_account  boolean      NOT NULL DEFAULT (false),
    has_cluster_access   boolean      NOT NULL DEFAULT (false),
    has_api_key          boolean      NOT NULL DEFAULT (false),
    has_api_key_in_vault boolean      NOT NULL DEFAULT (false),
    created_at           TIMESTAMP    NULL,
    completed_at         TIMESTAMP    NULL,

    CONSTRAINT process_pk PRIMARY KEY (id)
);

