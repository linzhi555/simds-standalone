for i in $(seq 2 9);do
    filename=./test${i}.yaml
    rm -rf $filename
    cp ./test1.yaml $filename
    sed -i "s/^NodeNum.*$/NodeNum:        $(expr 850 + 150 \* $i \* $i)/g" $filename
done
