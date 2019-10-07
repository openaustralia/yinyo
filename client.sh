#!/bin/bash

# Run a scraper using clay
#
# Dependencies:
# jq - https://stedolan.github.io/jq/

# exit when any command fails
set -e

usage() {
  echo "Runs a scraper in a local directory using clay"
  echo "Usage: $0 scraper_directory output_file"
  echo ""
  echo "e.g. $0 test/scrapers/test-python data.sqlite"
  echo "The output is written to the same local directory at the end. The output file path"
  echo "is given relative to the scraper directory"
  exit 1
}

if [ $# == 0 ]; then
  usage
fi

# Get rid of the parsed options so far
shift $((OPTIND-1))

scraper_directory=$1
output=$2

# This environment variable is used by clay.sh
CLAY_SERVER_URL=http://localhost:8080
export CLAY_SERVER_URL

create_result=$(./build/package/clay-scraper/clay.sh create "$scraper_directory")
run_name=$(echo "$create_result" | jq -r ".run_name")
run_token=$(echo "$create_result" | jq -r ".run_token")

# We want the tar to have the scraper files at its root
# Note that this doesn't include hidden files currently. Do we want this?
dir=$(pwd)
cd "$scraper_directory"
tar -zcf - * | "$dir/build/package/clay-scraper/clay.sh" put "$run_name" "$run_token" app
cd "$dir"

(cat "assets/client-storage/cache/$scraper_directory.tgz" 2> /dev/null | ./build/package/clay-scraper/clay.sh put "$run_name" "$run_token" cache) || true
./build/package/clay-scraper/clay.sh start "$run_name" "$run_token" "$output" SCRAPER_NAME "$scraper_directory"

if [ "$run_token" = "" ]; then
  echo "There was an error starting the scraper"
  exit 1
fi

./build/package/clay-scraper/clay.sh events "$run_name" "$run_token" | jq -r 'select(has("log")) | .log'
# Get the sqlite database from clay and save it away
./build/package/clay-scraper/clay.sh get "$run_name" "$run_token" output > "$scraper_directory/$output"
mkdir -p $(dirname "assets/client-storage/cache/$scraper_directory")
./build/package/clay-scraper/clay.sh get "$run_name" "$run_token" cache > "assets/client-storage/cache/$scraper_directory.tgz"
echo "exit data returned by clay:"
./build/package/clay-scraper/clay.sh get "$run_name" "$run_token" exit-data |  jq .
./build/package/clay-scraper/clay.sh delete "$run_name" "$run_token"
