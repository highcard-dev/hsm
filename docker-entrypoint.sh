#!/bin/sh
set -e

SESSION_FILE="${HOME}/.config/hsm/session.json"

# Auto-login if:
# - session.json doesn't exist AND
# - user is not explicitly running the login command
if [ ! -f "$SESSION_FILE" ] && [ "$1" != "login" ]; then
    hsm login
fi

hsm "$@"        