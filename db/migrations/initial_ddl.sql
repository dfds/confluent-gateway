CREATE TABLE cluster (
    id varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    admin_api_endpoint varchar(255) NOT NULL,
    admin_api_key_username varchar(255) NOT NULL,
    admin_api_key_password varchar(255) NOT NULL,
    bootstrap_endpoint varchar(255) NOT NULL,

    CONSTRAINT cluster_pk PRIMARY KEY(id)
);

CREATE TABLE service_account (
    id varchar(255) NOT NULL,
    capability_root_id varchar(255) NOT NULL,
    cluster_id varchar(255) NOT NULL,
    api_key_username varchar(255) NULL,
    api_key_password varchar(255) NULL,
    created_at timestamp NOT NULL,

    CONSTRAINT service_account_pk PRIMARY KEY(id)
);


CREATE TABLE process (
    id uuid NOT NULL,
    capability_root_id varchar(255) NOT NULL,
    cluster_id varchar(255) NOT NULL,
    topic_name varchar(255) NOT NULL,
    topic_partitions int NOT NULL,
    topic_retention int NOT NULL,
    service_account_id varchar(255) NULL,
    api_key_username varchar(255) NULL,
    api_key_password varchar(255) NULL,
    api_key_created_at timestamp NULL,
    created_at timestamp NULL,
    completed_at timestamp NULL,

    CONSTRAINT process_pk PRIMARY KEY(id),
    CONSTRAINT fk_service_account FOREIGN KEY(service_account_id) REFERENCES service_account(id)
);

CREATE TABLE acl (
    id uuid NOT NULL,
    process_id uuid NOT NULL,
	resource_type varchar(255) NOT NULL,
	resource_name varchar(255) NOT NULL,
	pattern_type varchar(255) NOT NULL,
	operation_type varchar(255) NOT NULL,
	permission_type varchar(255) NOT NULL,
    created_at timestamp NULL,

    CONSTRAINT acl_pk PRIMARY KEY(id),
    CONSTRAINT fk_process FOREIGN KEY(process_id) REFERENCES process(id)
);
