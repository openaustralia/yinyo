## Some tests on what a new morph.io might look like

### Guide to getting up and running quickly

#### Install dependencies

* Docker - On OS X use [Docker Desktop](https://docs.docker.com/docker-for-mac/install/). On Linux install [Docker Engine](https://docs.docker.com/install/linux/docker-ce/ubuntu/).
* Kubernetes - On OS X Kubernetes comes with Docker Desktop. You just need to enable it. For Linux use something like [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/).
* The [Go language](https://golang.org/doc/install) (For Ubuntu you can use [the PPA](https://github.com/golang/go/wiki/Ubuntu))
  * You'll need at least v1.11. There's [a bug](https://github.com/golang/go/issues/27215) that needs to be worked around for now; it should be fixed when v1.13 lands. In the meantime, you'll need to install a few things by hand:
  ````
  go get github.com/jteeuwen/go-bindata/
  go get github.com/progrium/gh-release/
  go get github.com/progrium/basht/
  ````
  * By default, on linux at least, go will be installing binaries into somewhere like `~/go/bin` or `/usr/local/go/bin` - the exact location depends on how you installed go. Find that location and add it to your path, if the installer didn't do it for you. `which go-bindata` should be able to find the `go-bindata` binary before you proceed.
* Clone https://github.com/mlandauer/herokuish; Change into the directory, then `git checkout -b only_copy_to_app_path_on_build origin/only_copy_to_app_path_on_build` to switch to our patched branch. 
  * If you're using `minikube` you'll need to set your docker context to use the minikube daemon so that the images you're about to build end up there, rather than in your local docker daemon. Run `eval $(minikube docker-env)`.
  * Run `go-bindata include`, then `make deps`, then `make build`.
* [Skaffold](https://skaffold.dev/docs/getting-started/)

#### The main bit

Run skaffold. This will build all the bits and pieces and deploy things to your local kubernetes for you. The first time it builds everything it it takes a few minutes. After that when you make any changes to the code it does everything much faster.
```
skaffold dev --cache-artifacts=true
```

Leave `skaffold dev` running and open a new terminal window.

One of things that's now running is [MinIO](https://min.io/). To access it go to http://localhost:9000. Login with username `admin` and password `changeme`.

Now, create a bucket called `clay` and a bucket called `morph`. You can do that from the control at the bottom right.

You can also check that the clay server is up and running. Go to http://localhost:8080. You should see a message letting you know that all is well and good.

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
