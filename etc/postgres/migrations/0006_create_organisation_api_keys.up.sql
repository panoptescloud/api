CREATE TABLE IF NOT EXISTS "organisation_api_keys" (
    "id" UUID NOT NULL PRIMARY KEY,
    "organisation_id" UUID NOT NULL,
    "name" TEXT NOT NULL,
    "token" TEXT NOT NULL,
    FOREIGN KEY ("organisation_id") REFERENCES "organisations"("id"),
    CONSTRAINT "organisation_api_key_name_unique" UNIQUE ("organisation_id", "name"),
    CONSTRAINT "organisation_api_key_token_unique" UNIQUE ("token")
);

COMMENT ON COLUMN "organisation_api_keys"."organisation_id" IS 'The ID of the organisation to which api key belongs.';
COMMENT ON COLUMN "organisation_api_keys"."name" IS 'Simply a reference to keep track of the different API keys that may be created. Must be unique across the organisation.';
COMMENT ON COLUMN "organisation_api_keys"."token" IS 'The token that is used for access; this will be encrypted and useless outside of the application code.';
COMMENT ON CONSTRAINT "organisation_api_key_name_unique" ON "organisation_api_keys" IS 'An api key name must be unique across an organisation.';
COMMENT ON CONSTRAINT "organisation_api_key_token_unique" ON "organisation_api_keys" IS 'A token must be unique across the whole system.';

