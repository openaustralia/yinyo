.PHONY: image

morph_scraper_name = morph-test-scrapers/test-ruby
morph_bucket = minio/morph

# To use the morph scraper name as a unique id for clay we need to substitute
# all non-alphanumeric characters with "-" and add a short bit of hash of the original
# string on to the end to ensure uniqueness.
# This way we get a name that is readable and close to the original and very likely unique.
clay_scraper_name = $(shell echo $(morph_scraper_name) | sed -e "s/[^[:alpha:]]/-/g")-$(shell echo $(morph_scraper_name) | shasum | head -c5)

all: run

# This runs the scraper on kubernetes
run: image copy-code
	./image/clay.sh start $(clay_scraper_name) data.sqlite
	./image/clay.sh logs $(clay_scraper_name)
	# Get the sqlite database from clay and save it away in a morph bucket
	./image/clay.sh output get $(clay_scraper_name) | mc pipe $(morph_bucket)/$(morph_scraper_name).sqlite
	./image/clay.sh cleanup $(clay_scraper_name)

# This checks out code from a scraper on github and plops it into the local blob storage
copy-code:
	rm -rf app
	# Checkout the code from github
	git clone --depth 1 https://github.com/$(morph_scraper_name).git app
	rm -rf app/.git app/.gitignore
	# Add the sqlite database
	-mc cat $(morph_bucket)/$(morph_scraper_name).sqlite > app/data.sqlite
	# And upload it to clay
	./image/clay.sh app put app $(clay_scraper_name)
	rm -rf app

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t clay /bin/bash

image:
	docker build -t clay image

lint:
	shellcheck image/run.sh image/clay.sh

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
