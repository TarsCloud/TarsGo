#!/bin/bash
set -ex
make
./PushServer --config=config.conf
