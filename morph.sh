#!/bin/bash

# Run a morph scraper using clay
#
# Dependencies:
# jq - https://stedolan.github.io/jq/
# mc - https://min.io/download

# exit when any command fails
set -e

usage() {
  echo "Runs a morph scraper using clay"
  echo "Usage: $0 [-d] scraper_name"
  echo ""
  echo "If -d is used interpret scraper_name as a path to a directory."
  echo "Otherwise interpret it as a name of a morph scraper stored on GitHub."
  echo "e.g. $0 morph-test-scrapers/test-ruby"
  exit 1
}

if [ $# == 0 ]; then
  usage
fi

scraper_name_is_directory=false

while getopts 'd' c
do
  case $c in
    d) scraper_name_is_directory=true ;;
    *) usage ;;
  esac
done

# Get rid of the parsed options so far
shift $((OPTIND-1))

morph_scraper_name=$1
morph_bucket="morph/morph"

store_access_key=$(grep store_access_key secrets-morph.env | cut -d "=" -f 2)
store_secret_key=$(grep store_secret_key secrets-morph.env | cut -d "=" -f 2)

# This command doesn't work with "-api s3v4". No idea why.
mc config host add morph "$(minikube service --url minio-service -n clay-system)" "$store_access_key" "$store_secret_key" -api s3v4

# This environment variable is used by clay.sh
CLAY_SERVER_URL=$(minikube service --url clay-server -n clay-system)
export CLAY_SERVER_URL

if [ "$scraper_name_is_directory" = true ]; then
  cp -R "$morph_scraper_name" app
  # Prepend scraper name with something different so that it can be cached separately
  morph_scraper_name="local/$morph_scraper_name"
else
  rm -rf app
  # Checkout the code from github
  git clone --quiet --depth 1 "https://github.com/$morph_scraper_name.git" app
  rm -rf app/.git app/.gitignore
  # Add the sqlite database
  (mc cat "$morph_bucket/db/$morph_scraper_name.sqlite" > app/data.sqlite 2> /dev/null) || true
fi

create_result=$(./images/clay-scraper/clay.sh create "$morph_scraper_name")
run_name=$(echo "$create_result" | jq -r ".run_name")
run_token=$(echo "$create_result" | jq -r ".run_token")

# TODO: Don't want to depend on tar directory being called app
tar -zcf - app | ./images/clay-scraper/clay.sh put "$run_name" "$run_token" app
rm -rf app

(mc cat "$morph_bucket/cache/$morph_scraper_name.tgz" 2> /dev/null | ./images/clay-scraper/clay.sh put "$run_name" "$run_token" cache) || true
./images/clay-scraper/clay.sh start "$run_name" "$run_token" data.sqlite MORPH_SCRAPER_NAME "$morph_scraper_name"

if [ "$run_token" = "" ]; then
  echo "There was an error starting the scraper"
  exit 1
fi

./images/clay-scraper/clay.sh logs "$run_name" "$run_token"
# Get the sqlite database from clay and save it away in a morph bucket
./images/clay-scraper/clay.sh get "$run_name" "$run_token" output | mc pipe "$morph_bucket/db/$morph_scraper_name.sqlite"
./images/clay-scraper/clay.sh get "$run_name" "$run_token" cache | mc pipe "$morph_bucket/cache/$morph_scraper_name.tgz"
echo "exit data returned by clay:"
./images/clay-scraper/clay.sh get "$run_name" "$run_token" exit-data |  jq .
./images/clay-scraper/clay.sh delete "$run_name" "$run_token"
