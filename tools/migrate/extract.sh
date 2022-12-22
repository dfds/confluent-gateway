#!/bin/bash

if [ "$1" != "run" ]; then

	op run --no-masking --env-file=.env -- ./$0 run
	exit $?

fi

cloudkey=$(echo -n "${CLOUD_API_USERNAME}:${CLOUD_API_PASSWORD}" | base64)
clusterkey_prod=$(echo -n "${CLUSTER_PROD_USERNAME}:${CLUSTER_PROD_PASSWORD}" | base64)
clusterkey_dev=$(echo -n "${CLUSTER_DEV_USERNAME}:${CLUSTER_DEV_PASSWORD}" | base64)

function get_topics() {
	(psql --csv <<EOF
select c."Name" as "capability_name",
       c."RootId" as "capability_root_id",
       c."Deleted" is null as active,
       k."ClusterId" as cluster_id,
       t."Name" as topic_name,
       t."Partitions" as partitions,
       t."Created" as created,
       ((t."Configurations"::json) ->> 'retention.ms')::bigint as retention
from "KafkaTopic" t
inner join "Capability" c on t."CapabilityId" = c."Id"
inner join "KafkaCluster" k on k."Id" = t."KafkaClusterId"
EOF
) | sed 's/,t,/,true,/g' | sed 's/,f,/,false,/g' > topics.csv
}


function get_service_accounts() {
	url="https://api.confluent.cloud/iam/v2/service-accounts"

	echo "serviceaccount_id,name" > service_accounts.csv

	while [ ! -z ${url} ]; do
		echo ${url}

		content=$(curl -sH "Authorization: Basic ${cloudkey}" ${url})
		echo ${content} | jq -r '.data[] | [.id, .display_name] | @csv' >> service_accounts.csv
		url=$(echo ${content} | jq  -r '.metadata.next | select(.!=null)')
	done
}

function get_api_keys() {
	url=https://api.confluent.cloud/iam/v2/api-keys

	echo "owner_id,resource_id,resource_kind,key" > api_keys.csv

	while [ ! -z ${url} ]; do
		echo ${url}

		content=$(curl -sH "Authorization: Basic ${cloudkey}" ${url})
		echo ${content} | jq -r '.data[] | [.spec.owner.id,  .spec.resource.id, .spec.resource.kind, .id] | @csv' >> api_keys.csv
		url=$(echo ${content} | jq  -r '.metadata.next | select(.!=null)')
	done
}

function get_users() {
	echo "user_id,capability_name,serviceaccount_id" > users.csv

	url="https://confluent.cloud/api/service_accounts"
	content=$(curl -sH "Authorization: Basic ${cloudkey}" ${url})
	echo ${content} | jq -r '.users[] | [.id, .service_name, .resource_id] | @csv' | \
		awk -F',' '{ print "User:"$0 }' >> users.csv
}

function get_acl() {
	echo "cluster_id,principal,resource_type,resource_name,pattern_type,operation,permission" > acl.csv

	url="https://pkc-e8wrm.eu-central-1.aws.confluent.cloud/kafka/v3/clusters/lkc-4npj6/acls"
	curl -H "Authorization: Basic ${clusterkey_prod}" ${url} | \
		jq -r '.data[] | [.cluster_id, .principal, .resource_type, .resource_name, .pattern_type, .operation, .permission] | @csv' >> acl.csv

	url="https://pkc-e8mp5.eu-west-1.aws.confluent.cloud/kafka/v3/clusters/lkc-3wqzw/acls"
	curl -H "Authorization: Basic ${clusterkey_dev}" ${url}  | \
		jq -r '.data[] | [.cluster_id, .principal, .resource_type, .resource_name, .pattern_type, .operation, .permission] | @csv' >> acl.csv
}

function get_ssm() {
	echo "key,secret" > ssm.csv

	aws ssm get-parameters-by-path \
		--path /capabilities \
		--recursive \
		--max-items 1000 \
		--with-decryption \
		--query 'Parameters[*].Value' | \
		jq -r '.[] | fromjson? | flatten | @csv' | \
		awk -F',' '!seen[$1]++' \
		>> ssm.csv

}

get_topics
get_service_accounts
get_api_keys
get_users
get_acl
get_ssm
