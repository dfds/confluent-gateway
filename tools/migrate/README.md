# Confluent Migration (ETL)

To setup existing Confluent state for the Confluent Gateway please run the two scripts:

```
./extract.sh
./transform.sh
./load.sh
```

## Script Requirements

- op (1password CLI)
- aws (AWS CLI)
- jq
- xsv
- psql
- authenticated session against the oxygen account

## `extract.sh`

Extract the following Confluent State locally:

1. Topics configuration from the Capability Service
1. Confluent Service Accounts
1. Confluent API Keys (without secret)
1. Confluent Users (to match ACLs with Service Accounts)
1. Confluent ACLs
1. API key secrets from AWS Parameter Store

## `transform.sh`

Generates SQL `insert` statements for the following tables:

1. `process.sql` (as completed)
1. `service_account.sql`
1. `cluster_access.sql`
1. `acl.sql`

## `load.sh`

Loads SQL `insert` statements for the following tables:

1. `process`
1. `service_account`
1. `cluster_access`
1. `acl`
