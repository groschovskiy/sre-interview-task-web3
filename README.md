# Interview Cloud Load Balancer

## Overview

This project implements a lightweight load balancer written in Go Lang. It supports distributing incoming HTTP and WebSocket traffic across multiple backend servers.

* HTTP and WebSocket Support: Handles both HTTP requests and WebSocket connections.
* Load Balancing Algorithm: Currently supports a configurable load balancing algorithm (default: round robin).
* Health Checks: Monitors the health of backend servers and removes unhealthy ones from the pool. (Nice to be have)
* Configuration: Configuration options for load balancing algorithm, health check intervals, and backend server addresses.

## Getting Started

### Prerequisites

* Git
* Make
* Go 1.22.3 or later (https://go.dev/doc/install)

### Clone the repository

```
git clone git@github.com:groschovskiy/sre-interview-task-web3.git
```

### Build the load balancer

```
cd sre-interview-task-web3
go mod tidy
go build -o cloud-lb ./cmd/main.go
```

### Run the load balancer

```
./cloud-lb -config config.json
```

### Run the testing containers
```
cd scripts
docker-compose up -d
```

## Go Directories

### `/cmd`

Main applications for this project.

### `/scripts`

Scripts to perform various build, install, analysis, etc operations. For a seamless testing experience, consider leveraging the `docker-compose.yaml` file, which provides an additional layer of flexibility and ease-of-use for testing your backend services.

### `/internal`

Private application and library code.

### `/configs`

Configuration file templates or default configs.
