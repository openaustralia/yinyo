#!/bin/bash

# exit when any command fails
set -e
set -o pipefail

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

/bin/clay.sh get "$SCRAPER_NAME" "$CLAY_RUN_TOKEN" app | tar xzf -

cp /usr/local/lib/Procfile /tmp/app/Procfile

(/bin/clay.sh get "$SCRAPER_NAME" "$CLAY_RUN_TOKEN" cache | tar xzf -) || true

/bin/herokuish buildpack build

tar -zcf - cache | /bin/clay.sh put "$SCRAPER_NAME" "$CLAY_RUN_TOKEN" cache

/bin/herokuish procfile start scraper

# Now take the filename given in $SCRAPER_OUTPUT and save that away
cd /app || exit
/bin/clay.sh put "$SCRAPER_NAME" "$CLAY_RUN_TOKEN" output < "$SCRAPER_OUTPUT"
