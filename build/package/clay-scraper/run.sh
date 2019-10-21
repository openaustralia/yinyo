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

# This is the header used for authorisation for some of the API calls
header_auth="Authorization: Bearer $CLAY_INTERNAL_RUN_TOKEN"
header_ct="Content-Type: application/json"

started() {
  data=$(jq -c -n --arg stage "$1" '{stage: $stage, type: "started"}')
  send-event "$data"
}

finished() {
  data=$(jq -c -n --arg stage "$1" '{stage: $stage, type: "finished"}')
  send-event "$data"
}

get() {
  curl -s -H "$header_auth" "$CLAY_SERVER_URL/runs/$RUN_NAME/$1"
}

put() {
  curl -s -X PUT -H "$header_auth" --data-binary @- --no-buffer "$CLAY_SERVER_URL/runs/$RUN_NAME/$1"
}

send-logs() {
  # Send each line of stdin as a separate POST
  while IFS= read -r text ;
  do
    # Send as json
    data=$(jq -c -n --arg text "$text" --arg stage "$1" --arg stream "$2" '{stage: $stage, type: "log", stream: $stream, text: $text}')
    send-event "$data"
  done
}

send-logs-all() {
  # This fairly hideous construction pipes stdout and stderr to seperate commands
  { $1 2>&3 | send-logs build stdout; } 3>&1 1>&2 | send-logs build stderr
}

send-event() {
  curl -s -X POST -H "$header_auth" -H "$header_ct" "$CLAY_SERVER_URL/runs/$RUN_NAME/events" -d "$1"
}

started build

cd /tmp || exit

mkdir -p app cache

get app | tar xzf - -C app

cp /usr/local/lib/Procfile /tmp/app/Procfile

(get cache | tar xzf - -C cache 2> /dev/null) || true

send-logs-all "/bin/usage.sh /tmp/usage_build.json /bin/herokuish buildpack build"

cd cache
tar -zcf - * | put cache
cd ..

finished build

started run
send-logs-all "/bin/usage.sh /tmp/usage_run.json /bin/herokuish procfile start scraper"

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
