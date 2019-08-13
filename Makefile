.PHONY: image server

morph_scraper_name = morph-test-scrapers/test-python

all: run

run: image
	./morph.sh $(morph_scraper_name)

# If you want an interactive shell in the container
shell: image
	docker run --rm -i -t clay-scraper /bin/bash

image:
	docker build -t clay-scraper image

# TODO: Figure out how to get this to run just before a scraper run every time
# The main problem is figuring out how to wait for the deployment to finish
# After this run you can access the clay server at http://localhost:8080
server:
	# TODO: Use multi-stage docker build for go app
	# TODO: Make minimal docker image
	# TODO: Use https://skaffold.dev/ for development workflow
	cd server; GOOS=linux go build -o ./app .
	docker build -t clay-server server
	kubectl replace -f kubernetes/clay-server.yaml --force

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
