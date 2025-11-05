#!/bin/bash
set -e

# This file is named in this way to avoid messing up the timescale initialisation
# as that uses the same entrypoint functionality to set up all the extensions and
# such.
# The files used by that are prefixed with 000 and 001, so this filename should
# run after timescales init.
# See: https://github.com/timescale/timescaledb-docker/tree/main/docker-entrypoint-initdb.d

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create users and database
    CREATE 
        USER $PANOPTES_POSTGRES_APP_USER 
        WITH PASSWORD '$PANOPTES_POSTGRES_APP_PASSWORD';
    CREATE 
        USER $PANOPTES_POSTGRES_MIGRATOR_USER 
        WITH PASSWORD '$PANOPTES_POSTGRES_MIGRATOR_PASSWORD';

    -- Grant all privileges, and set default privileges to grant on all future tables
    GRANT 
        ALL PRIVILEGES 
        ON ALL TABLES IN SCHEMA $PANOPTES_POSTGRES_SCHEMA
        TO $PANOPTES_POSTGRES_MIGRATOR_USER;
    GRANT 
        ALL PRIVILEGES 
        ON SCHEMA $PANOPTES_POSTGRES_SCHEMA
        TO $PANOPTES_POSTGRES_MIGRATOR_USER;

    ALTER DEFAULT PRIVILEGES 
        FOR USER $PANOPTES_POSTGRES_MIGRATOR_USER 
        IN SCHEMA $PANOPTES_POSTGRES_SCHEMA 
        GRANT ALL 
            ON TABLES 
            TO $PANOPTES_POSTGRES_MIGRATOR_USER;

    ALTER DEFAULT PRIVILEGES 
        FOR USER $PANOPTES_POSTGRES_MIGRATOR_USER 
        IN SCHEMA $PANOPTES_POSTGRES_SCHEMA 
        GRANT 
            SELECT, UPDATE, INSERT, DELETE 
            ON TABLES 
            TO $PANOPTES_POSTGRES_APP_USER;
EOSQL