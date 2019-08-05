.PHONY: image

scraper_namespace = morph-test-scrapers
scraper_name = test-ruby

all: run

# This runs the scraper on kubernetes
run: image copy-code
	# If namespace already exists, the following command errors, but continues
	-kubectl create namespace clay-$(scraper_namespace)
	sed "s/{{ SCRAPER_NAMESPACE }}/$(scraper_namespace)/g; s/{{ SCRAPER_NAME }}/$(scraper_name)/g" kubernetes/job-template.yaml > kubernetes/job.yaml
	kubectl apply -f kubernetes/job.yaml
	# Wait for the pod to be up and running and then stream the logs
	kubectl wait --for condition=Ready -l job-name=$(scraper_name) --namespace=clay-$(scraper_namespace) pods
	kubectl logs -f -l job-name=$(scraper_name) --namespace=clay-$(scraper_namespace)
	# Clean up manually
	kubectl delete -f kubernetes/job.yaml
	rm kubernetes/job.yaml

# This checks out code from a scraper on github and plops it into the local blob storage
copy-code:
	rm -rf app
	git clone --depth 1 https://github.com/$(scraper_namespace)/$(scraper_name).git app
	./clay.sh copy app $(scraper_namespace) $(scraper_name)
	rm -rf app

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t morph-ng /bin/bash

image:
	docker build -t morph-ng image

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
