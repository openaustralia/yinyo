#!/bin/bash

# Run a morph scraper using clay

# exit when any command fails
set -e

if [ $# == 0 ]; then
  echo "Runs a morph scraper using clay"
  echo "Usage: $0 scraper_name"
  echo "e.g. $0 morph-test-scrapers/test-ruby"
  exit 1
fi

morph_scraper_name=$1
morph_bucket="minio/morph"

# To use the morph scraper name as a unique id for clay we need to substitute
# all non-alphanumeric characters with "-" and add a short bit of hash of the original
# string on to the end to ensure uniqueness.
# This way we get a name that is readable and close to the original and very likely unique.
clay_scraper_name=$(echo "$morph_scraper_name" | sed -e "s/[^[:alpha:]]/-/g")-$(echo "$morph_scraper_name" | shasum | head -c5)

# TODO: Use /tmp for the app
rm -rf app
# Checkout the code from github
git clone --depth 1 "https://github.com/$morph_scraper_name.git" app
rm -rf app/.git app/.gitignore
# Add the sqlite database
(mc cat "$morph_bucket/$morph_scraper_name.sqlite" > app/data.sqlite) || true

run_token=$(./images/clay-scraper/clay.sh create "$clay_scraper_name")
./images/clay-scraper/clay.sh run app "$clay_scraper_name" "$run_token" data.sqlite

if [ "$run_token" = "" ]; then
  echo "There was an error starting the scraper"
  exit 1
fi

rm -rf app
./images/clay-scraper/clay.sh logs "$clay_scraper_name" "$run_token"
# Get the sqlite database from clay and save it away in a morph bucket
./images/clay-scraper/clay.sh output get "$clay_scraper_name" "$run_token" | mc pipe "$morph_bucket/$morph_scraper_name.sqlite"
./images/clay-scraper/clay.sh cleanup "$clay_scraper_name" "$run_token"
