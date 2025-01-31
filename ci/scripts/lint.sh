#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-observation-api
  make lint
popd