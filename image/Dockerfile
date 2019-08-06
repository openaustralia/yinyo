FROM herokuish:dev
MAINTAINER Matthew Landauer <matthew@oaf.org.au>

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD prerun.rb /usr/local/lib/prerun.rb

# Add standard Procfiles
ADD Procfile /usr/local/lib

# Add MinIO client
RUN wget https://dl.min.io/client/mc/release/linux-amd64/mc -O /bin/mc
RUN chmod +x /bin/mc

ADD run.sh start.sh clay.sh /bin/
RUN chmod +x /bin/run.sh /bin/start.sh /bin/clay.sh

# Override heroku ruby buildpack with patched version that allows us
# to get all binary assets from a local S3 rather than defaulting to
# the heroku S3 bucket for some (but not all) binary assets

RUN rm -rf /tmp/buildpacks/01_buildpack-ruby
RUN herokuish buildpack install https://github.com/mlandauer/heroku-buildpack-ruby.git consistent_buildpack_vendor_url_override
