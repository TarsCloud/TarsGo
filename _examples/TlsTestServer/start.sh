#!/bin/bash
set -ex
make
./TlsTestServer --config=config.conf
