#!/bin/bash

# exit when any command fails
set -e
set -o pipefail
# Enable globbing for hidden files
shopt -s dotglob

if [ $# == 0 ]; then
  echo "Compiles and runs a scraper"
  echo "Usage: $0 run_name run_output"
  echo "e.g. $0 -d test/scrapers/test-python"
  exit 1
fi

RUN_NAME=$1
RUN_OUTPUT=$2

CLAY_SERVER_URL=clay-server.clay-system:8080

# Turns on debugging output in herokuish
# export TRACE=true

# TODO: Probably don't want to do this as root

started() {
  data=$(jq -c -n --arg log "$line" --arg stage "$3" '{stage: $stage, type: "started"}')
  curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/events" -d "$data"
}

finished() {
  data=$(jq -c -n --arg log "$line" --arg stage "$3" '{stage: $stage, type: "finished"}')
  curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/events" -d "$data"
}

get() {
  curl -s -H "Authorization: Bearer $2" "$CLAY_SERVER_URL/runs/$1/$3"
}

put() {
  curl -s -X PUT -H "Authorization: Bearer $2" --data-binary @- --no-buffer "$CLAY_SERVER_URL/runs/$1/$3"
}

send-logs() {
  # Send each line of stdin as a separate POST
  while IFS= read -r text ;
  do
    # Send as json
    data=$(jq -c -n --arg text "$text" --arg stage "$3" --arg stream "$4" '{stage: $stage, type: "log", stream: $stream, text: $text}')
    curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/events" -d "$data"
  done
}

send-event() {
  curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/events" -d "$3"
}

started "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" build

cd /tmp || exit

mkdir -p app cache

get "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" app | tar xzf - -C app

cp /usr/local/lib/Procfile /tmp/app/Procfile

(get "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" cache | tar xzf - -C cache 2> /dev/null) || true

# This fairly hideous construction pipes stdout and stderr to seperate commands
{ /bin/usage.sh /tmp/usage_build.json /bin/herokuish buildpack build 2>&3 | send-logs "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" build stdout; } 3>&1 1>&2 | send-logs "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" build stderr

# TODO: If the build fails then it shouldn't try to run the scraper but it should record stats

cd cache
tar -zcf - * | put "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" cache
cd ..

finished "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" build

# TODO: Factor out common code from the build and run
started "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" run
{ /bin/usage.sh /tmp/usage_run.json /bin/herokuish procfile start scraper 2>&3 | send-logs "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" run stdout; } 3>&1 1>&2 | send-logs "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" run stderr

exit_code=${PIPESTATUS[0]}

build_statistics=$(cat /tmp/usage_build.json)
run_statistics=$(cat /tmp/usage_run.json)
overall_stats="{\"exit_code\": $exit_code, \"usage\": {\"build\": $build_statistics, \"run\": $run_statistics}}"
echo "$overall_stats" | put "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" exit-data

# Now take the filename given in $RUN_OUTPUT and save that away
cd /app || exit
# TODO: Do nothing if the output file doesn't exist
put "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" output < "$RUN_OUTPUT"

finished "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" run
# TODO: Make sure that this is always sent even, for instance, if the build fails
send-event "$RUN_NAME" "$CLAY_INTERNAL_RUN_TOKEN" "EOF"
