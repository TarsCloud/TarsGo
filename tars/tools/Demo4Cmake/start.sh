#!/bin/bash
set -ex
mkdir -p build
cd build
cmake ..
make
cd -
./build/bin/_SERVER_ --config=config.conf
