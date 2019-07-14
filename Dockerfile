FROM gliderlabs/herokuish:v0.5.0
MAINTAINER Matthew Landauer <matthew@oaf.org.au>

RUN apt-get update && apt-get install -y libsqlite3-dev
