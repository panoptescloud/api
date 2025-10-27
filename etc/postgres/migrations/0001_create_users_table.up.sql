CREATE TABLE IF NOT EXISTS "users"(
   "id" UUID PRIMARY KEY,
   "email" TEXT NOT NULL,
   "name" TEXT NOT NULL
);

COMMENT ON COLUMN "users"."id" IS 'Primary key and unique ID for a user in the system.';
COMMENT ON COLUMN "users"."email" IS 'Email address for the given user, used primarily for contacting them.';
COMMENT ON COLUMN "users"."name" IS 'Preferred name to display when referencing this user, free text could be a real name, partial name, nickname etc.';
