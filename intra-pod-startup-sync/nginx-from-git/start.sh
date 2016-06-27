#!/bin/sh

DATA=/usr/share/nginx/html/kubernetes

while [ ! -d ${DATA} ]; do
    echo "waiting for ${DATA}"
    sleep 2
done
echo "found ${DATA}"

exec nginx -g "daemon off;"
