# Get Started

## 1. Go to [Github](https://github.com/openaustralia/yinyo) and follow the instructions to setup a local Yinyo instance

## 2. Run your first scraper

```bash
yinyo test/scrapers/test-python --output data.sqlite
```

This will stream the console output of the scraper straight to you

```bash
-----> Python app detected
       !     Python has released a security update! Please consider upgrading to python-2.7.16
       Learn More: https://devcenter.heroku.com/articles/python-runtimes
-----> Installing requirements with pip
       Obtaining scraperwiki from git+http://github.com/openaustralia/scraperwiki-python.git@morph_defaults#egg=scraperwiki (from -r /tmp/build/requirements.txt (line 2))
       Cloning http://github.com/openaustralia/scraperwiki-python.git (to morph_defaults) to /app/.heroku/src/scraperwiki
       Installing collected packages: scraperwiki
       Running setup.py develop for scraperwiki
       Successfully installed scraperwiki

-----> Discovering process types
       Procfile declares types -> scraper
First a little test message to stderr
Hello from test-python!
1...
2...
3...
4...
5...
```

## 3. Do it all again! But this time using the API directly, step-by-step

### 1. Create a run

```bash
curl -X POST http://localhost:8080/runs
```

You'll get a `name` and a `token` back which you'll need in the following steps

```json
{ "name": "run-qjv4t" }
```

So, to make this a bit easier with less typing, let's set an environment variable

```bash
NAME=run-qjv4t
```

(Replace the run `name` with your own value)

### 2. Tar and compress the code

```bash
tar -C test/scrapers/test-python/ -zcf code.tgz .
```

### 3. Upload the code

```bash
curl -X PUT "http://localhost:8080/runs/$NAME/app" --data-binary @code.tgz
```

### 4. Start the run

Note that we're also passing the path to the file that we want to get at the end of the run.

```bash
curl -X POST -H "Content-Type: application/json" "http://localhost:8080/runs/$NAME/start" -d '{"output":"data.sqlite"}'
```

### 5. Stream the events

```bash
curl "http://localhost:8080/runs/$NAME/events"
```

This will output a stream of events formatted as JSON in real-time

