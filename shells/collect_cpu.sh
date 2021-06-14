#!/bin/bash
rm collectCPU.txt
rm shard*
interval=0.05
cnt=20
n=0
sum=0.0
while [ $n -lt $cnt ]
# while true
do 
    cpu=`ps h -o pcpu -C pow`
    echo "CPU: $cpu" >> collectCPU.txt
    echo "CPU: $cpu"
    let n+=1
    sum=$(echo "$cpu+$sum"|bc)
    echo $sum
    sleep $interval
done

end=$(printf "%.2f" `echo "scale=2; $sum / $cnt"|bc`)
echo "END: $end" >> collectCPU.txt
echo "END: $end"