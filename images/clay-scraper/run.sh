#!/bin/bash

# exit when any command fails
set -e
set -o pipefail

if [ $# == 0 ]; then
  echo "Compiles and runs a scraper"
  echo "Usage: $0 run_name run_output"
  echo "e.g. $0 -d scrapers/test-python"
  exit 1
fi

# TODO: Allow this script to be quit with control C

RUN_NAME=$1
RUN_OUTPUT=$2

# This environment variable is used by clay.sh
export CLAY_SERVER_URL=clay-server.clay-system:8080

# Turns on debugging output in herokuish
# export TRACE=true

# TODO: Probably don't want to do this as root

cd /tmp || exit

mkdir app
/bin/clay.sh get "$RUN_NAME" "$CLAY_RUN_TOKEN" app | tar xzf - -C app

cp /usr/local/lib/Procfile /tmp/app/Procfile

(/bin/clay.sh get "$RUN_NAME" "$CLAY_RUN_TOKEN" cache | tar xzf - 2> /dev/null) || true

# This fairly hideous construction pipes stdout and stderr to seperate commands
{ /bin/usage.sh /tmp/usage_build.json /bin/herokuish buildpack build 2>&3 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stdout; } 3>&1 1>&2 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stderr

# TODO: If the build fails then it shouldn't try to run the scraper but it should record stats

tar -zcf - cache | /bin/clay.sh put "$RUN_NAME" "$CLAY_RUN_TOKEN" cache

{ /bin/usage.sh /tmp/usage_run.json /bin/herokuish procfile start scraper 2>&3 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stdout; } 3>&1 1>&2 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stderr

exit_code=${PIPESTATUS[0]}

build_statistics=$(cat /tmp/usage_build.json)
run_statistics=$(cat /tmp/usage_run.json)
overall_stats="{\"exit_code\": $exit_code, \"usage\": {\"build\": $build_statistics, \"run\": $run_statistics}}"
echo "$overall_stats" | /bin/clay.sh put "$RUN_NAME" "$CLAY_RUN_TOKEN" exit-data

# Now take the filename given in $RUN_OUTPUT and save that away
cd /app || exit
# TODO: Do nothing if the output file doesn't exist
/bin/clay.sh put "$RUN_NAME" "$CLAY_RUN_TOKEN" output < "$RUN_OUTPUT"
