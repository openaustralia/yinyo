#!/bin/bash

BUCKET_CLAY="minio/clay"

if [ $# == 0 ]; then
    echo "$0 - Prototype of what the clay API methods could do"
    echo "USAGE:"
    echo "  $0 COMMAND"
    echo ""
    echo "COMMANDS:"
    echo "  copy DIRECTORY SCRAPER_NAME   Copy code and any data to the scraper"
    echo ""
    echo "e.g. $0 copy app morph-test-scrapers/test-ruby"
    exit 1
fi

COMMAND=$1

if [ "$COMMAND" = "copy" ]; then
    DIRECTORY=$2
    SCRAPER_NAME=$3

    # TODO: Check that $DIRECTORY exists
    tar --exclude .git -zcf - "$DIRECTORY" | mc pipe "$BUCKET_CLAY/$SCRAPER_NAME/app.tgz"
fi
