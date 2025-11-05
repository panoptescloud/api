CREATE TABLE IF NOT EXISTS "refresh_tokens" (
    "id" UUID NOT NULL PRIMARY KEY,
    "user_id" UUID NOT NULL,
    "value" TEXT NOT NULL,
    "issued_at" TIMESTAMPTZ NOT NULL,
    "expires_at" TIMESTAMPTZ NOT NULL,
    FOREIGN KEY ("user_id") REFERENCES "users"("id"),
    CONSTRAINT "unique_refresh_token_value" UNIQUE ("value")
);

COMMENT ON COLUMN "refresh_tokens"."user_id" IS 'The ID of the user that is associated with this refresh token.';
COMMENT ON COLUMN "refresh_tokens"."value" IS 'The actual refresh token that can grant a new access token. This is hashed before being stored so won''t be usable outside of the app.';
COMMENT ON COLUMN "refresh_tokens"."issued_at" IS 'When the refresh token was issued.';
COMMENT ON COLUMN "refresh_tokens"."issued_at" IS 'When the refresh token expires. A periodic cron job should delete anything that has already expired.';
COMMENT ON CONSTRAINT "unique_refresh_token_value" ON "refresh_tokens" IS 'A refresh token should always be unique across the system.';