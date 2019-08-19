.PHONY: image server

morph_scraper_name = morph-test-scrapers/test-python

all: run

run:
	./morph.sh $(morph_scraper_name)

# If you want an interactive shell in the container
shell:
	docker run --rm -i -t openaustralia/clay-scraper:v1 /bin/bash

lint:
	shellcheck image/run.sh image/clay.sh morph.sh

shellcheck:
	# This assumes OS X for the time being
	brew install shellcheck

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

buckets:
	mc config host add minio $(shell minikube service --url minio-service -n clay-system) admin changeme
	mc mb minio/clay
	mc mb minio/morph
