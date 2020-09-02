#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-observation-api
  make audit
popd   