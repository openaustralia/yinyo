FROM herokuish:dev
MAINTAINER Matthew Landauer <matthew@oaf.org.au>

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD prerun.rb /usr/local/lib/prerun.rb

ADD run.sh /bin
RUN chmod +x /bin/run.sh
