
{{- define "confluent-gateway.serviceaccount.awsRoleArn" -}}
{{- if .Values.serviceAccount.awsRoleArn }}
{{- .Values.serviceAccount.awsRoleArn }}
{{- else }}
{{- "arn:aws:iam::${ECR_AWS_ACCOUNT_ID}:role/CreateECRRepos" }}
{{- end }}
{{- end }}