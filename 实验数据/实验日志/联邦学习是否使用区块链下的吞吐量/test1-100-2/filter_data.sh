#!/bin/bash
mkdir -p filter_data
cat all.log | grep "aggregate global" | awk '{print $1}' > filter_data/aggregate.txt
cat all.log | grep -e "to aggregate" -e "aggregate global" | awk '{print $1}' > filter_data/wait.txt
logs=$(ls | grep 100)
for log in $logs
do
	cat $log | grep "start iteration" | awk '{print $1}' > 'filter_data/'"$log"'_iteration.txt'
	cat $log | grep "validate" | awk '{print $1}' > 'filter_data/'"$log"'_validate.txt'
	cat $log | grep "trainer/model/train" | awk '{print $1}' > 'filter_data/'"$log"'_train.txt'
done