```json
{"id":"1580183421503-0","time":"2020-01-28T03:50:21.472365664Z","type":"start","data":{"stage":"build"}}
{"id":"1580183426973-0","time":"2020-01-28T03:50:26.925073211Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       \u001b[1G-----\u003e Python app detected"}}
{"id":"1580183431263-0","time":"2020-01-28T03:50:31.247105651Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       !     Python has released a security update! Please consider upgrading to python-2.7.16"}}
{"id":"1580183431275-0","time":"2020-01-28T03:50:31.264161902Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Learn More: https://devcenter.heroku.com/articles/python-runtimes"}}
{"id":"1580183431278-0","time":"2020-01-28T03:50:31.276357179Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G-----\u003e Installing python-2.7.15"}}
{"id":"1580183574302-0","time":"2020-01-28T03:52:54.291338196Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G-----\u003e Installing pip"}}
{"id":"1580183584476-0","time":"2020-01-28T03:53:04.474015597Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G-----\u003e Installing SQLite3"}}
{"id":"1580183619959-0","time":"2020-01-28T03:53:39.869423145Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G-----\u003e Installing requirements with pip"}}
{"id":"1580183620457-0","time":"2020-01-28T03:53:40.45369831Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Obtaining scraperwiki from git+http://github.com/openaustralia/scraperwiki-python.git@morph_defaults#egg=scraperwiki (from -r /tmp/build/requirements.txt (line 2))"}}
{"id":"1580183620511-0","time":"2020-01-28T03:53:40.462302015Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Cloning http://github.com/openaustralia/scraperwiki-python.git (to morph_defaults) to /app/.heroku/src/scraperwiki"}}
{"id":"1580183623775-0","time":"2020-01-28T03:53:43.77187282Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting dumptruck\u003e=0.1.2 (from scraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1580183624610-0","time":"2020-01-28T03:53:44.607800268Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/15/27/3330a343de80d6849545b6c7723f8c9a08b4b104de964ac366e7e6b318df/dumptruck-0.1.6.tar.gz"}}
{"id":"1580183624915-0","time":"2020-01-28T03:53:44.90619589Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting requests (from scraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1580183625316-0","time":"2020-01-28T03:53:45.285716896Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/51/bd/23c926cd341ea6b7dd0b2a00aba99ae0f828be89d72b2190f27c11d4b7fb/requests-2.22.0-py2.py3-none-any.whl (57kB)"}}
{"id":"1580183625486-0","time":"2020-01-28T03:53:45.476952262Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting urllib3!=1.25.0,!=1.25.1,\u003c1.26,\u003e=1.21.1 (from requests-\u003escraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1580183625818-0","time":"2020-01-28T03:53:45.816865328Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/e8/74/6e4f91745020f967d09332bb2b8b9b10090957334692eb88ea4afe91b77f/urllib3-1.25.8-py2.py3-none-any.whl (125kB)"}}
{"id":"1580183626043-0","time":"2020-01-28T03:53:46.036071959Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting certifi\u003e=2017.4.17 (from requests-\u003escraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1580183626355-0","time":"2020-01-28T03:53:46.349999527Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/b9/63/df50cac98ea0d5b006c55a399c3bf1db9da7b5a24de7890bc9cfd5dd9e99/certifi-2019.11.28-py2.py3-none-any.whl (156kB)"}}
{"id":"1580183626572-0","time":"2020-01-28T03:53:46.566282329Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting chardet\u003c3.1.0,\u003e=3.0.2 (from requests-\u003escraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1580183626881-0","time":"2020-01-28T03:53:46.865673118Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/bc/a9/01ffebfb562e4274b6487b4bb1ddec7ca55ec7510b22e4c51f14098443b8/chardet-3.0.4-py2.py3-none-any.whl (133kB)"}}
{"id":"1580183627068-0","time":"2020-01-28T03:53:47.064912201Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Collecting idna\u003c2.9,\u003e=2.5 (from requests-\u003escraperwiki-\u003e-r /tmp/build/requirements.txt (line 2))"}}
{"id":"1580183627365-0","time":"2020-01-28T03:53:47.359353914Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Downloading https://files.pythonhosted.org/packages/14/2c/cd551d81dbe15200be1cf41cd03869a46fe7226e7450af7a6545bfc474c9/idna-2.8-py2.py3-none-any.whl (58kB)"}}
{"id":"1580183627480-0","time":"2020-01-28T03:53:47.459853005Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Installing collected packages: dumptruck, urllib3, certifi, chardet, idna, requests, scraperwiki"}}
{"id":"1580183627487-0","time":"2020-01-28T03:53:47.481636522Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Running setup.py install for dumptruck: started"}}
{"id":"1580183627810-0","time":"2020-01-28T03:53:47.759041216Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Running setup.py install for dumptruck: finished with status 'done'"}}
{"id":"1580183628160-0","time":"2020-01-28T03:53:48.156023412Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Running setup.py develop for scraperwiki"}}
{"id":"1580183628416-0","time":"2020-01-28T03:53:48.412788154Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Successfully installed certifi-2019.11.28 chardet-3.0.4 dumptruck-0.1.6 idna-2.8 requests-2.22.0 scraperwiki urllib3-1.25.8"}}
{"id":"1580183628838-0","time":"2020-01-28T03:53:48.830730376Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       "}}
{"id":"1580183629805-0","time":"2020-01-28T03:53:49.771548419Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       \u001b[1G-----\u003e Discovering process types"}}
{"id":"1580183629813-0","time":"2020-01-28T03:53:49.80782406Z","type":"log","data":{"stage":"build","stream":"stdout","text":"\u001b[1G       Procfile declares types -\u003e scraper"}}
{"id":"1580183629853-0","time":"2020-01-28T03:53:49.832331568Z","type":"finish","data":{"stage":"build","exit_data":{"exit_code":0,"usage":{"wall_time":208.322458275,"cpu_time":23.252974000000002,"max_rss":77045760,"network_in":50109438,"network_out":1307044}}}}
{"id":"1580183638633-0","time":"2020-01-28T03:53:58.625907685Z","type":"start","data":{"stage":"run"}}
{"id":"1580183641377-0","time":"2020-01-28T03:54:01.368828538Z","type":"log","data":{"stage":"run","stream":"stdout","text":"Hello from test-python!"}}
{"id":"1580183641386-0","time":"2020-01-28T03:54:01.380169548Z","type":"log","data":{"stage":"run","stream":"stdout","text":"1..."}}
{"id":"1580183641411-0","time":"2020-01-28T03:54:01.398367334Z","type":"log","data":{"stage":"run","stream":"stderr","text":"First a little test message to stderr"}}
{"id":"1580183642374-0","time":"2020-01-28T03:54:02.368949967Z","type":"log","data":{"stage":"run","stream":"stdout","text":"2..."}}
{"id":"1580183643372-0","time":"2020-01-28T03:54:03.370619368Z","type":"log","data":{"stage":"run","stream":"stdout","text":"3..."}}
{"id":"1580183644373-0","time":"2020-01-28T03:54:04.371837081Z","type":"log","data":{"stage":"run","stream":"stdout","text":"4..."}}
{"id":"1580183645376-0","time":"2020-01-28T03:54:05.37319053Z","type":"log","data":{"stage":"run","stream":"stdout","text":"5..."}}
{"id":"1580183646397-0","time":"2020-01-28T03:54:06.38666964Z","type":"finish","data":{"stage":"run","exit_data":{"exit_code":0,"usage":{"wall_time":7.75252631,"cpu_time":0.384125,"max_rss":136421376,"network_in":28585,"network_out":7125}}}}
{"id":"1580183648184-0","time":"2020-01-28T03:54:08.182780901Z","type":"last","data":{}}
```

You might notice that this is taking longer than when we ran this with `yinyo`. It's having to install python and some dependencies which takes some time. That's because we've ignored caching here just to keep things a bit simpler

### 6. Get the output

Now get the output file which we chose when we started the run and save it to a local file called `data.sqlite`.

```bash
curl "http://localhost:8080/runs/$NAME/output" --output data.sqlite
```

### 7. Clean up

```bash
curl -X DELETE "http://localhost:8080/runs/$NAME"
```

## 4. Check out the [API reference](/api) to see what more you can do
