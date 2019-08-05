.PHONY: image

all: run

run: image
	docker run --rm -v `pwd`/cache:/tmp/cache morph-ng /bin/run.sh morph-test-scrapers/test-ruby

# Clean the cache
clean:
	rm -rf cache

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t -v `pwd`/cache:/tmp/cache morph-ng /bin/bash

image:
	docker build -t morph-ng image

lint:
	shellcheck run.sh

shellcheck:
	# This assumes OS X for the time being
	brew install shellcheck

# This runs the scraper on kubernetes
kubernetes: image
	kubectl apply -f kubernetes/job.yaml
	# Wait for the pod to be up and running and then stream the logs
	kubectl wait --for condition=Ready -l job-name=scraper pods
	kubectl logs -f -l job-name=scraper
	# Clean up manually
	kubectl delete -f kubernetes/job.yaml

minio:
	kubectl apply -f kubernetes/minio-deployment.yaml

# Populate local bucket with copy of some of what's in the Heroku bucket
heroku-buildpack-ruby:
	mc mb -p minio/heroku-buildpack-ruby
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/heroku-18/ruby-2.5.1.tgz | mc pipe minio/heroku-buildpack-ruby/heroku-18/ruby-2.5.1.tgz
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/heroku-18/ruby-2.5.5.tgz | mc pipe minio/heroku-buildpack-ruby/heroku-18/ruby-2.5.5.tgz
	curl https://s3-external-1.amazonaws.com/heroku-buildpack-ruby/bundler/bundler-1.15.2.tgz | mc pipe minio/heroku-buildpack-ruby/bundler/bundler-1.15.2.tgz
