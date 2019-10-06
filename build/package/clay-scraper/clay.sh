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

if [ "$1" = "put" ]; then
  curl -s -X PUT -H "Authorization: Bearer $3" --data-binary @- --no-buffer "$CLAY_SERVER_URL/runs/$2/$4"
elif [ "$1" = "get" ]; then
  curl -s -H "Authorization: Bearer $3" "$CLAY_SERVER_URL/runs/$2/$4"
elif [ "$1" = "create" ]; then
  curl -s -G -X POST "$CLAY_SERVER_URL/runs" -d "name_prefix=$2"
elif [ "$1" = "start" ]; then
  # Send as json
  data=$(jq -c -n --arg output "$4" --arg env_name "$5" --arg env_value "$6" '{output: $output, env: [{name: $env_name, value: $env_value}]}')
  curl -s -X POST -H "Authorization: Bearer $3" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$2/start" -d "$data"
elif [ "$1" = "events" ]; then
  curl -s --no-buffer -H "Authorization: Bearer $3" "$CLAY_SERVER_URL/runs/$2/events"
elif [ "$1" = "delete" ]; then
  curl -s -X DELETE -H "Authorization: Bearer $3" "$CLAY_SERVER_URL/runs/$2"
elif [ "$1" = "send-logs" ]; then
  # Send each line of stdin as a separate POST
  # TODO: Chunk up lines that get sent close together into one request
  while IFS= read -r line ;
  do
    # Send as json
    data=$(jq -c -n --arg log "$line" --arg stage "$4" --arg stream "$5" '{stage: $stage, type: "log", stream: $stream, log: $log}')
    curl -s -X POST -H "Authorization: Bearer $3" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$2/events" -d "$data"
    # Also for the time being
    echo "$line"
  done
elif [ "$1" == "started" ]; then
  data=$(jq -c -n --arg log "$line" --arg stage "$4" '{stage: $stage, type: "started"}')
  curl -s -X POST -H "Authorization: Bearer $3" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$2/events" -d "$data"
elif [ "$1" == "finished" ]; then
  data=$(jq -c -n --arg log "$line" --arg stage "$4" '{stage: $stage, type: "finished"}')
  curl -s -X POST -H "Authorization: Bearer $3" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$2/events" -d "$data"
elif [ "$1" == "send-event" ]; then
  curl -s -X POST -H "Authorization: Bearer $3" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$2/events" -d "$4"
else
  echo "Unknown command" >&2
  exit 1
fi
