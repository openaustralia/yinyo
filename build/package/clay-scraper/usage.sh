#!/bin/bash

# exit when any command fails
set -e
set -o pipefail

if [ $# == 0 ]; then
  echo "Run an external command and measure its resource usage"
  echo "Usage: $0 output_file command"
  echo "e.g. $0 usage.json sleep 2"
  exit 1
fi

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

filename=$1
shift

snapshot_before=$(ip -s -j link show eth0)

/usr/bin/time -v -o /tmp/time_output.txt "$@"

snapshot_after=$(ip -s -j link show eth0)

rx_bytes_before=$(echo "$snapshot_before" | jq ".[0].stats64.rx.bytes")
tx_bytes_before=$(echo "$snapshot_before" | jq ".[0].stats64.tx.bytes")
rx_bytes_after=$(echo "$snapshot_after" | jq ".[0].stats64.rx.bytes")
tx_bytes_after=$(echo "$snapshot_after" | jq ".[0].stats64.tx.bytes")
rx_bytes=$(echo "$rx_bytes_after - $rx_bytes_before" | bc)
tx_bytes=$(echo "$tx_bytes_after - $tx_bytes_before" | bc)

wall_time=$(wall_time /tmp/time_output.txt)
max_rss=$(max_rss /tmp/time_output.txt)
cpu_time=$(cpu_time /tmp/time_output.txt)

# Returns result as JSON
echo "{\"wall_time\": $wall_time, \"cpu_time\": $cpu_time, \"max_rss\": $max_rss, \"network_in\": $rx_bytes, \"network_out\": $tx_bytes}" > "$filename"
rm /tmp/time_output.txt
