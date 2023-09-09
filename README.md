# Python Telegram Management Script


### Objective

Develop a script for a Telegram bot that finds all messages written in public chats and provide a function to remove it.


### Abstract & Considerations

Something worth mentioning before we start because it's not specified how script above supposed to be integrated. Besides business logic mentioned in "Objective" section, we need to supply some additional features:
1. Input&Output: handle external arguments and pass it to script as well as persist script results.
2. Robustness/Error handling.
3. Deploying as a one-time running script or as a worker to k8s.
4. CI

#### Logic view

What do final script need to do? 
1. Take a single or a list of Telegram API tokens of bots from some source.
2. Fetch and persist list of all messages with metadata from public chats for given bots.
3. Remove if requested to.
4. Store results and errors somewhere.
5. Give the end-user possibility to fetch results from storage.


Main script would be called `manage` and will allow to:


``` mermaid
%%{ init: { 'flowchart': { 'curve': 'stepAfter' } } }%%
flowchart TD
    start(Manage script)
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

#### Process view

- Business logic is implemented in Python. 
- Input data for script is passed via Environmental variables, which are validated w/ Pydantic. 
- Requests to Telegram API are done with pure `requests` library. 

