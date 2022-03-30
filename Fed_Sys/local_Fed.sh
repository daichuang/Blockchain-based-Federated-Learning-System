#!/bin/bash
start_time=$(date +%s)
clear

#Flushing out all iptables blocked rules
sudo iptables -F INPUT
sudo iptables -F OUTPUT

let nodes=$1
dataset=$2
# additionalArgs="-sa=true -vp=true -np=false -na=3 -nv=3 -nn=2 -ep=1"
argList=($additionalArgs)
let cnt=0

# #TODO: Take as input dataset name. Figure out dimensions based on name
# if [[ "$dataset" = "mnist" ]]; then
# 	dimensions=7850
# #	dimensions=164266
# elif [[ "$dataset" = "creditcard" ]]; then
# 	dimensions=25
# fi

echo $nodes
echo $dataset

# Single command that kills them
pkill Fed_Sys

cd ../Fed_Sys
go build

# Purge the logs
rm -f ./Fed_log/*.log

# #---------------------------------------------------Test 1: All nodes online--------------------------------------------------------------------

echo "Running tests: No failure case. All nodes online"

for (( totalnodes = $nodes; totalnodes < ($nodes + 1); totalnodes++ )); do
	
	echo "Running with" $totalnodes "nodes"

	for (( index = 0; index < totalnodes; index++ )); do
		
		LogFile=log_$index\_$totalnodes.log

		myAddress=127.0.0.1
		let thisPort=8000+$index
		echo $index
		echo $thisPort
		echo $myAddress

		commandToRun="./Fed_Sys -index=${index} -off_private_train=false --epsilon=2  -total_nodes=${totalnodes} -dataset=${dataset}  ${argList[@]}"
		# commandToRun="./Fed_Sys -index=${index} -off_private_train=true -byzantine_thresh=0.1 -total_nodes=${totalnodes} -dataset=${dataset}  ${argList[@]}"
		commandList=($commandToRun)
		echo "${commandList[@]}" 
		"${commandList[@]}"  2> ./Fed_log/$LogFile & 
	done

	wait
	echo "============ over ============"

done

end_time=$(date +%s)
cost_time=$[ $end_time-$start_time ]
echo "BlockFedurateSys Total  time is $(($cost_time/60))min $(($cost_time%60))s"
