.PHONY: image

morph_scraper_name = "morph-test-scrapers/test-ruby"

# To use the morph scraper name as a unique id for clay we need to substitute
# all non-alphanumeric characters with "-" and add a short bit of hash of the original
# string on to the end to ensure uniqueness.
# This way we get a name that is readable and close to the original and very likely unique.
clay_scraper_name = $(shell echo $(morph_scraper_name) | sed -e "s/[^[:alpha:]]/-/g")-$(shell echo $(morph_scraper_name) | shasum | head -c5)

all: run

# This runs the scraper on kubernetes
run: image copy-code
	echo $(sha)
	./clay.sh start $(clay_scraper_name)
	./clay.sh logs $(clay_scraper_name)
	./clay.sh cleanup $(clay_scraper_name)

# This checks out code from a scraper on github and plops it into the local blob storage
copy-code:
	rm -rf app
	git clone --depth 1 https://github.com/$(morph_scraper_name).git app
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
