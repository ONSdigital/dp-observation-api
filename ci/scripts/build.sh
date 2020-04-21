#!/bin/bash -eux

pushd dp-observation-api
  make build
  cp build/dp-observation-api Dockerfile.concourse ../build
popd
