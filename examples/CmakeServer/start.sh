#!/bin/bash

cp -rf ../../hack/cmake .
mkdir build
cd build
cmake ..
make
cd -
./build/bin/CmakeServer --config=EchoTestServer.conf
