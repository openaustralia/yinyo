#!/bin/bash

set -eo pipefail

if [ $# == 0 ]; then
  echo "$0 - Prototype of what the clay API methods could do"
  echo "USAGE:"
  echo "  $0 COMMAND"
  echo ""
  echo "COMMANDS:"
  echo "  create SCRAPER_NAME                          Returns run name and run token"
  echo "  put RUN_NAME RUN_TOKEN [app|cache|output]    Take stdin and upload"
  echo "  start RUN_NAME RUN_TOKEN SCRAPER_OUTPUT      Start the scraper"
  echo "  logs RUN_NAME RUN_TOKEN                      Stream the logs"
  echo "  get RUN_NAME RUN_TOKEN [app|cache|output]    Retrieve and send to stdout"
  echo "  delete RUN_NAME RUN_TOKEN                    Cleanup after everything has finished"
  echo ""
  echo "COMMANDS (only used from container):"
  echo "  send-logs RUN_NAME RUN_TOKEN STREAM          Take stdin and send them as logs"
  echo ""
  echo "SCRAPER_NAME is chosen by the user. It doesn't have to be unique and is only"
  echo "used as a base to generate the unique run name."
  echo ""
  echo "e.g. $0 copy app morph-test-scrapers-test-ruby"
  exit 1
fi

if [ -z "$CLAY_SERVER_URL" ]; then
  echo "Need to set environment variable CLAY_SERVER_URL" >&2
  exit 1
fi

if [ "$1" = "put" ]; then
  curl -s -X PUT -H "Clay-Run-Token: $3" --data-binary @- --no-buffer "$CLAY_SERVER_URL/runs/$2/$4"
elif [ "$1" = "get" ]; then
  curl -s -H "Clay-Run-Token: $3" "$CLAY_SERVER_URL/runs/$2/$4"
elif [ "$1" = "create" ]; then
  curl -s -G -X POST "$CLAY_SERVER_URL/runs" -d "scraper_name=$2"
elif [ "$1" = "start" ]; then
  curl -s -G -X POST -H "Clay-Run-Token: $3" "$CLAY_SERVER_URL/runs/$2/start" -d "output=$4"
elif [ "$1" = "logs" ]; then
  curl -s --no-buffer -H "Clay-Run-Token: $3" "$CLAY_SERVER_URL/runs/$2/logs"
elif [ "$1" = "delete" ]; then
  curl -s -X DELETE -H "Clay-Run-Token: $3" "$CLAY_SERVER_URL/runs/$2"
elif [ "$1" = "send-logs" ]; then
  # Send each line of stdin as a separate POST
  # TODO: Chunk up lines that get sent close together into one request
  # TODO: Send STREAM
  while IFS= read -r line ;
  do
     echo "$line" | curl -s -X POST -H "Clay-Run-Token: $3" --data-binary @- "$CLAY_SERVER_URL/runs/$2/logs"
     # Also for the time being
     echo "$line"
  done
else
  echo "Unknown command" >&2
  exit 1
fi
