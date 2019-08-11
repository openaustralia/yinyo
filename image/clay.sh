#!/bin/bash

set -eo pipefail

BUCKET_CLAY="minio/clay"

if [ $# == 0 ]; then
  echo "$0 - Prototype of what the clay API methods could do"
  echo "USAGE:"
  echo "  $0 COMMAND"
  echo ""
  echo "COMMANDS (public):"
  echo "  run DIRECTORY SCRAPER_NAME SCRAPER_OUTPUT    Upload code and data and run the scraper. Returns token"
  echo "  logs SCRAPER_NAME RUN_TOKEN                  Stream the logs"
  echo "  output get SCRAPER_NAME                      Get the file output for the scraper and send to stdout"
  echo "  cleanup SCRAPER_NAME                         Cleanup after everything has finished"
  echo ""
  echo "COMMANDS (private - used by containers):"
  echo "  app get SCRAPER_NAME DIRECTORY               Get the code and data for the scraper"
  echo "  cache get SCRAPER_NAME DIRECTORY             Retrieve the build cache"
  echo "  cache put DIRECTORY SCRAPER_NAME             Save away the build cache"
  echo "  output put SCRAPER_NAME                      Take stdin and save it away"
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
  local directory=$2

  cd "$directory" || exit
  storage get "$scraper_name" app tgz | tar xzf -
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
  local directory=$1
  local scraper_name=$2

  # TODO: Check that $directory exists
  tar -zcf - "$directory" | storage put "$scraper_name" cache tgz
}

# TODO: Make get and put work so that the directory in each case is the same
command-cache-get () {
  local scraper_name=$1
  local directory=$2

  cd "$directory" || exit
  (storage get "$scraper_name" cache tgz | tar xzf -) || true
}

command-output-put () {
  local scraper_name=$1

  storage put "$scraper_name" output
}

command-output-get () {
  local scraper_name=$1

  storage get "$scraper_name" output
}

command-run () {
  local directory=$1
  local scraper_name=$2
  local scraper_output=$3

  # TODO: Check that $directory exists
  tar -zcf - "$directory" | storage put "$scraper_name" app tgz

  local run_token
  run_token=$(openssl rand -hex 32)
  sed "s/{{ SCRAPER_NAME }}/$scraper_name/g; s/{{ RUN_TOKEN }}/$run_token/g" kubernetes/secret-template.yaml > kubernetes/secret.yaml
  kubectl apply -f kubernetes/secret.yaml > /dev/null
  rm kubernetes/secret.yaml
  sed "s/{{ SCRAPER_NAME }}/$scraper_name/g; s/{{ SCRAPER_OUTPUT }}/$scraper_output/g" kubernetes/job-template.yaml > kubernetes/job.yaml
  kubectl apply -f kubernetes/job.yaml > /dev/null
  rm kubernetes/job.yaml
  echo "$run_token"
}

command-logs () {
  local scraper_name=$1
  local run_token=$2

  # Check that run_token is valid
  local actual_run_token
  actual_run_token=$(kubectl get "secret/$scraper_name" -o=jsonpath="{.data.run_token}" | base64 -D)

  if [ "$run_token" != "$actual_run_token" ]; then
    echo "Invalid run token"
    exit 1
  fi

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

  kubectl delete "jobs/$scraper_name"
  kubectl delete "secrets/$scraper_name"
  # Also clear out code and output files
  storage delete "$scraper_name" app tgz
  storage delete "$scraper_name" output
}

if [ "$1" = "app" ] && [ "$2" = "get" ]; then
  command-app-get "$3" "$4"
elif [ "$1" = "cache" ] && [ "$2" = "put" ]; then
  command-cache-put "$3" "$4"
elif [ "$1" = "cache" ] && [ "$2" = "get" ]; then
  command-cache-get "$3" "$4"
elif [ "$1" = "output" ] && [ "$2" = "put" ]; then
  command-output-put "$3"
elif [ "$1" = "output" ] && [ "$2" = "get" ]; then
  command-output-get "$3"
elif [ "$1" = "run" ]; then
  command-run "$2" "$3" "$4"
elif [ "$1" = "logs" ]; then
  command-logs "$2" "$3"
elif [ "$1" = "cleanup" ]; then
  command-cleanup "$2"
else
  echo "Unknown command"
  exit 1
fi
