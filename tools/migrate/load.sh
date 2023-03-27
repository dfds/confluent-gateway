#!/bin/bash

PGHOST=localhost PGPORT=5432 PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'

PGHOST=localhost PGPORT=5432 PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -f topic.sql
PGHOST=localhost PGPORT=5432 PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -f service_account.sql
PGHOST=localhost PGPORT=5432 PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -f cluster_access.sql
PGHOST=localhost PGPORT=5432 PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -f acl.sql
