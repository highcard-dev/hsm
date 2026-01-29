#!/bin/bash

set -e

echo "Installing hytale server..."

#enforce HSM_URL is set
if [ -z "$HSM_URL" ]; then
    echo "HSM_URL is not set"
    exit 1
fi

echo "Fetching download URL from $HSM_URL..."

CURL_ARGS=(-sSf)
if [ -n "$JWT_TOKEN" ]; then
    CURL_ARGS+=(-H "Authorization: Bearer $JWT_TOKEN")
fi

DOWNLOAD_URL=$(curl "${CURL_ARGS[@]}" "$HSM_URL/download?patchline=${PATCHLINE:-release}")


if [ -z "$DOWNLOAD_URL" ]; then
    echo "Failed to get download URL"
    exit 1
fi

echo "Downloading hytale server..."
#download the latest version of the hytale server
curl -sSfL -o hytale-server.zip "$DOWNLOAD_URL"

echo "Downloaded hytale server"

echo "Unzipping hytale server..."
#unzip the hytale server
unzip -o hytale-server.zip

echo "Unzipped hytale server"

echo "Cleaning up..."
#clean up the zip file
rm hytale-server.zip

echo "Cleaned up"
