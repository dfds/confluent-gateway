-- 2023-01-05 15:03:23 : add user account id

ALTER TABLE service_account
    ADD COLUMN user_account_id INT DEFAULT(0);

ALTER TABLE cluster_access
    ADD COLUMN user_account_id INT DEFAULT(0);


