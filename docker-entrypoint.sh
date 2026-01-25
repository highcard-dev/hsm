#!/bin/sh
set -e

# Auto-login if:
# - session.json doesn't exist AND
# - user is not explicitly running the login command
if [ ! -f "$SESSION_FILE" ] && [ "$1" != "login" ]; then
    hsm login --session-location "$SESSION_FILE" 
fi

hsm "$@" --session-location "$SESSION_FILE"