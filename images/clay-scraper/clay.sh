#!/bin/bash

set -eo pipefail

BUCKET_CLAY="minio/clay"

if [ $# == 0 ]; then
  echo "$0 - Prototype of what the clay API methods could do"
  echo "USAGE:"
  echo "  $0 COMMAND"
  echo ""
  echo "COMMANDS (public):"
  echo "  create SCRAPER_NAME                          Returns run token"
  echo "  app put SCRAPER_NAME RUN_TOKEN               Take stdin and upload code and data"
  echo "  cache put SCRAPER_NAME RUN_TOKEN             Take stdin and upload the build cache"
  echo "  run SCRAPER_NAME RUN_TOKEN SCRAPER_OUTPUT    Run the scraper"
  echo "  logs SCRAPER_NAME RUN_TOKEN                  Stream the logs"
  echo "  output get SCRAPER_NAME RUN_TOKEN            Get the file output for the scraper and send to stdout"
  echo "  cache get SCRAPER_NAME RUN_TOKEN             Retrieve the build cache and send to stdout"
  echo "  cleanup SCRAPER_NAME RUN_TOKEN               Cleanup after everything has finished"
  echo ""
  echo "COMMANDS (private - used by containers):"
  echo "  app get SCRAPER_NAME RUN_TOKEN               Get the code and data for the scraper and send to stdout"
  echo "  output put SCRAPER_NAME RUN_TOKEN            Take stdin and save it away"
  echo ""
  echo "SCRAPER_NAME is chosen by the user and must be unique and only contain"
  echo "lower case alphanumeric characters and '-' up to maximum length"
  echo "of 253 characters."
  echo ""
  echo "e.g. $0 copy app morph-test-scrapers-test-ruby"
  exit 1
fi

storage () {
  local action=$1
  local scraper_name=$2
  # TODO: Check that file_name is one of app, cache, output
  local file_name=$3
  local file_extension=$4

  local path="$BUCKET_CLAY/$file_name/$scraper_name"
  if [ -n "$file_extension" ]; then
    path="$path.$file_extension"
  fi

  if [ "$action" = "get" ]; then
    mc cat "$path"
  elif [ "$action" = "put" ]; then
    mc pipe "$path"
  elif [ "$action" = "delete" ]; then
    mc rm "$path"
  else
    echo "Unknown action: $action"
    exit 1
  fi
}

# Get the source code of scraper into import directory. This needs to have
# already been copied to the appropriate place in the blob store.
# We do this because we don't want to assume that the code comes from Github.
# TODO: Make get and put work so that the directory in each case is the same
command-app-get () {
  local scraper_name=$1
  local run_token=$2

  check-run-token "$scraper_name" "$run_token"

  storage get "$scraper_name" app tgz
}

# This is where we save away the result of the build cache for future compiles
# For the time being we're just writing directly to the blob store (which has
# no authentication setup) but in future we'll do it by using an authentication
# token (available via an environment variable) which is only valid for the
# period of this scraper run and it can only be used for updating things
# during this scraper run. To make this work it will probably be necessary to
# create an API service which authenticates our request and proxies the request
# to the blob store.
command-cache-put () {
  local scraper_name=$1
  local run_token=$2

  # TODO: Use more conventional basic auth
  curl -X POST -H "Clay-Run-Token: $run_token" --data-binary @- --no-buffer "$(clay-host)/scrapers/$scraper_name/cache"
}

clay-host () {
  local host
  if [ -z "$KUBERNETES_SERVICE_HOST" ]; then
    echo "localhost:8080"
  else
    echo "clay-server:8080"
  fi
}

# TODO: Make get and put work so that the directory in each case is the same
command-cache-get () {
  local scraper_name=$1
  local run_token=$2

  check-run-token "$scraper_name" "$run_token"

  storage get "$scraper_name" cache tgz
}

command-output-put () {
  local scraper_name=$1
  local run_token=$2

  check-run-token "$scraper_name" "$run_token"

  storage put "$scraper_name" output
}

command-output-get () {
  local scraper_name=$1
  local run_token=$2

  check-run-token "$scraper_name" "$run_token"

  storage get "$scraper_name" output
}

command-create() {
  local scraper_name=$1

  # Use clay server running on kubernetes to do the work
  curl -X POST "$(clay-host)/scrapers/$scraper_name/create"
}

command-app-put () {
  local scraper_name=$1
  local run_token=$2

  # TODO: Use more conventional basic auth
  curl -X POST -H "Clay-Run-Token: $run_token" --data-binary @- --no-buffer "$(clay-host)/scrapers/$scraper_name/app"
}

command-run () {
  local scraper_name=$1
  local run_token=$2
  local scraper_output=$3

  # Use clay server running on kubernetes to do the work
  # TODO: Use more conventional basic auth
  # TODO: Put scraper output as a parameter in the url
  curl -X POST -H "Clay-Run-Token: $run_token" -H "Clay-Scraper-Output: $scraper_output" "$(clay-host)/scrapers/$scraper_name/run"
}

# base64 has different command line options on OS X and Linux
# So, make a little cross platform wrapper
decode-base64 () {
  if [ $(uname) = "Darwin" ]; then
    base64 -D
  else
    base64 -d
  fi
}

check-run-token () {
  local scraper_name=$1
  local run_token=$2

  local actual_run_token
  actual_run_token=$(kubectl get "secret/$scraper_name" -o=jsonpath="{.data.run_token}" | decode-base64)

  if [ "$run_token" != "$actual_run_token" ]; then
    echo "Invalid run token"
    exit 1
  fi
}

command-logs () {
  local scraper_name=$1
  local run_token=$2

  check-run-token "$scraper_name" "$run_token"

  # If $type is not empty that means the jobs has finished or completed
  type=$(kubectl get "jobs/$scraper_name" -o=jsonpath="{.status.conditions[0].type}")
  # If job is starting or running
  if [ -z "$type" ]; then
    # Then wait for the pod to be ready
    kubectl wait --for condition=Ready -l job-name="$scraper_name" pods
  fi
  # Only then start streaming the logs
  kubectl logs -f -l job-name="$scraper_name"
}

command-cleanup () {
  local scraper_name=$1
  local run_token=$2

  check-run-token "$scraper_name" "$run_token"

  kubectl delete "jobs/$scraper_name"
  kubectl delete "secrets/$scraper_name"
  # Also clear out the temporary state stored on blob store
  storage delete "$scraper_name" app tgz
  storage delete "$scraper_name" output
  storage delete "$scraper_name" cache tgz
}

if [ "$1" = "app" ] && [ "$2" = "put" ]; then
  command-app-put "$3" "$4"
elif [ "$1" = "app" ] && [ "$2" = "get" ]; then
  command-app-get "$3" "$4" "$5"
elif [ "$1" = "cache" ] && [ "$2" = "put" ]; then
  command-cache-put "$3" "$4"
elif [ "$1" = "cache" ] && [ "$2" = "get" ]; then
  command-cache-get "$3" "$4"
elif [ "$1" = "output" ] && [ "$2" = "put" ]; then
  command-output-put "$3" "$4"
elif [ "$1" = "output" ] && [ "$2" = "get" ]; then
  command-output-get "$3" "$4"
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
