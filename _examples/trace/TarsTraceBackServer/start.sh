#!/bin/bash
set -ex
make
./TarsTraceBackServer --config=config.conf
