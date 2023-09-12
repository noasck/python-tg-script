# Python Telegram Management Script


### Objective

Develop a script for a Telegram bot that finds all messages written in public chats and provide a function to remove it.


### Abstract & Considerations

There are 2 sections, that you would be interested at as a script user:
1. [Deployment](#deployment)
2. [Development](#development)

Something worth mentioning before we start because it's not specified how script above supposed to be integrated. Besides business logic mentioned in "Objective" section, we need to supply some additional features:
1. Input&Output: handle external arguments and pass it to script as well as persist script results.
2. Robustness/Error handling.
3. Deploying as a one-time running script.
4. CI

In addition, [Telegram BOT API](https://core.telegram.org/bots/api) Doesn't support any kinda fetching of message history, so we would use Telegram API (that serves custom TG clients). For obtaining token, please refer to the section [**Deployment**](#deployment) below.

#### Logic view


What do final script need to do? 
1. Take a single Telegram API session (passed as encrypted string)
2. Fetch and persist list of all messages with metadata from public chats for given bots.
3. Remove messages, if needed.
4. Store results and errors somewhere.
5. Give the end-user possibility to fetch results from storage.


#### Process view

Consists of 2 major parts:
##### Golang script `deploy`

- CLI script
- Handles user input as arguments
- Deploys script to k8s cluster
- Manages persistent volume if needed
- Loads script output to user's machine.

##### Python script `manage`
- Contains business logic
- Could be runned as a standalone script.
- Implements Protocols for Input/Output to 
- Handles throttling
- Validates input

#### Development view
- Host code on Github.
- Use Docker CR.
- Use Github Actions as a CI: on every merge request to main:
    - Build image
    - Bump version
    - Run tests
    - Commit new version to `Version` file in repo.
    - Retag image with new version.
    - Push image to Docker Hub and update latest image.
- Use Readme Driven Development.


#### Use cases
Main script is called `deploy` and allows to:


``` mermaid
%%{ init: { 'flowchart': { 'curve': 'stepAfter' } } }%%
flowchart TD
    start(`./deploy` script)
    ctor[start `manage`\n script]
    fetch[fetch Messages batched]
    fetch2[fetch Messages batched]
    pers{Need to store results?}
    pers_proc[Store results]
    load[Load to the \n local machine]
    clean[clean resourses]
    rem[remove Messages batched]
    fetch_or_remove{Do we need to \n fetch or remove \nmessages?}

    start --> ctor --> fetch_or_remove --> |Fetch All Messages| fetch --> pers --> |YES| pers_proc -.-> load
    fetch_or_remove --> |Remove All| fetch2 --> rem --> pers 
    fetch_or_remove --> |Remove by list of ids| rem
    pers --> |NO| clean 
```

### Deployment

#### Requirements

- USER Telegram Account, bound to some number.
- OS: \*nix or Windows probably
- Docker engine v24+
- k8s cluster locally or remotely with kubeconfig file available


#### 1. Obtaining session file
Follow this steps to get session file, used in script to operate with Telegram.
- Go to [Telegram Application creation](https://my.telegram.org/apps) page. Authorize and register an application here in `API Development tool`.
- Create file `./configs/.env.dev` with these lines:
``` bash
    TG_MANAGE_API_ID=<PASTE API ID HERE>
    TG_MANAGE_API_HASH=<PASTE API HASH HERE>
    TG_MANAGE_BOT_TOKEN=None#just leave it none
```
- Run `make login`.
- Enter the phone number (only digits, including country code)
- Enter one-time code that you could find on the device.
- If job succeded, you should be able to see new user session under `./sessions` folder.

Generated session file could be used for deploying various scripts 

#### 2. Deploying to k8s

If you chose a `-persist` option, it will create Persistent Volume and Persistent Volume Claim. 
You could fetch result from it later. In this case, please pass custom filename.

1. Open Github Releases page
2. Download latest `deploy` script binary for your platform.
3. Run a command `./deploy --help` and fill all required params

Example command to run a script:
``` bash
    ./deploy -kubeconfig ~/.kube/config \
    -session \../../sessions/userbot_2023-09-12 \
    -api-id 23724365 -api-hash=59ba78a315db19c7f83cbef4f4aa3d95 \
    --remove-all true
```

You could obtain results of fetched messages from PVC like:
``` bash
    kubectl cp {namespace}/{pod_name}:{source_path} {destination_path}
```

### Development

#### Requirements
- OS: \*nix or Windows probably
- Python 3.11+
- Go language runtime or go compiler
- Docker engine v24+


#### Makefile options
Run `make <command>`, where command could be the following:

- `format` will format python script `manage` with black and ruff. Optionally use `black` or `ruff` separately.
- `test` runs test suite for `manage` script.
- `build` builds the latest image of python dockerized script.
- `clean` removes recently runned container. If you need to force remove specific one, please add option `containername=<put-whatever>`.

Specific path to run pytest on specific folder could be passed as:
``` sh
make testpath=/sample/module/to/test test
```

#### Running script locally

First, you need to change your `./configs/.env.dev` file:
```
TG_MANAGE_API_ID=#<PASTE A REAL VALUE HERE>
TG_MANAGE_API_HASH=#<PASTE A REAL VALUE HERE>
TG_MANAGE_BOT_TOKEN=None
TG_MANAGE_PERSIST_PATH=./results/sample.result
TG_MANAGE_REMOVE_ALL=True # set to False if you don't want to remove messages
```

You could optionally specify:

```
TG_MANAGE_REMOVE_MESSAGE_IDS=[]#[Paste,IDs,here,in,a,list,of,integeres]
TG_MANAGE_REMOVE_CHAT_ID=None#Change to real chat id to remove messages from
```

Then run make command (linux):

``` bash
    make bottoken="$(cat ./sessions/$(ls ./sessions/ | head -n 1))" local
```
This will take first your stored session from `./sessions` folder.

## How to make it better
- [ ] Use 2-stage docker builds
- [ ] Wrap python script in setuptools with `sh` entrypoint.
- [ ] Handle logs 
