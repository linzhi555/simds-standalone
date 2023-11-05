last=100
for i in $(seq 2 5) ;do
    filename=./test${i}.yaml
    rm -rf $filename
    cp ./test1.yaml $filename
    now=$(expr  $last + $last \* 4 / 10)
    sed -i "s/^StateUpdatePeriod:.*$/StateUpdatePeriod:       $now/g" $filename
    last=$now
done
