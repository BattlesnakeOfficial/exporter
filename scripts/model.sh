#!/bin/bash
apis=("engine" "snake")
for api in "${apis[@]}"
do
  echo "Generating ${api}"
  rm -rf /tmp/gomodel
  mkdir /tmp/gomodel
  docker run --rm -v /tmp/gomodel:/local openapitools/openapi-generator-cli generate  \
    -i https://raw.githubusercontent.com/battlesnakeio/docs/master/apis/${api}/spec.yaml \
    -g go \
    -o /local/ > /dev/null 2>&1
  rm -rf ${api}
  mkdir ${api}
  dir=`pwd`
  cd /tmp/gomodel
  for filename in *model*.go; do
    newFilename=`echo ${filename} | cut -c 7-`
    echo "  ${filename} -> ${newFilename}"
    sed "s/,omitempty//g" ${filename}  \
      | sed "s/package openapi/package ${api}model/g" \
      > ${dir}/${api}/${newFilename}; 
  done
  cd ${dir}
done