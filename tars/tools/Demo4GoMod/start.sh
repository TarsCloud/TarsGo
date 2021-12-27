#!/bin/bash
set -ex
make
./_SERVER_ --config=config.conf
