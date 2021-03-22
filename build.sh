#!/bin/bash

TAG="highloadcup-goldrush"

cd src

docker build -t ${TAG} .
docker tag  ${TAG} stor.highloadcup.ru/rally/cat_shooter
docker push stor.highloadcup.ru/rally/cat_shooter
