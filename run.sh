#!/bin/bash

echo "Building test container..."
docker build . -t lit-docker-tester

echo "Stopping existing containers..."
docker-compose down > /dev/null

echo "Clearing data directory..."
sudo rm -rf data

echo "Creating new data directories..."
mkdir -p data/lit1
mkdir -p data/lit2
mkdir -p data/bitcoind 

echo "Copying template config files..."
cp lit.conf data/lit1
cp lit.conf data/lit2
cp bitcoin.conf data/bitcoind 

echo "Generating new LIT private keys..."
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit1/privkey.hex
hexdump -n 32 -e '8/4 "%08x"' /dev/urandom > data/lit2/privkey.hex

echo "Starting containers..."
docker-compose -f docker-compose-boot.yml up -d

sleep 2

echo "Mining the first 200 blocks"
docker exec -ti litdockertester_litbtcregtest_1 bitcoin-cli generate 200

echo "Booting the rest of the containers"
docker-compose up -d
docker logs litdockertester_littest_1 -f