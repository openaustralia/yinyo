# Get Started

## 1. Go to [Github](https://github.com/openaustralia/yinyo) and follow the instructions to setup a local Yinyo instance

## 2. Run your first scraper

```bash
yinyo client test/scrapers/test-python --output data.sqlite
```

This will stream the console output of the scraper straight to you

```

```

## 3. Do it all again! But this time using the API directly, step-by-step

### 1. Create a run

```bash
curl -X POST http://localhost:8080/runs
```

You'll get a `name` and a `token` back which you'll need in the following steps

```json
{ "name": "run-qjv4t", "token": "lLsBCZiBPYcTQb439YvPbz9GC3bPcYr5" }
```

So, to make this a bit easier with less typing, let's set a couple of environment variables

```bash
NAME=run-qjv4t
TOKEN=lLsBCZiBPYcTQb439YvPbz9GC3bPcYr5
```

(Replace the run `name` and `token` with your own values)

### 2. Tar and compress the code

```bash
tar -C test/scrapers/test-python/ -zcf code.tgz .
```

### 3. Upload the code

```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" "http://localhost:8080/runs/$NAME/app" --data-binary @code.tgz
```

### 4. Start the run

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" "http://localhost:8080/runs/$NAME/start" -d "{}"
```

### 5. Stream the events

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/runs/$NAME/events"
```

This will output a stream of events formatted as JSON in real-time

