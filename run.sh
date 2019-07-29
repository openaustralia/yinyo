#!/bin/sh

# Turns on debugging output in herokuish
# export TRACE=true

/bin/herokuish buildpack build
/bin/herokuish procfile start scraper
