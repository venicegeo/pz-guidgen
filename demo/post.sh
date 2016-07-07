#!/bin/bash

d=`date`

url="https://pz-uuidgen.int.geointservices.io/uuids?count=3"
echo
echo POST $url
echo "$json"

ret=$(curl -S -s -XPOST "$url")

echo RETURN:
echo "$ret"
echo
