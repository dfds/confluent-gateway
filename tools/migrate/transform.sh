#!/bin/bash

# Process

function get_topic_sql() {

    # Topic.csv
    # 1   capability_name
    # 2   capability_root_id
    # 3   active
    # 4   cluster_id
    # 5   topic_name
    # 6   partitions
    # 7   created
    # 8   retention
    # 9   topic_id

    cat topics.csv | \
        awk -F',' '{ printf "('\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', %d, %d, '\''%s'\'');\n", $9, $2, $4, $5, $6, $8, $7 }' | \
        sed 's/^/insert into topic (id, capability_id, cluster_id, name, partitions, retention, created_at) values /' | \
        tail -n +2

}

# ServiceAccount

function get_service_account_sql() {

    xsv join --no-case 'name' service_accounts.csv 'capability_name' topics.csv | \
        xsv join --no-case 'serviceaccount_id' /dev/stdin 'serviceaccount_id' users.csv | \
        xsv select 'serviceaccount_id,user_id,capability_root_id,created' | \
        xsv sort --select 'serviceaccount_id,created' | \
        awk -F',' '!seen[$1]++' | \
        awk -F',' '{ printf "('\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'');\n", $1, $2, $3, $4 }' | \
        sed 's/^/insert into service_account (id, user_account_id, capability_id, created_at) values /' | \
        tail -n +2

}

# ClusterAccess

function get_cluster_access_csv() {

    xsv join --no-case 'name' service_accounts.csv 'capability_name' topics.csv | \
        xsv join --no-case 'serviceaccount_id' /dev/stdin 'serviceaccount_id' users.csv | \
        xsv select 'serviceaccount_id,user_id,cluster_id,created' | \
        xsv join --no-case 'serviceaccount_id,cluster_id' /dev/stdin 'owner_id,resource_id' api_keys.csv | \
        xsv join --no-case 'key' /dev/stdin 'key' ssm.csv | \
        xsv select 'serviceaccount_id,user_id,cluster_id,key[0],secret,created' | \
        xsv sort --select 'serviceaccount_id,cluster_id' | \
        awk -F',' '!seen[$1$2$3]++' | \
        awk -v OFS=, '("uuidgen" | getline uuid) > 0 { if (NR!=1) { print uuid,$0 } else { print "cluster_access_id",$0 } } {close("uuidgen")}'
        
}

function get_cluster_access_sql() {

    cat cluster_access.csv | \
        awk -F',' '{ printf "('\''%s'\'','\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'');\n", $1, $2, $3, $4, $5, $6, $7 }' | \
        sed 's/^/insert into cluster_access (id, service_account_id, user_account_id, cluster_id, api_key_username, api_key_password, created_at) values /' | \
        tail -n +2

}

# ACL

function get_acl_sql() {

    xsv join --no-case 'user_id' users.csv 'principal' acl.csv | \
        xsv join --no-case 'serviceaccount_id,cluster_id' /dev/stdin 'serviceaccount_id,cluster_id' cluster_access.csv | \
        xsv select 'cluster_access_id,resource_type,resource_name,pattern_type,operation,permission,created' | \
        awk -v OFS=, '("uuidgen" | getline uuid) > 0 { if (NR!=1) { print uuid,$0 } else { print "id",$0 } } {close("uuidgen")}' | \
        awk -F',' '{ printf "('\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'', '\''%s'\'');\n", $1, $2, $3, $4, $5, $6, $7, $8 }' | \
        sed 's/^/insert into acl (id, cluster_access_id, resource_type, resource_name, pattern_type, operation_type, permission_type, created_at) values /' | \
        sed 's/'\'\''/'\'\'\''/g' | \
        tail -n +2

}

get_topic_sql > topic.sql
get_service_account_sql > service_account.sql
get_cluster_access_csv > cluster_access.csv
get_cluster_access_sql > cluster_access.sql
get_acl_sql > acl.sql

