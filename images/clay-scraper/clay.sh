#!/bin/bash

set -eo pipefail

if [ $# == 0 ]; then
  echo "$0 - Prototype of what the clay API methods could do"
  echo "USAGE:"
  echo "  $0 COMMAND"
  echo ""
  echo "COMMANDS:"
  echo "  create SCRAPER_NAME                          Returns run name and run token"
  echo "  put [app|cache|output] RUN_NAME RUN_TOKEN    Take stdin and upload"
  echo "  run RUN_NAME RUN_TOKEN SCRAPER_OUTPUT        Run the scraper"
  echo "  logs RUN_NAME RUN_TOKEN                      Stream the logs"
  echo "  get [app|cache|output] RUN_NAME RUN_TOKEN    Retrieve and send to stdout"
  echo "  cleanup RUN_NAME RUN_TOKEN                   Cleanup after everything has finished"
  echo ""
  echo "SCRAPER_NAME is chosen by the user. It doesn't have to be unique and is only"
  echo "used as a base to generate the unique run name. However it must only contain"
  echo "lower case alphanumeric characters and '-' up to maximum length"
  echo "of 253 characters."
  echo ""
  echo "e.g. $0 copy app morph-test-scrapers-test-ruby"
  exit 1
fi

command-store () {
  local method=$1
  local type=$2
  local run_name=$3
  local run_token=$4

  # TODO: Use more conventional basic auth
  if [ "$method" = "get" ]; then
    curl -s -H "Clay-Run-Token: $run_token" "$(clay-host)/scrapers/$run_name/$type"
  elif [ "$method" = "put" ]; then
    curl -s -X POST -H "Clay-Run-Token: $run_token" --data-binary @- --no-buffer "$(clay-host)/scrapers/$run_name/$type"
  else
    echo "Unexpected method: $method"
    exit 1
  fi
}

clay-host () {
  local host
  if [ -z "$KUBERNETES_SERVICE_HOST" ]; then
    echo "localhost:8080"
  else
    echo "clay-server:8080"
  fi
}

command-create() {
  local scraper_name=$1

  # Use clay server running on kubernetes to do the work
  curl -s -X POST "$(clay-host)/scrapers/$scraper_name/create"
}

command-run () {
  local run_name=$1
  local run_token=$2
  local scraper_output=$3

  # Use clay server running on kubernetes to do the work
  # TODO: Use more conventional basic auth
  # TODO: Put scraper output as a parameter in the url
  curl -s -X POST -H "Clay-Run-Token: $run_token" -H "Clay-Scraper-Output: $scraper_output" "$(clay-host)/scrapers/$run_name/run"
}

command-logs () {
  local run_name=$1
  local run_token=$2

  curl -s --no-buffer -H "Clay-Run-Token: $run_token" "$(clay-host)/scrapers/$run_name/logs"
}

command-cleanup () {
  local run_name=$1
  local run_token=$2

  curl -s -X POST -H "Clay-Run-Token: $run_token" "$(clay-host)/scrapers/$run_name/cleanup"
}

if [ "$1" = "put" ]; then
  command-store "$1" "$2" "$3" "$4"
elif [ "$1" = "get" ]; then
  command-store "$1" "$2" "$3" "$4"
elif [ "$1" = "create" ]; then
  command-create "$2"
elif [ "$1" = "run" ]; then
  command-run "$2" "$3" "$4" "$5"
elif [ "$1" = "logs" ]; then
  command-logs "$2" "$3"
elif [ "$1" = "cleanup" ]; then
  command-cleanup "$2" "$3"
else
  echo "Unknown command"
  exit 1
fi
