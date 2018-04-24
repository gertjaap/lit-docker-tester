#!/bin/bash
set -e


echo "Building test container..."
docker build . --quiet -t lit-docker-tester 

echo "Setting up everything..."
docker-compose down > /dev/null 2> /dev/null

sudo rm -rf data
mkdir -p data/lit1
mkdir -p data/lit2
mkdir -p data/bitcoind 
cp lit.conf data/lit1
cp lit.conf data/lit2
cp bitcoin.conf data/bitcoind 
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit1/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit2/privkey.hex
docker-compose -f docker-compose-boot.yml up -d > /dev/null 2> /dev/null
sleep 1
docker exec -ti litdockertester_litbtcregtest_1 bitcoin-cli generate 200 > /dev/null
docker-compose up -d > /dev/null 2> /dev/null

terminator -l littest > /dev/null 2> /dev/null &
