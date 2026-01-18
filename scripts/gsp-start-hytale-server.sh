#!/bin/bash

set -e
set -a

echo "Starting hytale server..."

#enforce HSM_URL is set
if [ -z "$HSM_URL" ]; then
    echo "HSM_URL is not set"
    exit 1
fi

echo "Fetching game session from $HSM_URL..."

#download environment variables for game session and apply them
CURL_ARGS=(-sSf -L -o .env -X POST)
if [ -n "$JWT_TOKEN" ]; then
    echo "Using JWT token for authentication"
    CURL_ARGS+=(-H "Authorization: Bearer $JWT_TOKEN")
fi

curl "${CURL_ARGS[@]}" "$HSM_URL/game-session"

echo "Loaded game session environment"

source .env

echo "Starting Java server..."

java -jar Server/HytaleServer.jar --assets Assets.zip
