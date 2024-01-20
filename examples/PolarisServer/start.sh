#!/bin/bash
set -ex
make
./PolarisServer --config=config/config.conf
