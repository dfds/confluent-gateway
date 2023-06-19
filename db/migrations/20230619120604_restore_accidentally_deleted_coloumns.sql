-- 2023-06-19 12:06:04 : restore_accidentally_deleted_coloumns

ALTER TABLE create_process
  ADD COLUMN IF NOT EXISTS has_service_account BOOLEAN DEFAULT true,
  ADD COLUMN IF NOT EXISTS has_cluster_access BOOLEAN DEFAULT true,
  ADD COLUMN IF NOT EXISTS has_api_key BOOLEAN DEFAULT true,
  ADD COLUMN IF NOT EXISTS has_api_key_in_vault BOOLEAN DEFAULT true;
