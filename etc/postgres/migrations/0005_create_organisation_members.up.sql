CREATE TABLE IF NOT EXISTS "organisation_members" (
    "organisation_id" UUID NOT NULL,
    "member_id" UUID NOT NULL,
    "role" TEXT NOT NULL,
    PRIMARY KEY ("organisation_id", "member_id"),
    FOREIGN KEY ("organisation_id") REFERENCES "organisations"("id"),
    CONSTRAINT "organisation_member_unique" UNIQUE ("organisation_id", "member_id")
);

COMMENT ON COLUMN "organisation_members"."organisation_id" IS 'The ID of the organisation to which this membership applies.';
COMMENT ON COLUMN "organisation_members"."member_id" IS 'The ID of the user that is a member of this organisation. Should be an existing user, but it crosses a bounded context so no foreign keys.\
We may later change the storage mechanism for users and we shouldn''t know anything about that here.';
COMMENT ON COLUMN "organisation_members"."role" IS 'The role that the user has within this organisation.';
COMMENT ON CONSTRAINT "organisation_member_unique" ON "organisation_members" IS 'A user should only appear in an organisation once, with a single role.';

