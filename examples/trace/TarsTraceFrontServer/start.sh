#!/bin/bash
set -ex
make
./TarsTraceFrontServer --config=config.conf
