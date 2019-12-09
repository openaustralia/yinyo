# Some tests on what a new morph.io might look like

<!-- TOC -->

- [Some tests on what a new morph.io might look like](#some-tests-on-what-a-new-morphio-might-look-like)
    - [Guide to getting up and running quickly](#guide-to-getting-up-and-running-quickly)
        - [Main dependencies](#main-dependencies)
        - [The main bit](#the-main-bit)
    - [Notes for debugging and testing](#notes-for-debugging-and-testing)
        - [To see what's on the blob storage (Minio)](#to-see-whats-on-the-blob-storage-minio)
        - [To see what Kubernetes is doing](#to-see-what-kubernetes-is-doing)
        - [Accessing Redis](#accessing-redis)
        - [Testing callback URLs](#testing-callback-urls)
        - [Reclaiming diskspace in minikube](#reclaiming-diskspace-in-minikube)

<!-- /TOC -->

## Guide to getting up and running quickly

### Main dependencies

- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/)
- [KubeDB operator](https://kubedb.com/docs/v0.13.0-rc.0/setup/install/)
- [Skaffold](https://skaffold.dev/docs/quickstart/)
- [kustomize](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md)
- [jq](https://stedolan.github.io/jq/)
- [MinIO client (mc)](https://min.io/download)

- Clay's web interface needs to be accessible on http://localhost:8080/. If you have something already listening on this port, you won't get any errors, but you won't be able to connect to Clay to start a scraper. You'll need to clear that port.

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

In your `skaffold` window, you should now see that Clay is ready:

````
[clay-server-754b5d5fc-cks2x clay-server] main.go:273: Clay is ready and waiting.
````

Now you're ready to run your first scraper.

```bash
make run
```

The first time you run this it will take a little while (and you'll probably see some messages about some keys not existing. You can ignore that).

Now, if you run the same scraper again

```bash
make run
```

It should run significantly faster.

## Notes for debugging and testing

### To see what's on the blob storage (Minio)

Point your web browser at [http://localhost:9000](http://localhost:9000). Login with the credentials in the file `configs/secrets-minio.env`.

### To see what Kubernetes is doing

```bash
make dashboard
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
./client.sh test/scrapers/test-python data.sqlite https://webhook.site/#!/uuid-specific-to-you
```

### Reclaiming diskspace in minikube

Sometimes after a while of testing and debugging the minikube VM runs out of disk space. You'll either see this as kubernetes refusing to run anything because the node is "tainted" or minio refusing to do anything because it doesn't have enough space. Luckily there is an easy way to clear space.

```bash
minikube ssh
docker system prune
```
