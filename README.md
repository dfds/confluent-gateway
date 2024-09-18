# Confluent Gateway
[![Build Status](https://dfds.visualstudio.com/CloudEngineering/_apis/build/status%2Fconfluent-gateway)][https://dfds.visualstudio.com/CloudEngineering/_apis/build/status%2Fconfluent-gateway]

A dedicated service in Golang made for interfacing with the third party service Confluent Cloud.
Reading from this service happens through the REST-like API.
State-changing communication happens asynchronously through Kafka only.

## How to run
The following steps describe how to start an instance of the confluent gateway on your local machine.
This instance communicates with a custom dummy version of Confluent Cloud.

These commands assume the working directory to be the root of the repository.

1. start dependencies using
```
docker compose up -d --build
```

This will start a fake Confluent Cloud (::5051), postgres (::5432), kafka (::9092), and some supporting containers.
This may take a while depending on which images are available locally.
If it fails it's usually because the ports are already in use (often by other docker containers you have running).


2. once the above are up and running, run
```
make run
```
to start the confluent gateway.
The confluent gateway uses kafka to listen for incoming requests. It also exposes an API to support some queries.

## Checks and tests
To quickly check for issues simply run ```make build``` and ```make tests```.

There is currently no dedicated testing environment involving an actual connection to an actual instance of Confluent Cloud.

Please note, that there are no tests for Kubernetes setup currently, and that misconfiguring this may break production even if tests are passing.

## Issue and branch workflow
The process for managing changes in this process uses both Github issues, branching conventions, and pull requests.
Please refer to the [Selfservice Development Guide](https://wiki.dfds.cloud/en/ce-private/selfservice/development) for the full story.

