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
  local rx_bytes="$2"
  local tx_bytes="$3"

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

  local cpu_time=$(echo "$user_time + $system_time" | bc)

  # There's a bug in GNU time 1.7 which wrongly reports the maximum resident
  # set size on the version of Ubuntu that we're using.
  # See https://groups.google.com/forum/#!topic/gnu.utils.help/u1MOsHL4bhg
  # Let's fix it up
  local max_rss=$(echo "$max_rss * 1024 / $page_size" | bc)

  # Returns result as JSON
  echo "{\"wall_time\": $wall_time, \"cpu_time\": $cpu_time, \"max_rss\": $max_rss, \"network_in\": $rx_bytes, \"network_out\": $tx_bytes}"
}

# TODO: Probably don't want to do this as root

cd /tmp || exit

/bin/clay.sh get "$RUN_NAME" "$CLAY_RUN_TOKEN" app | tar xzf -

cp /usr/local/lib/Procfile /tmp/app/Procfile

(/bin/clay.sh get "$RUN_NAME" "$CLAY_RUN_TOKEN" cache | tar xzf - 2> /dev/null) || true

snapshot=$(ip -s -j link show eth0)
rx_bytes_before=$(echo "$snapshot" | jq ".[0].stats64.rx.bytes")
tx_bytes_before=$(echo "$snapshot" | jq ".[0].stats64.tx.bytes")

# This fairly hideous construction pipes stdout and stderr to seperate commands
{ /usr/bin/time -v -o /tmp/time_output_build.txt /bin/herokuish buildpack build 2>&3 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stdout; } 3>&1 1>&2 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stderr

snapshot=$(ip -s -j link show eth0)
rx_bytes_after=$(echo "$snapshot" | jq ".[0].stats64.rx.bytes")
tx_bytes_after=$(echo "$snapshot" | jq ".[0].stats64.tx.bytes")
rx_bytes_build=$(echo "$rx_bytes_after - $rx_bytes_before" | bc)
tx_bytes_build=$(echo "$tx_bytes_after - $tx_bytes_before" | bc)

# TODO: If the build fails then it shouldn't try to run the scraper but it should record stats

tar -zcf - cache | /bin/clay.sh put "$RUN_NAME" "$CLAY_RUN_TOKEN" cache

snapshot=$(ip -s -j link show eth0)
rx_bytes_before=$(echo "$snapshot" | jq ".[0].stats64.rx.bytes")
tx_bytes_before=$(echo "$snapshot" | jq ".[0].stats64.tx.bytes")

{ /usr/bin/time -v -o /tmp/time_output_run.txt /bin/herokuish procfile start scraper 2>&3 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stdout; } 3>&1 1>&2 | /bin/clay.sh send-logs "$RUN_NAME" "$CLAY_RUN_TOKEN" stderr

snapshot=$(ip -s -j link show eth0)
rx_bytes_after=$(echo "$snapshot" | jq ".[0].stats64.rx.bytes")
tx_bytes_after=$(echo "$snapshot" | jq ".[0].stats64.tx.bytes")
rx_bytes_run=$(echo "$rx_bytes_after - $rx_bytes_before" | bc)
tx_bytes_run=$(echo "$tx_bytes_after - $tx_bytes_before" | bc)

exit_code=${PIPESTATUS[0]}

build_statistics=$(stats /tmp/time_output_build.txt $rx_bytes_build $tx_bytes_build)
run_statistics=$(stats /tmp/time_output_run.txt $rx_bytes_run $tx_bytes_run)
overall_stats="{\"exit_code\": $exit_code, \"usage\": {\"build\": $build_statistics, \"run\": $run_statistics}}"
echo $overall_stats | /bin/clay.sh put "$RUN_NAME" "$CLAY_RUN_TOKEN" exit-data

# Now take the filename given in $RUN_OUTPUT and save that away
cd /app || exit
# TODO: Do nothing if the output file doesn't exist
/bin/clay.sh put "$RUN_NAME" "$CLAY_RUN_TOKEN" output < "$RUN_OUTPUT"
