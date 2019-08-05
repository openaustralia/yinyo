## Some tests on what a new morph.io might look like

### Dependencies

* Docker
* https://github.com/mlandauer/herokuish/tree/only_copy_to_app_path_on_build - this is a fork of https://github.com/gliderlabs/herokuish - to build local docker dev image run `make build`.
* Kubernetes
* MinIO - Install by `make minio`

### What is the purpose of MinIO?

In their words [MinIO](https://min.io/) is "The 100% Open Source, Enterprise-Grade,
Amazon S3 Compatible Object Storage". We'll be using it to store sqlite databases,
caches for compiling scrapers, buildpack resources (mirroring S3) and backups.

If we end up deploying this whole thing to AWS then of course we could just use S3
instead but at least by taking this approach at the outset we're not locking
ourselves in to a particular bit of proprietary software.

Once Minio is up and running on Kubernetes by running `make minio` you can access the web UI via http://localhost:9000. You'll need to use the access and secret keys listed in `kubernetes/minio-deployment.yaml`.
