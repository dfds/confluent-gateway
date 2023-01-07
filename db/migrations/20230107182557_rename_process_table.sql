-- 2023-01-07 18:25:57 : rename process table

ALTER TABLE process RENAME TO create_process;

ALTER TABLE create_process RENAME CONSTRAINT process_pk TO create_process_pk;