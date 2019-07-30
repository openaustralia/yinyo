#!/bin/bash

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

/bin/herokuish buildpack build
/bin/herokuish procfile start scraper
