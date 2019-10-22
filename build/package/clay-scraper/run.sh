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

# We allow some settings to be overridden for the purposes of testing.
# We don't allow users to change any environment variables that start with CLAY_INTERNAL_SERVER_
# so they can't change any of these

if [ -z "$CLAY_INTERNAL_SERVER_URL" ]; then
  CLAY_INTERNAL_SERVER_URL=clay-server.clay-system:8080
fi

if [ -z "$CLAY_INTERNAL_BUILD_COMMAND" ]; then
  CLAY_INTERNAL_BUILD_COMMAND="/bin/herokuish buildpack build"
fi

if [ -z "$CLAY_INTERNAL_RUN_COMMAND" ]; then
  CLAY_INTERNAL_RUN_COMMAND="/bin/herokuish procfile start scraper"
fi

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
  curl -s -H "$header_auth" "$CLAY_INTERNAL_SERVER_URL/runs/$RUN_NAME/$1"
}

put() {
  curl -s -X PUT -H "$header_auth" --data-binary @- --no-buffer "$CLAY_INTERNAL_SERVER_URL/runs/$RUN_NAME/$1"
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
  local stage="$1"
  shift
  { "$@" 2>&3 | send-logs "$stage" stdout; } 3>&1 1>&2 | send-logs "$stage" stderr
}

send-event() {
  curl -s -X POST -H "$header_auth" -H "$header_ct" "$CLAY_INTERNAL_SERVER_URL/runs/$RUN_NAME/events" -d "$1"
}

extract_value() {
  local filename
  local label
  local line
  filename="$1"
  label="$2"
  line=$(grep "$label" "$filename")
  echo "${line#*$label: }"
}

wall_time() {
  local filename="$1"

  local wall_time_formatted
  wall_time_formatted=$(extract_value "$filename" "Elapsed (wall clock) time (h:mm:ss or m:ss)")

  local part1
  local part2
  local part3
  part1=$(echo "$wall_time_formatted" | cut -d ':' -f 1)
  part2=$(echo "$wall_time_formatted" | cut -d ':' -f 2)
  part3=$(echo "$wall_time_formatted" | cut -d ':' -f 3)

  local wall_time
  # If part3 is empty (time is in m:ss)
  if [ -z "$part3" ]; then
    wall_time=$(echo "$part1 * 60.0 + $part2" | bc)
  # Else time is in h:mm:ss
  else
    wall_time=$(echo "($part1 * 60.0 + $part2) * 60.0 + $part3" | bc)
  fi
  echo "$wall_time"
}

cpu_time() {
  local filename="$1"

  local user_time
  local system_time
  local cpu_time
  user_time=$(extract_value "$filename" "User time (seconds)")
  system_time=$(extract_value "$filename" "System time (seconds)")
  cpu_time=$(echo "$user_time + $system_time" | bc)
  echo "$cpu_time"
}

max_rss() {
  local filename="$1"

  local max_rss
  local page_size
  max_rss=$(extract_value "$filename" "Maximum resident set size (kbytes)")
  page_size=$(extract_value "$filename" "Page size (bytes)")

  # There's a bug in GNU time 1.7 which wrongly reports the maximum resident
  # set size on the version of Ubuntu that we're using.
  # See https://groups.google.com/forum/#!topic/gnu.utils.help/u1MOsHL4bhg
  # Let's fix it up
  max_rss=$(echo "$max_rss * 1024 / $page_size" | bc)
  echo "$max_rss"
}

usage() {
  local filename=$1
  shift

  # Doing this temporary hack to allow this script to be tested under OS X
  # TODO: Remove this temporary hack when we can
  if [ "$OSTYPE" == "darwin19" ]; then
    "$@"
    echo "{}" > "$filename"
  else
    local snapshot_before
    snapshot_before=$(ip -s -j link show eth0)

    /usr/bin/time -v -o /tmp/time_output.txt "$@"

    local snapshot_after
    snapshot_after=$(ip -s -j link show eth0)

    local rx_bytes_before
    local rx_bytes_after
    local tx_bytes_before
    local tx_bytes_after
    local rx_bytes
    local tx_bytes
    rx_bytes_before=$(echo "$snapshot_before" | jq ".[0].stats64.rx.bytes")
    tx_bytes_before=$(echo "$snapshot_before" | jq ".[0].stats64.tx.bytes")
    rx_bytes_after=$(echo "$snapshot_after" | jq ".[0].stats64.rx.bytes")
    tx_bytes_after=$(echo "$snapshot_after" | jq ".[0].stats64.tx.bytes")
    rx_bytes=$(echo "$rx_bytes_after - $rx_bytes_before" | bc)
    tx_bytes=$(echo "$tx_bytes_after - $tx_bytes_before" | bc)

    local wall_time
    local max_rss
    local cpu_time
    wall_time=$(wall_time /tmp/time_output.txt)
    max_rss=$(max_rss /tmp/time_output.txt)
    cpu_time=$(cpu_time /tmp/time_output.txt)

    # Returns result as JSON
    echo "{\"wall_time\": $wall_time, \"cpu_time\": $cpu_time, \"max_rss\": $max_rss, \"network_in\": $rx_bytes, \"network_out\": $tx_bytes}" > "$filename"
    rm /tmp/time_output.txt
  fi
}

started build

# Do initial setup. Go to our working directory and
# setup the app and cache directory
cd /tmp || exit
mkdir -p app cache

# Fill app directory
get app | tar xzf - -C app
echo "scraper: /bin/start.sh" > /tmp/app/Procfile

# Fill cache directory
(get cache | tar xzf - -C cache 2> /dev/null) || true

# Do the build
send-logs-all build usage /tmp/usage_build.json $CLAY_INTERNAL_BUILD_COMMAND

cd cache
tar -zcf - * | put cache
cd ..

finished build

# Do the actual run
started run
send-logs-all run usage /tmp/usage_run.json $CLAY_INTERNAL_RUN_COMMAND

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
