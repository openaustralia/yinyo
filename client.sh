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

CLAY_SERVER_URL=http://localhost:8080

create() {
  curl -s -G -X POST "$CLAY_SERVER_URL/runs" -d "name_prefix=$1"
}

get() {
  curl -s -H "Authorization: Bearer $2" "$CLAY_SERVER_URL/runs/$1/$3"
}

put() {
  curl -s -X PUT -H "Authorization: Bearer $2" --data-binary @- --no-buffer "$CLAY_SERVER_URL/runs/$1/$3"
}

start() {
  # Send as json
  data=$(jq -c -n --arg output "$3" --arg env_name "$4" --arg env_value "$5" '{output: $output, env: [{name: $env_name, value: $env_value}]}')
  curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/start" -d "$data"
}

events() {
  curl -s --no-buffer -H "Authorization: Bearer $2" "$CLAY_SERVER_URL/runs/$1/events"
}

delete() {
  curl -s -X DELETE -H "Authorization: Bearer $2" "$CLAY_SERVER_URL/runs/$1"
}

create_result=$(create "$scraper_directory")
run_name=$(echo "$create_result" | jq -r ".run_name")
run_token=$(echo "$create_result" | jq -r ".run_token")

# We want the tar to have the scraper files at its root
# Note that this doesn't include hidden files currently. Do we want this?
dir=$(pwd)
cd "$scraper_directory"
tar -zcf - * | put "$run_name" "$run_token" app
cd "$dir"

(cat "assets/client-storage/cache/$scraper_directory.tgz" 2> /dev/null | put "$run_name" "$run_token" cache) || true
start "$run_name" "$run_token" "$output" SCRAPER_NAME "$scraper_directory"

if [ "$run_token" = "" ]; then
  echo "There was an error starting the scraper"
  exit 1
fi

events "$run_name" "$run_token" | jq -r 'select(has("text")) | .text'
# Get the sqlite database from clay and save it away
get "$run_name" "$run_token" output > "$scraper_directory/$output"
mkdir -p $(dirname "assets/client-storage/cache/$scraper_directory")
get "$run_name" "$run_token" cache > "assets/client-storage/cache/$scraper_directory.tgz"
echo "exit data returned by clay:"
get "$run_name" "$run_token" exit-data |  jq .
delete "$run_name" "$run_token"
