<img width="100" heigth="100" src="https://yinyo.io/logo.svg">

# Yinyo: A wonderfully simple API driven service to reliably execute many long running scrapers in a super scaleable way

- Easily run as many scrapers as you like across a cluster of machines without having to sweat the details. Powered by [Kubernetes](https://kubernetes.io/).
- Use the language and libraries you love for writing scrapers. Supports Python, JavaScript, Ruby, PHP and Perl via Heroku Buildpacks.
- Supports many different use cases through a simple, yet flexible API that can operate synchronously or asynchronously.
- Made specifically for developers of scraper systems be it open source or commercial. No chance of vendor lock-in because it's open source, Apache licensed.

[![Build Status](https://github.com/openaustralia/yinyo/workflows/test%20and%20build/badge.svg)](https://github.com/openaustralia/yinyo/actions?workflow=test%20and%20build)
[![Coverage Status](https://coveralls.io/repos/github/openaustralia/yinyo/badge.svg?t=8kV8YE)](https://coveralls.io/github/openaustralia/yinyo)
[![Go Report Card](https://goreportcard.com/badge/github.com/openaustralia/yinyo)](https://goreportcard.com/report/github.com/openaustralia/yinyo)
[![DeepSource](https://static.deepsource.io/deepsource-badge-light.svg)](https://deepsource.io/gh/openaustralia/yinyo/?ref=repository-badge)

## Who is this README for?

This README is focused on getting developers of the core system up and running. It does not yet include
a guide for people who are just interested in being users of the API.

## Table of Contents

<!-- TOC -->

- [Yinyo: A wonderfully simple API driven service to reliably execute many long running scrapers in a super scaleable way](#yinyo-a-wonderfully-simple-api-driven-service-to-reliably-execute-many-long-running-scrapers-in-a-super-scaleable-way)
  - [Who is this README for?](#who-is-this-readme-for)
  - [Table of Contents](#table-of-contents)
  - [Development: Guide to getting up and running quickly](#development-guide-to-getting-up-and-running-quickly)
    - [Main dependencies](#main-dependencies)
    - [The main bit](#the-main-bit)
  - [Getting the website running locally](#getting-the-website-running-locally)
    - [Dependencies](#dependencies)
    - [Running a local development server for the website](#running-a-local-development-server-for-the-website)
  - [The custom herokuish docker image](#the-custom-herokuish-docker-image)
  - [Notes for debugging and testing](#notes-for-debugging-and-testing)
    - [To run the tests](#to-run-the-tests)
    - [To see what's on the blob storage (Minio)](#to-see-whats-on-the-blob-storage-minio)
    - [To see what Kubernetes is doing](#to-see-what-kubernetes-is-doing)
    - [Accessing Redis](#accessing-redis)
    - [Testing callback URLs](#testing-callback-urls)
    - [Reclaiming diskspace in minikube](#reclaiming-diskspace-in-minikube)

<!-- /TOC -->

## Development: Guide to getting up and running quickly

### Main dependencies

- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/)
- [KubeDB operator](https://kubedb.com/docs/v0.13.0-rc.0/setup/install/)
- [Skaffold](https://skaffold.dev/docs/quickstart/)
- [kustomize](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md)
- [MinIO client (mc)](https://min.io/download)
- [Go 1.13](https://golang.org/dl/)

  - Ubuntu - use `make ppa` or [read instructions](https://github.com/golang/go/wiki/Ubuntu)
  - [MacOS package installer](https://golang.org/doc/install#macos)

- Yinyo's web interface needs to be accessible on [http://localhost:8080/](http://localhost:8080/). If you have something already listening on this port, you won't get any errors, but you won't be able to connect to Yinyo to start a scraper. You'll need to clear that port.

### The main bit

First, follow the links to install the [main dependencies](main-dependencies)

Start Minikube if you haven't already

```bash
make minikube
```

Run skaffold. This will build all the bits and pieces and deploy things to your local kubernetes for you. The first time it builds everything it it takes a few minutes. After that when you make any changes to the code it does everything much faster.

```bash
make skaffold
```

Leave `skaffold` running and open a new terminal window.

Now setup the storage buckets on Minio

```bash
make buckets
```

Now compile and install the binary into your GOPATH that allows you to run a scraper

```bash
make install
```

Now you're ready to run your first scraper. The first time you run this it will take a little while.

```bash
yinyo client test/scrapers/test-python --output data.sqlite
```

Now, if you run the same scraper again it should run significantly faster.

```bash
yinyo client test/scrapers/test-python --output data.sqlite
```

## Getting the website running locally

### Dependencies

There are some extra dependencies required for building the website and associated API documentation.

- [Hugo](https://gohugo.io/) v0.60.0 or later - a static website generator
- [Shins](https://github.com/Mermade/shins) - a Node.js Slate markdown renderer
- [Widdershins](https://github.com/mermade/widdershins) - Converts OpenAPI definitions to Slate. Make sure you're using a version which includes a fix for rendering callbacks https://github.com/Mermade/widdershins/commit/5d7223f070e8d295e29a3390c3d42b4798748c55. As of December 2019 this is likely to be on master and not in one of the released versions.

### Running a local development server for the website

Do this after you've installed the dependencies (above):

```bash
make website
```

Then point your web browser at [http://localhost:1313](http://localhost:1313).

## The custom herokuish docker image

The project currently depends on a custom version of the herokuish docker image [mlandauer/herokuish:for-morph-ng](https://hub.docker.com/layers/mlandauer/herokuish/for-morph-ng/images/sha256-d39b31894660dd038c05a408db260a6bb013325e843b03ae80b528477de83d92) which is built from the Github repo [mlandauer/herokuish](https://github.com/mlandauer/herokuish/tree/for-morph-ng) and pushed to docker hub manually.

There is [an open pull request](https://github.com/gliderlabs/herokuish/pull/467) to try to get the bug
fix in our modified version merged upstream.

If this PR doesn't get merged we could use a [workaround used by Dokku](https://github.com/gliderlabs/herokuish/pull/467#issue-298708746).

## Notes for debugging and testing

### To run the tests

From the top level directory:

```bash
make test
```

### To see what's on the blob storage (Minio)

Point your web browser at [http://localhost:9000](http://localhost:9000). Login with the credentials in the file `configs/secrets-minio.env`.

### To see what Kubernetes is doing

```bash
make dashboard
```

You'll want to look in the "yinyo-system" and "yinyo-scrapers" namespaces.

### Accessing Redis

```bash
> kubectl exec -it redis-0 -n yinyo-system sh
/data # redis-cli
127.0.0.1:6379> auth changeme123
OK
127.0.0.1:6379> ping
PONG
```

### Testing callback URLs

Use [webhook.site](https://webhook.site) to see calls to a specific URL in real time. Very handy.
You can run the test scraper and get the events directed to webhook.site. For example:

```bash
yinyo client test/scrapers/test-python --output data.sqlite --callback https://webhook.site/#!/uuid-specific-to-you
```

### Reclaiming diskspace in minikube

Sometimes after a while of testing and debugging the minikube VM runs out of disk space. You'll either see this as kubernetes refusing to run anything because the node is "tainted" or minio refusing to do anything because it doesn't have enough space. Luckily there is an easy way to clear space.

```bash
minikube ssh
docker system prune
```
