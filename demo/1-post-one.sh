#!/bin/bash

url="https://pz-uuidgen.$PZDOMAIN/uuids"
echo
echo POST $url
echo "$json"

ret=$(curl -S -s -XPOST "$url")

echo RETURN:
echo "$ret"
echo
