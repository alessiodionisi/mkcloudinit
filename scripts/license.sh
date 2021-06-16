#!/usr/bin/env bash

for i in $(find * -iname "*.go")
do
  if ! grep -q Copyright $i
  then
    cat ./scripts/license.txt $i >$i.new && mv $i.new $i
  fi
done
