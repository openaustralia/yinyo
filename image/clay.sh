#!/bin/bash

set -eo pipefail

BUCKET_CLAY="minio/clay"

if [ $# == 0 ]; then
    echo "$0 - Prototype of what the clay API methods could do"
    echo "USAGE:"
    echo "  $0 COMMAND"
    echo ""
    echo "COMMANDS (public):"
    echo "  app put DIRECTORY SCRAPER_NAME            Copy code and any data to the scraper"
    echo "  start SCRAPER_NAME SCRAPER_OUTPUT         Start the scraper"
    echo "  logs SCRAPER_NAME                         Stream the logs"
    echo "  output get SCRAPER_NAME                   Get the file output for the scraper and send to stdout"
    echo "  cleanup SCRAPER_NAME                      Cleanup after everything has finished"
    echo ""
    echo "COMMANDS (private - used by containers):"
    echo "  app get SCRAPER_NAME DIRECTORY            Get the code and data for the scraper"
    echo "  cache get SCRAPER_NAME DIRECTORY          Retrieve the build cache"
    echo "  cache put DIRECTORY SCRAPER_NAME          Save away the build cache"
    echo "  output put SCRAPER_NAME                   Take stdin and save it away"
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

COMMAND=$1

if [ "$COMMAND" = "app" ]; then
    SUBCOMMAND=$2
    if [ "$SUBCOMMAND" = "put" ]; then
      DIRECTORY=$3
      SCRAPER_NAME=$4

      # TODO: Check that $DIRECTORY exists
      tar -zcf - "$DIRECTORY" | storage put "$SCRAPER_NAME" app tgz
    elif [ "$SUBCOMMAND" = "get" ]; then
      # Get the source code of scraper into import directory. This needs to have
      # already been copied to the appropriate place in the blob store.
      # We do this because we don't want to assume that the code comes from Github.

      # TODO: Make get and put work so that the directory in each case is the same
      SCRAPER_NAME=$3
      DIRECTORY=$4

      cd "$DIRECTORY" || exit
      storage get "$SCRAPER_NAME" app tgz | tar xzf -
    else
      echo "Unknown subcommand: $SUBCOMMAND"
      exit 1
    fi
elif [ "$COMMAND" = "cache" ]; then
    # TODO: Extract common code out of app and cache command
    SUBCOMMAND=$2
    if [ "$SUBCOMMAND" = "put" ]; then
      # This is where we save away the result of the build cache for future compiles
      # For the time being we're just writing directly to the blob store (which has
      # no authentication setup) but in future we'll do it by using an authentication
      # token (available via an environment variable) which is only valid for the
      # period of this scraper run and it can only be used for updating things
      # during this scraper run. To make this work it will probably be necessary to
      # create an API service which authenticates our request and proxies the request
      # to the blob store.

      DIRECTORY=$3
      SCRAPER_NAME=$4

      # TODO: Check that $DIRECTORY exists
      tar -zcf - "$DIRECTORY" | storage put "$SCRAPER_NAME" cache tgz
    elif [ "$SUBCOMMAND" = "get" ]; then
      # TODO: Handle situation where the cache doesn't yet exist
      # TODO: Make get and put work so that the directory in each case is the same
      SCRAPER_NAME=$3
      DIRECTORY=$4

      cd "$DIRECTORY" || exit
      (storage get "$SCRAPER_NAME" cache tgz | tar xzf -) || true
    else
      echo "Unknown subcommand: $SUBCOMMAND"
      exit 1
    fi
elif [ "$COMMAND" = "output" ]; then
    SUBCOMMAND=$2
    SCRAPER_NAME=$3

    if [ "$SUBCOMMAND" = "put" ]; then
      storage put "$SCRAPER_NAME" output
    elif [ "$SUBCOMMAND" = "get" ]; then
      storage get "$SCRAPER_NAME" output
    else
      echo "Unknown subcommand: $SUBCOMMAND"
      exit 1
    fi
elif [ "$COMMAND" = "start" ]; then
    SCRAPER_NAME=$2
    SCRAPER_OUTPUT=$3

    sed "s/{{ SCRAPER_NAME }}/$SCRAPER_NAME/g; s/{{ SCRAPER_OUTPUT }}/$SCRAPER_OUTPUT/g" kubernetes/job-template.yaml > kubernetes/job.yaml
    kubectl apply -f kubernetes/job.yaml
    rm kubernetes/job.yaml
elif [ "$COMMAND" = "logs" ]; then
    SCRAPER_NAME=$2

    # Wait for the pod to be up and running and then stream the logs
    kubectl wait --for condition=Ready -l job-name="$SCRAPER_NAME" pods
    kubectl logs -f -l job-name="$SCRAPER_NAME"
elif [ "$COMMAND" = "cleanup" ]; then
    SCRAPER_NAME=$2

    kubectl delete "jobs/$SCRAPER_NAME"
    # Also clear out code and output files
    storage delete "$SCRAPER_NAME" app tgz
    storage delete "$SCRAPER_NAME" output
else
    echo "Unknown command: $COMMAND"
    exit 1
fi
