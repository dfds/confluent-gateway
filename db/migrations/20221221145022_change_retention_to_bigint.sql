-- 2022-12-21 14:50:22 : change retention to bitint

ALTER TABLE process
    ALTER COLUMN topic_retention TYPE BIGINT;


