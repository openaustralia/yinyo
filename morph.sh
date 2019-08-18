#!/bin/bash

# Run a morph scraper using clay
#
# Dependencies:
# jq - https://stedolan.github.io/jq/
# mc - https://min.io/download

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

mc config host add minio $(minikube service --url minio-service) admin changeme

# This environment variable is used by clay.sh
export CLAY_SERVER_URL=$(minikube service --url clay-server)

# TODO: Use /tmp for the app
rm -rf app
# Checkout the code from github
git clone --depth 1 "https://github.com/$morph_scraper_name.git" app
rm -rf app/.git app/.gitignore
# Add the sqlite database
(mc cat "$morph_bucket/db/$morph_scraper_name.sqlite" > app/data.sqlite) || true

create_result=$(./images/clay-scraper/clay.sh create "$morph_scraper_name")
run_name=$(echo "$create_result" | jq -r ".run_name")
run_token=$(echo "$create_result" | jq -r ".run_token")
tar -zcf - app | ./images/clay-scraper/clay.sh put "$run_name" "$run_token" app
(mc cat "$morph_bucket/cache/$morph_scraper_name.tgz" | ./images/clay-scraper/clay.sh put "$run_name" "$run_token" cache) || true
./images/clay-scraper/clay.sh start "$run_name" "$run_token" data.sqlite

if [ "$run_token" = "" ]; then
  echo "There was an error starting the scraper"
  exit 1
fi

rm -rf app
./images/clay-scraper/clay.sh logs "$run_name" "$run_token"
# Get the sqlite database from clay and save it away in a morph bucket
./images/clay-scraper/clay.sh get "$run_name" "$run_token" output | mc pipe "$morph_bucket/db/$morph_scraper_name.sqlite"
./images/clay-scraper/clay.sh get "$run_name" "$run_token" cache | mc pipe "$morph_bucket/cache/$morph_scraper_name.tgz"
./images/clay-scraper/clay.sh delete "$run_name" "$run_token"
