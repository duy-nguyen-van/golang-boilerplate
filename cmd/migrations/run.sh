#!/bin/bash

urlEncode() {
    string="$1"
    encoded=""
    for ((i=0; i<${#string}; i++)); do
        char="${string:$i:1}"
        if [[ "$char" =~ [a-zA-Z0-9\.\_\~\-] ]]; then
            encoded+="$char"
        else
            printf -v hex '%02x' "'$char"
            encoded+="%"$hex
        fi
    done
    echo "$encoded"
}

encoded_user=$(urlEncode $POSTGRESQL_USER)
encoded_password=$(urlEncode $POSTGRESQL_PASSWORD)

if [ "$POSTGRESQL_SSL" = "true" ]; then
    SSL_MODE="require"
else
    SSL_MODE="disable"
fi

export DATABASE_URL="postgresql://$encoded_user:$encoded_password@$POSTGRESQL_HOST:$POSTGRESQL_PORT/$POSTGRESQL_DB?sslmode=$SSL_MODE&search_path=${POSTGRESQL_SCHEMA}"

echo "Starting Atlas migrations on DATABASE_URL: $DATABASE_URL"

# Step Inspect the database schema
echo "Inspecting database schema..."
if ! atlas schema inspect --url "$DATABASE_URL"; then
    echo "Schema inspection failed. Exiting..."
    exit 1
fi
echo "Schema inspection successful."

# Step Apply migrations
apply_args=(migrate apply --dir "file://migrations" --url "$DATABASE_URL")
if [ "${ATLAS_ALLOW_DIRTY}" = "true" ]; then
    apply_args+=(--allow-dirty)
fi
if [ -n "${ATLAS_BASELINE}" ]; then
    apply_args+=(--baseline "${ATLAS_BASELINE}")
fi

if ! atlas "${apply_args[@]}"; then
    echo "Migration failed. Exiting..."
    exit 1
fi
echo "Migration completed successfully."
