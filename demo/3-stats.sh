#!/bin/bash

url="https://pz-uuidgen.$PZDOMAIN/admin/stats"
echo
echo GET $url
echo "$json"

ret=$(curl -S -s -XGET "$url")

echo RETURN:
echo "$ret"
echo
