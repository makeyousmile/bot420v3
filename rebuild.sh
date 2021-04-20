#!/bin/bash
rm bot420
git clone https://github.com/makeyousmile/bot420v3.git
cd bot420
docker build . -t bot420
docker rm -f bot
docker run --name bot --network host -d bot420 -tor 127.0.0.0:666