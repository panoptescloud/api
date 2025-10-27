CREATE TABLE IF NOT EXISTS "github_users" (
    "user_id" UUID NOT NULL,
    "node_id" TEXT NOT NULL,
    PRIMARY KEY ("user_id", "node_id"),
    FOREIGN KEY ("user_id") REFERENCES "users"("id")
);


COMMENT ON COLUMN "github_users"."node_id" IS 'The node ID of this user from github, so we can map them back when necessary. 
Mostly necessary during the login process.';