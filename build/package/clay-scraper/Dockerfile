FROM mlandauer/herokuish:for-morph-ng

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD prerun.rb /usr/local/lib/prerun.rb

# Add perl buildpack
RUN /bin/herokuish buildpack install https://github.com/miyagawa/heroku-buildpack-perl.git 2da7480a8339f01968ce3979655555a0ade20564

RUN apt-get install -y jq time bc net-tools

ADD run.sh start.sh /bin/
RUN chmod +x /bin/run.sh /bin/start.sh
