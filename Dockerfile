FROM herokuish:dev
MAINTAINER Matthew Landauer <matthew@oaf.org.au>

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD prerun.rb /usr/local/lib/prerun.rb

# Add standard Procfiles
ADD Procfile-ruby /usr/local/lib

# Add MinIO client
RUN wget https://dl.min.io/client/mc/release/linux-amd64/mc -O /bin/mc
RUN chmod +x /bin/mc

ADD run.sh /bin
RUN chmod +x /bin/run.sh
