#!/bin/bash

BUCKET_CLAY="minio/clay"

if [ $# == 0 ]; then
    echo "$0 - Prototype of what the clay API methods could do"
    echo "USAGE:"
    echo "  $0 COMMAND"
    echo ""
    echo "COMMANDS (public):"
    echo "  app put DIRECTORY SCRAPER_NAME    Copy code and any data to the scraper"
    echo "  start SCRAPER_NAME                Start the scraper"
    echo "  logs SCRAPER_NAME                 Stream the logs"
    echo "  cleanup SCRAPER_NAME              Cleanup after everything has finished"
    echo ""
    echo "COMMANDS (private - used by containers):"
    echo "  app get SCRAPER_NAME DIRECTORY    Get the code and data for the scraper"
    echo "  cache put DIRECTORY SCRAPER_NAME  Save away the build cache"
    echo "  cache get SCRAPER_NAME DIRECTORY  Retrieve the build cache"
    echo ""
    echo "SCRAPER_NAME is chosen by the user and must be unique and only contain"
    echo "lower case alphanumeric characters and '-' up to maximum length"
    echo "of 253 characters."
    echo ""
    echo "e.g. $0 copy app morph-test-scrapers-test-ruby"
    exit 1
fi

COMMAND=$1

if [ "$COMMAND" = "app" ]; then
    SUBCOMMAND=$2
    if [ "$SUBCOMMAND" = "put" ]; then
      DIRECTORY=$3
      SCRAPER_NAME=$4

      # TODO: Check that $DIRECTORY exists
      tar --exclude .git -zcf - "$DIRECTORY" | mc pipe "$BUCKET_CLAY/app/$SCRAPER_NAME.tgz"
    elif [ "$SUBCOMMAND" = "get" ]; then
      # TODO: Make get and put work so that the directory in each case is the same
      SCRAPER_NAME=$3
      DIRECTORY=$4

      cd $DIRECTORY || exit
      mc cat "$BUCKET_CLAY/app/$SCRAPER_NAME.tgz" | tar xzf -
    else
      echo "Unknown subcommand: $SUBCOMMAND"
      exit 1
    fi
elif [ "$COMMAND" = "cache" ]; then
    # TODO: Extract common code out of app and cache command
    SUBCOMMAND=$2
    if [ "$SUBCOMMAND" = "put" ]; then
      DIRECTORY=$3
      SCRAPER_NAME=$4

      # TODO: Check that $DIRECTORY exists
      tar --exclude .git -zcf - "$DIRECTORY" | mc pipe "$BUCKET_CLAY/cache/$SCRAPER_NAME.tgz"
    elif [ "$SUBCOMMAND" = "get" ]; then
      # TODO: Handle situation where the cache doesn't yet exist
      # TODO: Make get and put work so that the directory in each case is the same
      SCRAPER_NAME=$3
      DIRECTORY=$4

      cd $DIRECTORY || exit
      mc cat "$BUCKET_CLAY/cache/$SCRAPER_NAME.tgz" | tar xzf -
    else
      echo "Unknown subcommand: $SUBCOMMAND"
      exit 1
    fi
elif [ "$COMMAND" = "start" ]; then
    SCRAPER_NAME=$2

    sed "s/{{ SCRAPER_NAME }}/$SCRAPER_NAME/g" kubernetes/job-template.yaml > kubernetes/job.yaml
    kubectl apply -f kubernetes/job.yaml
    rm kubernetes/job.yaml
elif [ "$COMMAND" = "logs" ]; then
    SCRAPER_NAME=$2

    # Wait for the pod to be up and running and then stream the logs
    kubectl wait --for condition=Ready -l job-name="$SCRAPER_NAME" pods
    kubectl logs -f -l job-name="$SCRAPER_NAME"
elif [ "$COMMAND" = "cleanup" ]; then
    SCRAPER_NAME=$2

    kubectl delete jobs/$SCRAPER_NAME
else
    echo "Unknown command: $COMMAND"
    exit 1
fi
