#!/bin/bash

# Turns on debugging output in herokuish
# export TRACE=true

# Checkout latest revision of the source code of scraper into import directory
# TODO: Probably don't want to do this as root
git clone --depth 1 https://github.com/morph-test-scrapers/test-ruby.git /tmp/app

# This is where we would recognise the code as being ruby and add the Procfile.
# Alternatively we could add a standard Procfile that runs a script that recognises
# the language and runs the correct command

# For the time being just assume it's Ruby
cp /usr/local/lib/Procfile-ruby /tmp/app/Procfile

/bin/herokuish buildpack build
/bin/herokuish procfile start scraper
