#!/bin/bash

# Just for the time being assume we're running a ruby scraper
# TODO: Check what scraper language we're running and kick off in the appropriate way
bundle exec ruby -r/usr/local/lib/prerun.rb scraper.rb
