#!/usr/bin/env bash

script_dir=$(dirname "${BASH_SOURCE[0]}")
abs_path=$(echo "$(cd "${script_dir}" && pwd)")

docker build -t status-server -f ${script_dir}/../build/Dockerfile ${script_dir}/../

docker network create prometheus

docker run -d --name=prometheus -p 9090:9090 --network prometheus -v ${abs_path}/prometheus.yaml:/etc/prometheus/prometheus.yml  prom/prometheus

docker run -d -p 3000:3000 --name grafana -v ${abs_path}/provisioning:/etc/grafana/provisioning --network prometheus grafana/grafana

docker run -d --name status-server --network prometheus status-server
