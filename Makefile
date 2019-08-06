.PHONY: image

scraper_namespace = morph-test-scrapers
scraper_name = test-ruby

# TODO: Add a few characters from an md5 of the github path to ensure name is unique
clay_scraper_name = "$(scraper_namespace)-$(scraper_name)"

all: run

# This runs the scraper on kubernetes
run: image copy-code
	./clay.sh start $(clay_scraper_name)
	./clay.sh logs $(clay_scraper_name)
	./clay.sh cleanup $(clay_scraper_name)

# This checks out code from a scraper on github and plops it into the local blob storage
copy-code:
	rm -rf app
	git clone --depth 1 https://github.com/$(scraper_namespace)/$(scraper_name).git app
	./clay.sh copy app $(clay_scraper_name)
	rm -rf app

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t clay /bin/bash

image:
	docker build -t clay image

lint:
	shellcheck image/run.sh

shellcheck:
	# This assumes OS X for the time being
	brew install shellcheck

minio:
	kubectl apply -f kubernetes/minio-deployment.yaml

# Populate local bucket with copy of some of what's in the Heroku bucket
heroku-buildpack-ruby:
	mc mb -p minio/heroku-buildpack-ruby
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/heroku-18/ruby-2.5.1.tgz | mc pipe minio/heroku-buildpack-ruby/heroku-18/ruby-2.5.1.tgz
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/heroku-18/ruby-2.5.5.tgz | mc pipe minio/heroku-buildpack-ruby/heroku-18/ruby-2.5.5.tgz
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/bundler/bundler-1.15.2.tgz | mc pipe minio/heroku-buildpack-ruby/bundler/bundler-1.15.2.tgz
