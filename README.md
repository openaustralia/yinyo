## Some tests on what a new morph.io might look like

### Guide to getting up and running quickly

#### Main dependencies

* [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/).
* [Skaffold](https://skaffold.dev/docs/getting-started/)
* [kustomize](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md)
* [jq](https://stedolan.github.io/jq/)
* [MinIO client (mc)](https://min.io/download)

#### The main bit

Start Minikube if you haven't already
```
minikube start
```

Run skaffold. This will build all the bits and pieces and deploy things to your local kubernetes for you. The first time it builds everything it it takes a few minutes. After that when you make any changes to the code it does everything much faster.
```
skaffold dev --port-forward=true
```

Leave `skaffold dev` running and open a new terminal window.

Now setup the storage buckets on Minio
```
make buckets
```
This might not work immediately because Minio might not be ready

Now you're ready to run your first scraper.

```
./morph.sh morph-test-scrapers/test-python
```

The first time you run this it will take a little while (and you'll probably see some messages about some keys not existing. You can ignore that).

Now, if you run the same scraper again

```
./morph.sh morph-test-scrapers/test-python
```

It should run significantly faster.

### To see what's on the blob storage (Minio)

Point your web browser at http://localhost:9000. Login with the credentials in the file `secrets-minio.env`.

### To see what Kubernetes is doing

```
minikube dashboard
```
You'll want to look in the "clay-system" and "clay-scrapers" namespaces.