```json
{"id":"1576640952345-0","time":"2019-12-18T03:49:12.285014316Z","type":"start","data":{"stage":"build"}}
{"id":"1576640962365-0","time":"2019-12-18T03:49:22.33779607Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       \u001b[1G-----\u003e Python app detected"}}
{"id":"1576640967054-0","time":"2019-12-18T03:49:27.048339914Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       !     Python has released a security update! Please consider upgrading to python-2.7.16"}}
{"id":"1576640967078-0","time":"2019-12-18T03:49:27.048378034Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Learn More: https://devcenter.heroku.com/articles/python-runtimes"}}
{"id":"1576640967107-0","time":"2019-12-18T03:49:27.048946543Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G-----\u003e Installing python-2.7.15"}}
{"id":"1576641084764-0","time":"2019-12-18T03:51:24.745406175Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G-----\u003e Installing pip"}}
{"id":"1576641099010-0","time":"2019-12-18T03:51:39.005971326Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G-----\u003e Installing SQLite3"}}
{"id":"1576641144639-0","time":"2019-12-18T03:52:24.624972935Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G-----\u003e Installing requirements with pip"}}
{"id":"1576641146085-0","time":"2019-12-18T03:52:26.070645396Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Obtaining scraperwiki from git+http://github.com/openaustralia/scraperwiki-python.git@morph_defaults#egg=scraperwiki (from -r /tmp/build/requirements.txt (line 2))"}}
{"id":"1576641146120-0","time":"2019-12-18T03:52:26.082751315Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Cloning http://github.com/openaustralia/scraperwiki-python.git (to morph_defaults) to /app/.heroku/src/scraperwiki"}}
{"id":"1576641149773-0","time":"2019-12-18T03:52:29.75256733Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting dumptruck\u003e=0.1.2 (from scraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1576641151134-0","time":"2019-12-18T03:52:31.131517475Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/15/27/3330a343de80d6849545b6c7723f8c9a08b4b104de964ac366e7e6b318df/dumptruck-0.1.6.tar.gz"}}
{"id":"1576641151488-0","time":"2019-12-18T03:52:31.483998763Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting requests (from scraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1576641151896-0","time":"2019-12-18T03:52:31.887582711Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/51/bd/23c926cd341ea6b7dd0b2a00aba99ae0f828be89d72b2190f27c11d4b7fb/requests-2.22.0-py2.py3-none-any.whl (57kB)"}}
{"id":"1576641152064-0","time":"2019-12-18T03:52:32.061817221Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting urllib3!=1.25.0,!=1.25.1,\u003c1.26,\u003e=1.21.1 (from requests-\u003escraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1576641152451-0","time":"2019-12-18T03:52:32.442629194Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/b4/40/a9837291310ee1ccc242ceb6ebfd9eb21539649f193a7c8c86ba15b98539/urllib3-1.25.7-py2.py3-none-any.whl (125kB)"}}
{"id":"1576641152701-0","time":"2019-12-18T03:52:32.69377889Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting certifi\u003e=2017.4.17 (from requests-\u003escraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1576641153017-0","time":"2019-12-18T03:52:33.016132637Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/b9/63/df50cac98ea0d5b006c55a399c3bf1db9da7b5a24de7890bc9cfd5dd9e99/certifi-2019.11.28-py2.py3-none-any.whl (156kB)"}}
{"id":"1576641153329-0","time":"2019-12-18T03:52:33.323776631Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting chardet\u003c3.1.0,\u003e=3.0.2 (from requests-\u003escraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1576641153624-0","time":"2019-12-18T03:52:33.616683475Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/bc/a9/01ffebfb562e4274b6487b4bb1ddec7ca55ec7510b22e4c51f14098443b8/chardet-3.0.4-py2.py3-none-any.whl (133kB)"}}
{"id":"1576641153939-0","time":"2019-12-18T03:52:33.937781156Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting idna\u003c2.9,\u003e=2.5 (from requests-\u003escraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1576641154265-0","time":"2019-12-18T03:52:34.257812273Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/14/2c/cd551d81dbe15200be1cf41cd03869a46fe7226e7450af7a6545bfc474c9/idna-2.8-py2.py3-none-any.whl (58kB)"}}
{"id":"1576641154388-0","time":"2019-12-18T03:52:34.382274166Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Installing collected packages: dumptruck, urllib3, certifi, chardet, idna, requests, scraperwiki"}}
{"id":"1576641154395-0","time":"2019-12-18T03:52:34.382720177Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Running setup.py install for dumptruck: started"}}
{"id":"1576641154689-0","time":"2019-12-18T03:52:34.654314862Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Running setup.py install for dumptruck: finished with status 'done'"}}
{"id":"1576641155319-0","time":"2019-12-18T03:52:35.3103251Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Running setup.py develop for scraperwiki"}}
{"id":"1576641155892-0","time":"2019-12-18T03:52:35.88213814Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Successfully installed certifi-2019.11.28 chardet-3.0.4 dumptruck-0.1.6 idna-2.8 requests-2.22.0 scraperwiki urllib3-1.25.7"}}
{"id":"1576641156439-0","time":"2019-12-18T03:52:36.431459224Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       "}}
{"id":"1576641158793-0","time":"2019-12-18T03:52:38.789674423Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       \u001b[1G-----\u003e Discovering process types"}}
{"id":"1576641158850-0","time":"2019-12-18T03:52:38.842850945Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Procfile declares types -\u003e scraper"}}
{"id":"1576641158870-0","time":"2019-12-18T03:52:38.859658314Z","type":"finish","data":{"stage":"build"}}
{"id":"1576641171742-0","time":"2019-12-18T03:52:51.731830713Z","type":"start","data":{"stage":"run"}}
{"id":"1576641175194-0","time":"2019-12-18T03:52:55.187718678Z","type":"log","data":{"stage":"run","stream":"stderr","text":"First a little test message to stderr"}}
{"id":"1576641175228-0","time":"2019-12-18T03:52:55.188223928Z","type":"log","data":{"stage":"run","stream":"stdout","text":"Hello from test-python!"}}
{"id":"1576641175242-0","time":"2019-12-18T03:52:55.188240717Z","type":"log","data":{"stage":"run","stream":"stdout","text":"1..."}}
{"id":"1576641176194-0","time":"2019-12-18T03:52:56.189727622Z","type":"log","data":{"stage":"run","stream":"stdout","text":"2..."}}
{"id":"1576641177193-0","time":"2019-12-18T03:52:57.191833268Z","type":"log","data":{"stage":"run","stream":"stdout","text":"3..."}}
{"id":"1576641178195-0","time":"2019-12-18T03:52:58.193058211Z","type":"log","data":{"stage":"run","stream":"stdout","text":"4..."}}
{"id":"1576641179195-0","time":"2019-12-18T03:52:59.193313896Z","type":"log","data":{"stage":"run","stream":"stdout","text":"5..."}}
{"id":"1576641180235-0","time":"2019-12-18T03:53:00.233747709Z","type":"finish","data":{"stage":"run"}}
{"id":"1576641180238-0","time":"2019-12-18T03:53:00.235891186Z","type":"last","data":{}}
```

You might notice that this is taking longer than when we ran this with `yinyo client`. It's having to install python and some dependencies which takes some time. That's because we've ignored caching here just to keep things a bit simpler

### 6. Clean up

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" "http://localhost:8080/runs/$NAME"
```

## 4. Check out the [API reference](/api) to see what more you can do
