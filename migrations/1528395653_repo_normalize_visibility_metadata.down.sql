BEGIN;

ALTER TABLE repo DROP COLUMN IF EXISTS private;

COMMIT;