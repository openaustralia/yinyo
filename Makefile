.PHONY: image server test build ppa run website apidocs minikube buckets clean skaffold dashboard mocks

all: run

run: install
	yinyo test/scrapers/test-python --output data.sqlite

test:
	go test -short -cover ./...

integration:
	go test -cover ./...

install:
	go install cmd/yinyo/yinyo.go

ppa:
	sudo add-apt-repository ppa:longsleep/golang-backports
	sudo apt-get update
	sudo apt-get install golang-go

mocks:
	mockery -all -keeptree

website: apidocs
	cd site; npm install
	# Starts a development web server at http://localhost:1313
	hugo server -s site -D

publish-website:
	cd site; /bin/sh ./publish_to_ghpages.sh

apidocs:
	widdershins --summary openapi/definition.yaml -o openapi/definition.md
	shins openapi/definition.md --layout $(shell pwd)/site/layout.ejs -o site/content/api.html --inline --logo site/static/logo.svg --logo-url / --css site/api-overrides.css

minikube:
	minikube start --memory=3072 --disk-size='30gb'
	# We're using helm to install kubedb because that works with Kubernetes > 1.15
	helm repo add appscode https://charts.appscode.com/stable/
	helm repo update
	helm install kubedb-operator appscode/kubedb --version v0.13.0-rc.0 --namespace kube-system
	helm install kubedb-catalog appscode/kubedb-catalog --version v0.13.0-rc.0 --namespace kube-system

dashboard:
	minikube dashboard

skaffold:
	skaffold dev --port-forward=true --status-check=false

minio_access_key = $(shell grep access_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_secret_key = $(shell grep secret_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_yinyo_access_key = $(shell grep store_access_key configs/secrets-yinyo-server.env | cut -d "=" -f 2)
minio_yinyo_secret_key = $(shell grep store_secret_key configs/secrets-yinyo-server.env | cut -d "=" -f 2)

buckets:
	echo "Waiting for Minio to start up..."
	kubectl wait --for condition=ready pod -l app=minio --timeout=60s
	echo "Minio is running..."
	mc config host add minio http://localhost:9000 $(minio_access_key) $(minio_secret_key)
	mc admin user add minio $(minio_yinyo_access_key) $(minio_yinyo_secret_key)
	mc admin policy add minio yinyo configs/minio-yinyo-policy.json
	mc admin policy set minio yinyo user=$(minio_yinyo_access_key)
	mc mb -p minio/yinyo

clean:
	minikube delete
