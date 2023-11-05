last=50000
for i in $(seq 2 5) ;do
    filename=./test${i}.yaml
    rm -rf $filename
    cp ./test1.yaml $filename
    now=$(expr  $last + $last \* 1 / 10)
    sed -i "s/^SchedulerPerformance:.*$/SchedulerPerformance: $now/g" $filename
    last=$now
done
