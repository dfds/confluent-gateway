-- 2023-03-27 18:11:59 : add schema

CREATE TABLE schema_process (
    id uuid NOT NULL,
    cluster_id varchar(255) NOT NULL,
    message_contract_id varchar(255) NOT NULL,
    topic_id varchar(255) NOT NULL,
    message_type varchar(255) NOT NULL,
    description varchar(255) NOT NULL,
    subject varchar(255) NOT NULL,
    schema text not null,
    created_at timestamp not null,
    completed_at timestamp null,

    CONSTRAINT schema_process_pk PRIMARY KEY(id)
);

-- create table schema (
--     id uuid not null,
--     capability_id varchar(255) not null,
--     cluster_id varchar(255) not null,
--     topic_name varchar(255) not null,
--     subject varchar(255) not null,
--     schema text not null,
--     created_at timestamp not null,
--     registered_at timestamp null,

--     constraint schema_pk primary key(id)
-- );