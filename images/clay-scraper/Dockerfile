FROM herokuish:dev
MAINTAINER Matthew Landauer <matthew@oaf.org.au>

RUN apt-get update && apt-get install -y libsqlite3-dev

# Add prerun script which will disable output buffering for ruby
ADD prerun.rb /usr/local/lib/prerun.rb

# Add standard Procfiles
ADD Procfile /usr/local/lib

# Add perl buildpack
RUN /bin/herokuish buildpack install https://github.com/miyagawa/heroku-buildpack-perl.git 2da7480a8339f01968ce3979655555a0ade20564

# Temporarily install kubectl
# TODO: Remove
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN apt-get install -y kubectl

ADD run.sh start.sh clay.sh /bin/
RUN chmod +x /bin/run.sh /bin/start.sh /bin/clay.sh
