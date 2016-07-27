#!/bin/bash

# Used for manual local testing only under OSX (with docker-machine).

docker-compose up -d

# Hack around the fact this version of mgfs isn't working in OSX correctly.
# May be fixable later, but want to be able to test for now.
docker build -t mgfs_osx_dev .

echo "Fetching mongodb container ID and port"
mongodb_docker_network=$(docker network ls | grep mgfs | awk '{print $2}')
mongodb_container_name=$(docker-compose ps mongodb | tail -n 1 | awk '{print $1}')
#mongodb_port=$(docker-compose port mongodb 27017 | cut -d: -f2)
echo "Done"

# Runs bash by default.
echo "Starting mgfs container"
docker run -it --privileged --net "${mongodb_docker_network}" -e MONGODB_HOST=$mongodb_container_name mgfs_osx_dev
