#!/bin/bash

set -eo pipefail

if [ $# == 0 ]; then
  echo "$0 - Prototype of what the clay API methods could do"
  echo "USAGE:"
  echo "  $0 COMMAND"
  echo ""
  echo "COMMANDS:"
  echo "  create SCRAPER_NAME                                             Returns run name and run token"
  echo "  put RUN_NAME RUN_TOKEN [app|cache|output|exit-data]             Take stdin and upload"
  # TODO: Add support for more than one environment variable
  echo "  start RUN_NAME RUN_TOKEN SCRAPER_OUTPUT [ENV_NAME ENV_VALUE]    Start the scraper"
  echo "  logs RUN_NAME RUN_TOKEN                                         Stream the logs"
  echo "  get RUN_NAME RUN_TOKEN [app|cache|output|exit-data]             Retrieve and send to stdout"
  echo "  delete RUN_NAME RUN_TOKEN                                       Cleanup after everything has finished"
  echo ""
  echo "COMMANDS (only used from container):"
  echo "  send-logs RUN_NAME RUN_TOKEN STAGE STREAM                        Take stdin and send them as logs"
  echo ""
  echo "SCRAPER_NAME is chosen by the user. It doesn't have to be unique and is only"
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
  curl -s -G -X POST "$CLAY_SERVER_URL/runs" -d "scraper_name=$2"
elif [ "$1" = "start" ]; then
  # Send as json
  data=$(jq -c -n --arg output "$4" --arg env_name "$5" --arg env_value "$6" '{output: $output, env: [{name: $env_name, value: $env_value}]}')
  curl -s -X POST -H "Authorization: Bearer $3" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$2/start" -d "$data"
elif [ "$1" = "logs" ]; then
  curl -s --no-buffer -H "Authorization: Bearer $3" "$CLAY_SERVER_URL/runs/$2/logs"
elif [ "$1" = "delete" ]; then
  curl -s -X DELETE -H "Authorization: Bearer $3" "$CLAY_SERVER_URL/runs/$2"
elif [ "$1" = "send-logs" ]; then
  # Send each line of stdin as a separate POST
  # TODO: Chunk up lines that get sent close together into one request
  while IFS= read -r line ;
  do
    # Send as json
    data=$(jq -c -n --arg log "$line" --arg stage "$4" --arg stream "$5" '{stage: $stage, stream: $stream, log: $log}')
    curl -s -X POST -H "Authorization: Bearer $3" -H "Content-Type: application/json" "$CLAY_SERVER_URL/runs/$2/logs" -d "$data"
    # Also for the time being
    echo "$line"
  done
else
  echo "Unknown command" >&2
  exit 1
fi
