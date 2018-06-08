#!/bin/bash
set -e


echo "Building test container..."
docker build . --quiet -t lit-docker-tester 

echo "Setting up everything..."
docker-compose down > /dev/null 2> /dev/null

sudo rm -rf data
mkdir -p data/lit1
mkdir -p data/lit2
mkdir -p data/lit3
mkdir -p data/lit4
mkdir -p data/lit5
mkdir -p data/lit6
mkdir -p data/lit7
mkdir -p data/lit8
mkdir -p data/lit9
mkdir -p data/lit10
mkdir -p data/bitcoind 
cp lit.conf data/lit1
cp lit.conf data/lit2
cp lit.conf data/lit3
cp lit.conf data/lit4
cp lit.conf data/lit5
cp lit.conf data/lit6
cp lit.conf data/lit7
cp lit.conf data/lit8
cp lit.conf data/lit9
cp lit.conf data/lit10
cp bitcoin.conf data/bitcoind 
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit1/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit2/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit3/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit4/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit5/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit6/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit7/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit8/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit9/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit10/privkey.hex
docker-compose -f docker-compose-boot.yml up -d > /dev/null 2> /dev/null
sleep 1
docker exec -ti litdockertester_litbtcregtest_1 bitcoin-cli generate 200 > /dev/null
docker-compose up -d > /dev/null 2> /dev/null

docker logs litdockertester_littest_1 -f

