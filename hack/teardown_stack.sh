#!/usr/bin/env bash

docker rm -f prometheus grafana status-server
docker network rm prometheus
