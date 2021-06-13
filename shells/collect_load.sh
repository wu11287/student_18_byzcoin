#!/bin/bash
rm collectLoad.txt
rm shard*
interval=1
cnt=20
n=0
sum=0.0
while [ $n -lt $cnt ]
do 
    load=`uptime|awk '{print $10+0}'`
    echo "load: $load" >> collectLoad.txt
    let n+=1
    sum=$(echo "$load+$sum"|bc)
    echo $sum
    sleep $interval
done

end=$(printf "%.2f" `echo "scale=2; $sum / $cnt"|bc`)
echo "END: $end" >> collectLoad.txt