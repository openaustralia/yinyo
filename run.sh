#!/bin/sh

/bin/herokuish buildpack build
/bin/herokuish procfile start scraper
