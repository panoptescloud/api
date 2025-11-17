CREATE TABLE IF NOT EXISTS "organisations" (
    "id" UUID NOT NULL PRIMARY KEY,
    "name" TEXT NOT NULL,
    CONSTRAINT "unique_name" UNIQUE ("name")
);

COMMENT ON COLUMN "organisations"."name" IS 'The name of the organisation, must be unique across the system.';
COMMENT ON CONSTRAINT "unique_name" ON "organisations" IS 'An organisation name should always be unique across the system.';