# tracks

---

*tracks* is part of the *gostream* project. *gostream* is simple music database. *tracks* is a service for track management.

Features:

- query tracks
- create tracks
- update tracks
- delete tracks

---

## Quickstart

For a quick start with *gostream*, use the official deployment repository: [deployment](https://github.com/gostream-official/deployment)

For a quick start with *tracks*, use the official docker container:

```sh
$ docker pull ghcr.io/gostream-official/tracks:latest
```

or start with a docker-compose file:

```yml
version: '3.8'

services:

  tracks:
    image: ghcr.io/gostream-official/tracks:latest
    container_name: tracks
    environment:
      MONGO_USERNAME: root
      MONGO_PASSWORD: example
      MONGO_HOST: mongo:27017
    ports:
      - "9871:9871"

  mongo:
    image: mongo:latest
    container_name: mongo
    ports:
      - 27017:27017
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: example
```

## Setup

To get *tracks* up and running, follow the instructions below.

### Platforms

Officially supported development platforms are:

- Windows
- MacOS
- Linux

### Go

The *tracks* project is written in *Go*, hence it is required to install *Go*. For the latest version of *Go*, check: https://go.dev/doc/install

## Build and Run

Build the *tracks* project using:

```sh
$ go build -o bin/tracks cmd/main.go
```

Run the *tracks* project using:

```sh
$ MONGO_USERNAME=root MONGO_PASSWORD=example go run cmd/main.go
```

## Debugging

Debug the *tracks* project using the provided `launch.json` file for *Visual Studio Code*.

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/main.go",
      "showLog": true,
      "internalConsoleOptions": "openOnSessionStart",
      "env": {
        "MONGO_USERNAME": "root",
        "MONGO_PASSWORD": "example",
        "MONGO_HOST": "127.0.0.1:27017"
      }
    }
  ]
}
```