ALTER TABLE credentials
    ADD COLUMN totp_secret TEXT NOT NULL DEFAULT '';

PRAGMA user_version = 3;
