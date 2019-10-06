#!/bin/bash

# Run a scraper using clay
#
# Dependencies:
# jq - https://stedolan.github.io/jq/

# exit when any command fails
set -e

usage() {
  echo "Runs a scraper using clay"
  echo "Usage: $0 [-d] scraper_name"
  echo ""
  echo "If -d is used interpret scraper_name as a path to a directory."
  echo "Otherwise interpret it as a name of a scraper stored on GitHub."
  echo "e.g. $0 -d test/scrapers/test-python"
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

scraper_name=$1

# This environment variable is used by clay.sh
CLAY_SERVER_URL=http://localhost:8080
export CLAY_SERVER_URL

rm -rf app
if [ "$scraper_name_is_directory" = true ]; then
  cp -R "$scraper_name" app
  # Prepend scraper name with something different so that it can be cached separately
  scraper_name="local/$scraper_name"
else
  # Checkout the code from github
  git clone --quiet --depth 1 "https://github.com/$scraper_name.git" app
  rm -rf app/.git app/.gitignore
  # Add the sqlite database
  (cat "assets/client-storage/db/$scraper_name.sqlite" > app/data.sqlite 2> /dev/null) || true
  scraper_name="github/$scraper_name"
fi

create_result=$(./build/package/clay-scraper/clay.sh create "$scraper_name")
run_name=$(echo "$create_result" | jq -r ".run_name")
run_token=$(echo "$create_result" | jq -r ".run_token")

# We want the tar to have the scraper files at its root
# Note that this doesn't include hidden files currently. Do we want this?
dir=$(pwd)
cd app
tar -zcf - * | "$dir/build/package/clay-scraper/clay.sh" put "$run_name" "$run_token" app
cd "$dir"
rm -rf app

(cat "assets/client-storage/cache/$scraper_name.tgz" 2> /dev/null | ./build/package/clay-scraper/clay.sh put "$run_name" "$run_token" cache) || true
./build/package/clay-scraper/clay.sh start "$run_name" "$run_token" data.sqlite SCRAPER_NAME "$scraper_name"

if [ "$run_token" = "" ]; then
  echo "There was an error starting the scraper"
  exit 1
fi

./build/package/clay-scraper/clay.sh events "$run_name" "$run_token" | jq -r 'select(has("log")) | .log'
# Get the sqlite database from clay and save it away
mkdir -p $(dirname "assets/client-storage/db/$scraper_name")
./build/package/clay-scraper/clay.sh get "$run_name" "$run_token" output > "assets/client-storage/db/$scraper_name.sqlite"
mkdir -p $(dirname "assets/client-storage/cache/$scraper_name")
./build/package/clay-scraper/clay.sh get "$run_name" "$run_token" cache > "assets/client-storage/cache/$scraper_name.tgz"
echo "exit data returned by clay:"
./build/package/clay-scraper/clay.sh get "$run_name" "$run_token" exit-data |  jq .
./build/package/clay-scraper/clay.sh delete "$run_name" "$run_token"
