#!/bin/bash

# exit when any command fails
set -e
set -o pipefail

if [ $# == 0 ]; then
  echo "Downloads a scraper from Github, compiles it and runs it"
  echo "Usage: $0 run_name run_output"
  echo "e.g. $0 morph-test-scrapers-test-ruby"
  exit 1
fi

# TODO: Allow this script to be quit with control C

RUN_NAME=$1
RUN_OUTPUT=$2

# This environment variable is used by clay.sh
export CLAY_SERVER_URL=clay-server.clay-system:8080

# Turns on debugging output in herokuish
# export TRACE=true

extract_value() {
  local filename="$1"
  local label="$2"
  local line=$(grep "$label" "$filename")
  echo "${line#*$label: }"
}

stats() {
  local filename="$1"
  local exit_code="$2"

  echo "Exit code: $exit_code"

  local wall_time_formatted=$(extract_value "$filename" "Elapsed (wall clock) time (h:mm:ss or m:ss)")
  local user_time=$(extract_value "$filename" "User time (seconds)")
  local system_time=$(extract_value "$filename" "System time (seconds)")
  local max_rss=$(extract_value "$filename" "Maximum resident set size (kbytes)")
  local page_size=$(extract_value "$filename" "Page size (bytes)")

  local part1=$(echo $wall_time_formatted | cut -d ':' -f 1)
  local part2=$(echo $wall_time_formatted | cut -d ':' -f 2)
  local part3=$(echo $wall_time_formatted | cut -d ':' -f 3)

  # If part3 is empty (time is in m:ss)
  if [ -z "$part3" ]; then
    local wall_time=$(echo "$part1 * 60.0 + $part2" | bc)
  # Else time is in h:mm:ss
  else
    local wall_time=$(echo "($part1 * 60.0 + $part2) * 60.0 + $part3" | bc)
  fi
  echo "wall_time (in seconds): $wall_time"

  local cpu_time=$(echo "$user_time + $system_time" | bc)
  echo "cpu_time (in seconds): $cpu_time"

  # There's a bug in GNU time 1.7 which wrongly reports the maximum resident
  # set size on the version of Ubuntu that we're using.
  # See https://groups.google.com/forum/#!topic/gnu.utils.help/u1MOsHL4bhg
  # Let's fix it up
  local max_rss=$(echo "$max_rss * 1024 / $page_size" | bc)
  echo "max_rss (in kbytes): $max_rss"
}

# TODO: Probably don't want to do this as root

cd /tmp || exit

/bin/clay.sh get "$RUN_NAME" "$CLAY_RUN_TOKEN" app | tar xzf -

cp /usr/local/lib/Procfile /tmp/app/Procfile

(/bin/clay.sh get "$RUN_NAME" "$CLAY_RUN_TOKEN" cache | tar xzf - 2> /dev/null) || true

# TODO: Collect separate stats (from the scraper run) for build process
# This fairly hideous construction pipes stdout and stderr to seperate commands
{ /bin/herokuish buildpack build 2>&3 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stdout; } 3>&1 1>&2 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stderr

tar -zcf - cache | /bin/clay.sh put "$RUN_NAME" "$CLAY_RUN_TOKEN" cache

# TODO: Send return code and stats about run to clay server
{ /usr/bin/time -v -o /tmp/time_output_run.txt /bin/herokuish procfile start scraper 2>&3 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stdout; } 3>&1 1>&2 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stderr

# TODO: Collect network in/out as well

exit_code=${PIPESTATUS[0]}

stats /tmp/time_output_run.txt "$exit_code"

# Now take the filename given in $RUN_OUTPUT and save that away
cd /app || exit
/bin/clay.sh put "$RUN_NAME" "$CLAY_RUN_TOKEN" output < "$RUN_OUTPUT"
