#!/bin/bash

if [ -f "scraper.rb" ]; then
  bundle exec ruby -r/usr/local/lib/prerun.rb scraper.rb
elif [ -f "scraper.php" ]; then
  php -d include_path=.:/app/vendor/openaustralia/scraperwiki scraper.php
elif [ -f "scraper.py" ]; then
  # -u turns off buffering for stdout and stderr
  python -u scraper.py
elif [ -f "scraper.pl" ]; then
  perl -Mlib=/app/local/lib/perl5 scraper.pl
elif [ -f "scraper.js" ]; then
  node --expose-gc scraper.js
else
  # TODO: Make a better error message
  echo "Can'f find scraper to run"
  exit 1
fi
