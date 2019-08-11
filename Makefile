.PHONY: image

morph_scraper_name = morph-test-scrapers/test-ruby

all: run

run: image
	./morph.sh $(morph_scraper_name)

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t clay /bin/bash

image:
	docker build -t clay image

lint:
	shellcheck image/run.sh image/clay.sh morph.sh

shellcheck:
	# This assumes OS X for the time being
	brew install shellcheck

install: install-minio install-logging

install-minio:
	kubectl apply -f kubernetes/minio-deployment.yaml

install-logging:
	# The following can't be run multiple times
	# TODO: Make this more sensible
	kubectl create namespace logging
	# TODO: Use oss image for elasticsearch & kibana
	helm install --name elasticsearch stable/elasticsearch --namespace logging
	helm install --name kibana stable/kibana --set env.ELASTICSEARCH_HOSTS=http://elasticsearch-client:9200 --namespace logging
	kubectl apply -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-service-account.yaml
	kubectl apply -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-role.yaml
	kubectl apply -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-role-binding.yaml
	kubectl apply -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/output/elasticsearch/fluent-bit-configmap.yaml
	kubectl apply -f kubernetes/fluent-bit-ds.yaml

# Populate local bucket with copy of some of what's in the Heroku bucket
heroku-buildpack-ruby:
	mc mb -p minio/heroku-buildpack-ruby
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/heroku-18/ruby-2.5.1.tgz | mc pipe minio/heroku-buildpack-ruby/heroku-18/ruby-2.5.1.tgz
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/heroku-18/ruby-2.5.5.tgz | mc pipe minio/heroku-buildpack-ruby/heroku-18/ruby-2.5.5.tgz
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/bundler/bundler-1.15.2.tgz | mc pipe minio/heroku-buildpack-ruby/bundler/bundler-1.15.2.tgz
