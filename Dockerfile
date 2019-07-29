FROM herokuish:dev
MAINTAINER Matthew Landauer <matthew@oaf.org.au>

RUN apt-get update && apt-get install -y libsqlite3-dev

ADD run.sh /bin
RUN chmod +x /bin/run.sh
