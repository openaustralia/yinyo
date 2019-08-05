#!/bin/bash

BUCKET_CLAY="minio/clay"

if [ $# == 0 ]; then
    echo "$0 - Prototype of what the clay API methods could do"
    echo "USAGE:"
    echo "  $0 COMMAND"
    echo ""
    echo "COMMANDS:"
    echo "  copy DIRECTORY SCRAPER_NAMESPACE SCRAPER_NAME   Copy code and any data to the scraper"
    echo "  start SCRAPER_NAMESPACE SCRAPER_NAME            Start the scraper"
    echo "  logs SCRAPER_NAMESPACE SCRAPER_NAME             Stream the logs"
    echo ""
    echo "e.g. $0 copy app morph-test-scrapers/test-ruby"
    exit 1
fi

COMMAND=$1

if [ "$COMMAND" = "copy" ]; then
    DIRECTORY=$2
    SCRAPER_NAMESPACE=$3
    SCRAPER_NAME=$4

    # TODO: Check that $DIRECTORY exists
    tar --exclude .git -zcf - "$DIRECTORY" | mc pipe "$BUCKET_CLAY/$SCRAPER_NAMESPACE/$SCRAPER_NAME/app.tgz"
elif [ "$COMMAND" = "start" ]; then
    SCRAPER_NAMESPACE=$2
    SCRAPER_NAME=$3

    # If namespace already exists, the following command errors, but continues
    kubectl create namespace "clay-$SCRAPER_NAMESPACE" || true
    sed "s/{{ SCRAPER_NAMESPACE }}/$SCRAPER_NAMESPACE/g; s/{{ SCRAPER_NAME }}/$SCRAPER_NAME/g" kubernetes/job-template.yaml > kubernetes/job.yaml
    kubectl apply -f kubernetes/job.yaml
elif [ "$COMMAND" = "logs" ]; then
    SCRAPER_NAMESPACE=$2
    SCRAPER_NAME=$3

    # Wait for the pod to be up and running and then stream the logs
    kubectl wait --for condition=Ready -l job-name="$SCRAPER_NAME" --namespace="clay-$SCRAPER_NAMESPACE" pods
    kubectl logs -f -l job-name="$SCRAPER_NAME" --namespace="clay-$SCRAPER_NAMESPACE"
fi
