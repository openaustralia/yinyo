## Some tests on what a new morph.io might look like

### Guide to getting up and running quickly

#### Main dependencies

* [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/).
* [Skaffold](https://skaffold.dev/docs/getting-started/)
* [jq](https://stedolan.github.io/jq/)
* [MinIO client (mc)](https://min.io/download)

#### Install herokuish

First checkout our patched branch of herokuish
```
git clone https://github.com/mlandauer/herokuish
cd herokuish
git checkout for-morph-ng
```

Make your local docker client use the docker daemon on minikube
```
eval $(minikube docker-env)
```

Then, build the `herokuish:dev` docker image with
```
make build-in-docker
```

#### The main bit

Run skaffold. This will build all the bits and pieces and deploy things to your local kubernetes for you. The first time it builds everything it it takes a few minutes. After that when you make any changes to the code it does everything much faster.
```
skaffold dev --cache-artifacts=true
```

Leave `skaffold dev` running and open a new terminal window.

One of things that's now running is [MinIO](https://min.io/). To access it
```
minikube service minio-service -n clay
```
This will open your web browser at the url for Minio running on minikube. Login with the username `admin` and password `changeme`.

Now, create a bucket called `clay` and a bucket called `morph`. You can do that from the control at the bottom right.

You can also check that the clay server is up and running.
```
minikube service clay-server -n clay
```
You should see a message letting you know that all is well and good.

Now you're ready to run your first scraper.

```
./morph.sh morph-test-scrapers/test-python
```

The first time you run this it will take a little while (and you'll probably see some messages about some keys not existing. You can ignore that). If you go back to MinIO you'll see that morph has saved an sqlite database and that clay has saved a cache of the build.

Now, if you run the same scraper again

```
./morph.sh morph-test-scrapers/test-python
```

It should run significantly faster.
