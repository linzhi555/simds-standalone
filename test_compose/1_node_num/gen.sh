last=1000
for i in $(seq 2 8) ;do
    filename=./test${i}.yaml
    rm -rf $filename
    cp ./test1.yaml $filename
    now=$(expr  $last + $last \* 4 / 10)
    sed -i "s/^NodeNum.*$/NodeNum:       $now/g" $filename
    last=$now
done
