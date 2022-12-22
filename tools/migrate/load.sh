#!/bin/bash

PGHOST=localhost PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'

PGHOST=localhost PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -f process.sql
PGHOST=localhost PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -f service_account.sql
PGHOST=localhost PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -f cluster_access.sql
PGHOST=localhost PGDATABASE=db PGUSER=postgres PGPASSWORD=p psql -f acl.sql
