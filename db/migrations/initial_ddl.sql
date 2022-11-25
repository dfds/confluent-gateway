CREATE TABLE cluster (
    id varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    admin_api_endpoint varchar(255) NOT NULL,
    admin_api_key_username varchar(255) NOT NULL,
    admin_api_key_password varchar(255) NOT NULL,
    bootstrap_endpoint varchar(255) NOT NULL,

    CONSTRAINT cluster_pk PRIMARY KEY(id)
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

    CONSTRAINT process_pk PRIMARY KEY(id)
);

CREATE TABLE acl (
    id  INT GENERATED ALWAYS AS IDENTITY,
    process_id uuid NOT NULL,
	resource_type varchar(255) NOT NULL,
	resource_name varchar(255) NOT NULL,
	pattern_type varchar(255) NOT NULL,
	operation_type varchar(255) NOT NULL,
	permission_type varchar(255) NOT NULL,

    CONSTRAINT fk_process FOREIGN KEY(process_id)  REFERENCES process(id)
);
