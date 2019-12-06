.PHONY: image server test build ppa run website apidocs skaffold kubectl minikube kustomize mc ppa go kubedb provision

all: build

build:
	go build ./...

run: install
	yinyo client test/scrapers/test-python --output data.sqlite

test: install
	go test ./...

install:
	go install ./...

ppa: go

kubedb:
	curl -fsSL https://github.com/kubedb/installer/raw/v0.13.0-rc.0/deploy/kubedb.sh | bash

mocks:
	mockery -all -inpkg

website: apidocs
	# Starts a development web server at http://localhost:1313
	hugo server -s site -D

apidocs:
	widdershins --summary openapi/definition.yaml -o openapi/definition.md
	shins openapi/definition.md -o site/content/api.html --inline

minio_access_key = $(shell grep access_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_secret_key = $(shell grep secret_key configs/secrets-minio.env | cut -d "=" -f 2)
minio_yinyo_access_key = $(shell grep store_access_key configs/secrets-yinyo-server.env | cut -d "=" -f 2)
minio_yinyo_secret_key = $(shell grep store_secret_key configs/secrets-yinyo-server.env | cut -d "=" -f 2)

/usr/local/bin/kubectl:
	curl -#LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
	chmod +x ./kubectl
	sudo install ./kubectl /usr/local/bin/

/usr/local/bin/minikube:
	curl -#Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
	chmod +x minikube
	sudo install ./minikube /usr/local/bin/

/usr/local/bin/skaffold:
	curl -#Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
	chmod +x skaffold
	sudo install ./skaffold /usr/local/bin

/usr/local/bin/kustomize:
	curl -#LO https://api.github.com/repos/kubernetes-sigs/kustomize/releases | grep browser_download | grep linux | cut -d '"' -f 4 | grep /kustomize/v | sort | tail -n 1 | xargs curl -sSOL
	tar xzf ./kustomize_v*_linux_amd64.tar.gz
	chmod +x kustomize
	sudo install ./kustomize /usr/local/bin

/usr/local/bin/mc:
	curl -#Lo mc https://dl.min.io/client/mc/release/linux-amd64/mc
	chmod +x mc
	sudo install ./mc /usr/local/bin

/usr/bin/go:
	sudo add-apt-repository ppa:longsleep/golang-backports
	sudo apt-get update
	sudo apt-get install -y golang-go

/usr/bin/docker:
	curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
	sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $$(lsb_release -cs) stable"
	sudo apt install -y socat docker-ce docker-ce-cli containerd.io

go: /usr/bin/go

kustomize: /usr/local/bin/kustomize

kubectl: /usr/local/bin/kubectl

minikube: docker kubectl /usr/local/bin/minikube
	sudo minikube start --vm-driver=none  --kubernetes-version='v1.15.2'
	
skaffold: /usr/local/bin/skaffold

mc: /usr/local/bin/mc docker kubedb

docker: /usr/bin/docker

provision: docker minikube skaffold kustomize mc go
	curl -fsSL https://github.com/kubedb/installer/raw/v0.13.0-rc.0/deploy/kubedb.sh | bash
	touch provision


buckets:
	echo "Waiting for Minio to start up..."
	kubectl wait --for condition=ready pod -l app=minio --timeout=60s --namespace yinyo-system
	echo "Minio is running..."
	mc config host add minio http://localhost:9000 $(minio_access_key) $(minio_secret_key)
	mc admin user add minio $(minio_yinyo_access_key) $(minio_yinyo_secret_key)
	mc admin policy add minio yinyo configs/minio-yinyo-policy.json
	mc admin policy set minio yinyo user=$(minio_yinyo_access_key)
	mc mb -p minio/yinyo
