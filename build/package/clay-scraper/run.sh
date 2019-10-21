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
  data=$(jq -c -n --arg stage "$1" '{stage: $stage, type: "started"}')
  curl -s -X POST -H "Authorization: Bearer $CLAY_INTERNAL_RUN_TOKEN" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$RUN_NAME/events" -d "$data"
}

finished() {
  data=$(jq -c -n --arg stage "$1" '{stage: $stage, type: "finished"}')
  curl -s -X POST -H "Authorization: Bearer $CLAY_INTERNAL_RUN_TOKEN" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$RUN_NAME/events" -d "$data"
}

get() {
  curl -s -H "Authorization: Bearer $CLAY_INTERNAL_RUN_TOKEN" "$CLAY_SERVER_URL/runs/$RUN_NAME/$1"
}

put() {
  curl -s -X PUT -H "Authorization: Bearer $CLAY_INTERNAL_RUN_TOKEN" --data-binary @- --no-buffer "$CLAY_SERVER_URL/runs/$RUN_NAME/$1"
}

send-logs() {
  # Send each line of stdin as a separate POST
  while IFS= read -r text ;
  do
    # Send as json
    data=$(jq -c -n --arg text "$text" --arg stage "$1" --arg stream "$2" '{stage: $stage, type: "log", stream: $stream, text: $text}')
    curl -s -X POST -H "Authorization: Bearer $CLAY_INTERNAL_RUN_TOKEN" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$RUN_NAME/events" -d "$data"
  done
}

send-event() {
  curl -s -X POST -H "Authorization: Bearer $CLAY_INTERNAL_RUN_TOKEN" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$RUN_NAME/events" -d "$1"
}

started build

cd /tmp || exit

mkdir -p app cache

get app | tar xzf - -C app

cp /usr/local/lib/Procfile /tmp/app/Procfile

(get cache | tar xzf - -C cache 2> /dev/null) || true

# This fairly hideous construction pipes stdout and stderr to seperate commands
{ /bin/usage.sh /tmp/usage_build.json /bin/herokuish buildpack build 2>&3 | send-logs build stdout; } 3>&1 1>&2 | send-logs build stderr

cd cache
tar -zcf - * | put cache
cd ..

finished build

# TODO: Factor out common code from the build and run
started run
{ /bin/usage.sh /tmp/usage_run.json /bin/herokuish procfile start scraper 2>&3 | send-logs run stdout; } 3>&1 1>&2 | send-logs run stderr

exit_code=${PIPESTATUS[0]}

build_statistics=$(cat /tmp/usage_build.json)
run_statistics=$(cat /tmp/usage_run.json)
overall_stats="{\"exit_code\": $exit_code, \"usage\": {\"build\": $build_statistics, \"run\": $run_statistics}}"
echo "$overall_stats" | put exit-data

# Now take the filename given in $RUN_OUTPUT and save that away
cd /app || exit
put output < "$RUN_OUTPUT"

finished run
send-event "EOF"
