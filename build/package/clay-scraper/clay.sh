#!/bin/bash

set -eo pipefail

if [ $# == 0 ]; then
  echo "$0 - Prototype of what the clay API methods could do"
  echo "USAGE:"
  echo "  $0 COMMAND"
  echo ""
  echo "COMMANDS:"
  echo "  create NAME_PREFIX                                              Returns run name and run token"
  echo "  put RUN_NAME RUN_TOKEN [app|cache|output|exit-data]             Take stdin and upload"
  # TODO: Add support for more than one environment variable
  echo "  start RUN_NAME RUN_TOKEN SCRAPER_OUTPUT [ENV_NAME ENV_VALUE]    Start the scraper"
  echo "  events RUN_NAME RUN_TOKEN                                       Stream the events json"
  echo "  get RUN_NAME RUN_TOKEN [app|cache|output|exit-data]             Retrieve and send to stdout"
  echo "  delete RUN_NAME RUN_TOKEN                                       Cleanup after everything has finished"
  echo ""
  echo "COMMANDS (only used from container):"
  echo "  send-logs RUN_NAME RUN_TOKEN STAGE STREAM                       Take stdin and send them as logs"
  echo "  started RUN_NAME RUN_TOKEN STAGE                                Let the world know that a stage is starting"
  echo "  finished RUN_NAME RUN_TOKEN STAGE                               Let the world know that a stage is finished"
  echo "  send-event RUN_NAME RUN_TOKEN JSON                              Send arbitrary json string as an event"
  echo ""
  echo "NAME_PREFIX is chosen by the user. It doesn't have to be unique and is only"
  echo "used as a base to generate the unique run name."
  echo "STAGE can be either build or run"
  echo ""
  echo "e.g. $0 create my-first-scraper"
  exit 1
fi

if [ -z "$CLAY_SERVER_URL" ]; then
  echo "Need to set environment variable CLAY_SERVER_URL" >&2
  exit 1
fi

put() {
  curl -s -X PUT -H "Authorization: Bearer $2" --data-binary @- --no-buffer "$CLAY_SERVER_URL/runs/$1/$3"
}

get() {
  curl -s -H "Authorization: Bearer $2" "$CLAY_SERVER_URL/runs/$1/$3"
}

create() {
  curl -s -G -X POST "$CLAY_SERVER_URL/runs" -d "name_prefix=$1"
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

send-logs() {
  # Send each line of stdin as a separate POST
  # TODO: Chunk up lines that get sent close together into one request
  while IFS= read -r text ;
  do
    # Send as json
    data=$(jq -c -n --arg text "$text" --arg stage "$3" --arg stream "$4" '{stage: $stage, type: "log", stream: $stream, text: $text}')
    curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/events" -d "$data"
    # Also for the time being
    # TODO: No need for this anymore
    echo "$line"
  done
}

started() {
  data=$(jq -c -n --arg log "$line" --arg stage "$3" '{stage: $stage, type: "started"}')
  curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/events" -d "$data"
}

finished() {
  data=$(jq -c -n --arg log "$line" --arg stage "$3" '{stage: $stage, type: "finished"}')
  curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/events" -d "$data"
}

send-event() {
  curl -s -X POST -H "Authorization: Bearer $2" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$1/events" -d "$3"
}

if [ "$1" = "put" ]; then
  put $2 $3 $4
elif [ "$1" = "get" ]; then
  get $2 $3 $4
elif [ "$1" = "create" ]; then
  create $2
elif [ "$1" = "start" ]; then
  start $2 $3 $4 $5 $6
elif [ "$1" = "events" ]; then
  events $2 $3
elif [ "$1" = "delete" ]; then
  delete $2 $3
elif [ "$1" = "send-logs" ]; then
  send-logs $2 $3 $4 $5
elif [ "$1" == "started" ]; then
  started $2 $3 $4
elif [ "$1" == "finished" ]; then
  finished $2 $3 $4
elif [ "$1" == "send-event" ]; then
  send-event $2 $3 $4
else
  echo "Unknown command" >&2
  exit 1
fi
