#!/bin/sh

# Testmodus: Wenn DEV auf "false" gesetzt ist, starte /run.sh
if [ "$DEV" = "false" ]; then
    echo "Starting test mode"
    exec /run.sh
fi

echo "Starting development mode"

# Erkennung der Distribution anhand vorhandener Dateien
if [ -f /etc/alpine-release ]; then
    echo "Detected Alpine Linux"
    exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
elif [ -f /etc/debian_version ]; then
    echo "Detected Debian/Ubuntu"
    exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
else
    echo "ERROR: Unsupported base image"
    exit 1
fi
