# Python Telegram Management Script


### Objective

Develop a script for a Telegram bot that finds all messages written in public chats and provide a function to remove it.


### Abstract & Considerations

Something worth mentioning before we start because it's not specified how script above supposed to be integrated. Besides business logic mentioned in "Objective" section, we need to supply some additional features:
1. Input&Output: handle external arguments and pass it to script as well as persist script results.
2. Robustness/Error handling.
3. Deploying as a one-time running script or as a worker to k8s.
4. CI

In addition, [Telegram BOT API](https://core.telegram.org/bots/api) Doesn't support any kinda fetching of message history, so we would use Telegram API (that servers custom TG clients). For obtaining token, please refer to the section **Deployment** below.

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
- Could be runned as a standalone script or as worker process subscribed to some source.
- Implements Adapters: PersistenceAdapter, InputAdapter -- we need to support multiple IO sources and location.
- Requests to Telegram Bot API performs with pure `requests` library. 
- If input is passed via ENVs, load ENVs via pydantic.


#### Development view
- Host code on Github.
- Use Docker CR.
- There are 2 Github Actions Jobs as a CI:
    - On merge to `main`: **Build** & **Test** & **Bump** version & **Push** to Docker Hub
    - On commit to open PR: **Build** & **Test** 
- Use Readme Driven Development.


#### Use cases
Main script is called `deploy` and allows to:


``` mermaid
%%{ init: { 'flowchart': { 'curve': 'stepAfter' } } }%%
flowchart TD
    start(Deploy script starts)
    command{What do we want?}
    create_storage[Set up persistent storage \n or use an external one]
    run_job[Run Telegram management script]
    fetch_results[Load results to local machine]
    run_job_as_worker[Run script as worker \n in e.g. consumer mode]
    run_sample_client[Optionally.\n Run sample event producer]

    clean[Clean resources]
    finish[Finish job]

    start --> command
    command --> |Run once| create_storage --> run_job --> fetch_results --> clean --> finish
    command --> |Deploy worker| run_job_as_worker --> run_sample_client --> finish
```

### Deployment

#### Requirements

- USER Telegram Account, bound to some number.
- OS: \*nix or Windows probably
- Docker engine v24+


#### Obtaining session file
Follow this steps to get session file, used in script to operate with Telegram.
- Go to [Telegram Application creation](https://my.telegram.org/apps) page. Authorize and register an application here in `API Development tool`.
- Create file `./configs/.env.dev` with these lines:
``` bash
    TG_MANAGE_API_ID=# PASTE REAL ID HERE
    TG_MANAGE_API_HASH=# PASTE HASH without quotes
    TG_MANAGE_BOT_TOKEN=leave-it-any-string
    TG_MANAGE_SECRET_KEY=leave-it-any-string
```
- Run `make login`.
- Enter the phone number (only digits, including country code)
- Enter one-time code that you could find on the device.
- If job succeded, you should be able to see new user session under `./sessions` folder.

Generated session file could be used for deploying various scripts 
 

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

## How to make it better
- [ ] Use 2-stage docker builds
- [ ] Wrap python script in setuptools with `sh` entrypoint.
- [ ] Handle logs 
