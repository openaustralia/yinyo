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
  echo "SCRAPER_NAME is chosen by the user. It doesn't have to be unique and is only"
  echo "used as a base to generate the unique run name."
  echo ""
  echo "e.g. $0 copy app morph-test-scrapers-test-ruby"
  exit 1
fi

clay-host () {
  local host
  if [ -z "$KUBERNETES_SERVICE_HOST" ]; then
    echo "localhost:8080"
  else
    echo "clay-server:8080"
  fi
}

if [ "$1" = "put" ]; then
  curl -s -X PUT -H "Clay-Run-Token: $3" --data-binary @- --no-buffer "$(clay-host)/runs/$2/$4"
elif [ "$1" = "get" ]; then
  curl -s -H "Clay-Run-Token: $3" "$(clay-host)/runs/$2/$4"
elif [ "$1" = "create" ]; then
  curl -s -G -X POST "$(clay-host)/runs" -d "scraper_name=$2"
elif [ "$1" = "start" ]; then
  curl -s -G -X POST -H "Clay-Run-Token: $3" "$(clay-host)/runs/$2/start" -d "output=$4"
elif [ "$1" = "logs" ]; then
  curl -s --no-buffer -H "Clay-Run-Token: $3" "$(clay-host)/runs/$2/logs"
elif [ "$1" = "delete" ]; then
  curl -s -X DELETE -H "Clay-Run-Token: $3" "$(clay-host)/runs/$2"
else
  echo "Unknown command"
  exit 1
fi
