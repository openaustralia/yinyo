#!/bin/bash

# Give admin access to the local blob store. Only doing this for ease
# of development.
# TODO: REMOVE THIS AS SOON AS POSSIBLE
mc config host add minio http://minio-service:9000 admin changeme

if [ $# == 0 ]; then
    echo "Downloads a scraper from Github, compiles it and runs it"
    echo "Usage: $0 scraper_name scraper_output"
    echo "e.g. $0 morph-test-scrapers-test-ruby"
    exit 1
fi

# TODO: Allow this script to be quit with control C

SCRAPER_NAME=$1
SCRAPER_OUTPUT=$2

# Turns on debugging output in herokuish
# export TRACE=true

# TODO: Probably don't want to do this as root

cd /tmp || exit

/bin/clay.sh app get "$SCRAPER_NAME" /tmp

# This is where we would recognise the code as being ruby and add the Procfile.
# Alternatively we could add a standard Procfile that runs a script that recognises
# the language and runs the correct command

# For the time being just assume it's Ruby
cp /usr/local/lib/Procfile-ruby /tmp/app/Procfile

# Use local minio for getting buildpack binaries
export BUILDPACK_VENDOR_URL=http://minio-service:9000/heroku-buildpack-ruby

/bin/clay.sh cache get "$SCRAPER_NAME" /tmp

/bin/herokuish buildpack build

/bin/clay.sh cache put cache "$SCRAPER_NAME"

/bin/herokuish procfile start scraper

# Now take the filename given in $SCRAPER_OUTPUT and save that away
cd /app || exit
/bin/clay.sh output put "$SCRAPER_NAME" "${SCRAPER_OUTPUT##*.}" < "$SCRAPER_OUTPUT"
