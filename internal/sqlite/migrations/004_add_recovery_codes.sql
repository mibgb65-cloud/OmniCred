ALTER TABLE credentials
    ADD COLUMN recovery_codes TEXT NOT NULL DEFAULT '[]';

PRAGMA user_version = 4;
