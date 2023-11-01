echo $1
rm -rf  $1/target/
for config in $(ls $1/*.yaml);do
    echo $config
    base=$(basename $config)
    configname=${base/%".yaml"}
    cluster=""
    ls $1/*.run 2>/dev/null
    if [ $? -eq 0 ];then
        for clusterf in $(ls $1/*.run);do
            temp=$(basename $clusterf)
            cluster="$cluster ${temp/%".run"}"
        done
    else
        cluster="center share dcss"
    fi

    echo "start test from $config for $cluster"
    for c in $cluster;do

        echo "start test from $config for $c"
        make ${c}Test Config=$config TargetFolder="$1/target/${configname}_${c}" > /tmp/test_compose.log 2>&1
        if [ $? -eq 0 ];then echo "successd";else cat /tmp/test_compos.log;exit  1 ;fi
    done
done
