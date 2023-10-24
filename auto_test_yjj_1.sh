#!/bin/bash
export PYTHONPATH=$PWD/Flamby:$PYTHONPATH
export IPFS_PATH=/mnt/E/ipfs
nohup python ./Flamby/flamby/benchmarks/server.py 1 &
sleep 10
./main 1 30
mykill flamby
nohup python ./Flamby/flamby/benchmarks/server.py 5 &
sleep 10
./main 1 50
mykill flamby
nohup python ./Flamby/flamby/benchmarks/server.py 10 &
sleep 10
./main 1 75
mykill flamby
nohup python ./Flamby/flamby/benchmarks/server.py 15 &
sleep 10
./main 1 100
mykill flamby
nohup python ./Flamby/flamby/benchmarks/server.py 20 &
sleep 10
./main 1 120
mykill flamby
