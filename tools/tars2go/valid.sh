#!/bin/bash

bin=$(pwd)/tars2go
cd /home
#cp -r tafjce _tafjce
cd tafjce
dirs=$(find . -iname '*.jce' | sed -E 's/[^\/]+jce$//' | sort -u)
root=$(pwd)
for d in $dirs
do
  echo parseing jce in $d
  cd $root
  cd $d
  $bin *.jce | grep -v 'parse include' | grep -v '文件读取错误' | grep -v '不要重复定义module'
  find . -iname '*.go' -print0 | xargs -0 rm
done
