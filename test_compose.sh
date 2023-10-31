echo $1
rm -rf  $1/target/
for config in $(ls $1/*.yaml);do
    echo $config
    base=$(basename $config)
    configname=${base/%".yaml"}
    cluster=""
    ls $1/*.run
    if [ $? -eq 0 ];then
        for clusterf in $(ls $1/*.run);do
            temp=$(basename $clusterf)
            cluster="$cluster ${temp/%".run"}"
        done
    else
        cluster="center share dcss"
    fi

    echo "start test for $cluster"
    for c in $cluster;do
        make ${c}Test Config=$config TargetFolder="$1/target/${configname}_${c}"
    done
done
