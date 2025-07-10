#!/bin/bash
mkdir -p filter_data
cat all.log | grep "aggregate global"| awk '{print $1}' > filter_data/aggregate.txt
cat all.log | grep -e "to aggregate" -e "global model begin"| awk '{print $1}' > filter_data/wait.txt
logs=$(ls | grep 100)
for log in $logs
do
	cat $log | grep "start iteration"| awk '{print $1}' > 'filter_data/'"$log"'_iteration.txt'
	cat $log | grep -e "begin to wait " -e "receive all validate"| awk '{print $1}' > 'filter_data/'"$log"'_validate.txt'
	cat $log | grep -e "begin to train the global" -e "trainnging finished, iteration" | awk '{print $1}' > 'filter_data/'"$log"'_train.txt'
done
