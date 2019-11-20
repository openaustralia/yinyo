# Some tests on what a new morph.io might look like

## Guide to getting up and running quickly

### Main dependencies

- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/).
- [KubeDB operator](https://kubedb.com/docs/v0.13.0-rc.0/setup/install/)
- [Skaffold](https://skaffold.dev/docs/quickstart/)
- [kustomize](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md)
- [jq](https://stedolan.github.io/jq/)
- [MinIO client (mc)](https://min.io/download)
- [Go 1.13](https://golang.org/dl/)

- Clay's web interface needs to be accessible on [http://localhost:8080/](http://localhost:8080/). If you have something already listening on this port, you won't get any errors, but you won't be able to connect to Clay to start a scraper. You'll need to clear that port.

### The main bit

First, follow the links to install the [main dependencies](main-dependencies)

Start Minikube if you haven't already

```bash
minikube start --memory=3072 --disk-size='30gb' --kubernetes-version='v1.15.2'
```

Minikube by default starts with 2GB of memory and 20GB of disk space for the VM which is not enough in
our case.

Now, [install the KubeDB operator](https://kubedb.com/docs/v0.13.0-rc.0/setup/install/).

Run skaffold. This will build all the bits and pieces and deploy things to your local kubernetes for you. The first time it builds everything it it takes a few minutes. After that when you make any changes to the code it does everything much faster.

```bash
skaffold dev --port-forward=true
```

Leave `skaffold dev` running and open a new terminal window.

Now setup the storage buckets on Minio

```bash
make buckets
```

This might not work immediately because Minio might not be ready

Now compile and install the binary into your GOPATH that allows you to run a scraper

```bash
make install
```

Now you're ready to run your first scraper. The first time you run this it will take a little while.

```bash
clay client test/scrapers/test-python data.sqlite
```

Now, if you run the same scraper again it should run significantly faster.

```bash
clay client test/scrapers/test-python data.sqlite
```

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
minikube dashboard
```

You'll want to look in the "clay-system" and "clay-scrapers" namespaces.

### Accessing Redis

```bash
> kubectl exec -it redis-0 -n clay-system sh
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
clay client test/scrapers/test-python data.sqlite https://webhook.site/#!/uuid-specific-to-you
```

### Reclaiming diskspace in minikube

Sometimes after a while of testing and debugging the minikube VM runs out of disk space. You'll either see this as kubernetes refusing to run anything because the node is "tainted" or minio refusing to do anything because it doesn't have enough space. Luckily there is an easy way to clear space.

```bash
minikube ssh
docker system prune
```
