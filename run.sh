#!/bin/bash

mc config host add minio http://minio-service:9000 admin changeme

if [ $# == 0 ]; then
    echo "Downloads a scraper from Github, compiles it and runs it"
    echo "Usage: $0 scraper_name"
    echo "e.g. $0 morph-test-scrapers/test-ruby"
    exit 1
fi

# TODO: Allow this script to be quit with control C

SCRAPER_NAME=$1

# Turns on debugging output in herokuish
# export TRACE=true

# Checkout latest revision of the source code of scraper into import directory
# TODO: Probably don't want to do this as root
git clone --depth 1 "https://github.com/$SCRAPER_NAME.git" /tmp/app

# This is where we would recognise the code as being ruby and add the Procfile.
# Alternatively we could add a standard Procfile that runs a script that recognises
# the language and runs the correct command

# For the time being just assume it's Ruby
cp /usr/local/lib/Procfile-ruby /tmp/app/Procfile

# Use local minio for getting buildpack binaries
export BUILDPACK_VENDOR_URL=http://minio-service:9000/heroku-buildpack-ruby

/bin/herokuish buildpack build

# This is where we save away the result of the build cache for future compiles
# For the time being we're just writing directly to the blob store (which has
# no authentication setup) but in future we'll do it by using an authentication
# token (available via an environment variable) which is only valid for the
# period of this scraper run and it can only be used for updating things
# during this scraper run. To make this work it will probably be necessary to
# create an API service which authenticates our request and proxies the request
# to the blob store.

tar zcPf /tmp/cache.tgz /tmp/cache/
mc cp /tmp/cache.tgz "minio/morph/$SCRAPER_NAME/"

/bin/herokuish procfile start scraper
