# DIR is the folder path of test compose
DIR=$1
echo $DIR
rm -rf  $DIR/target/
mkdir -p $DIR/target/

# Clusters is the clusters that will be tested
Clusters=""
ls $DIR/*.run  >/dev/null 2>&1
if [ $? -eq 0 ];then
    for clusterf in $(ls $DIR/*.run);do
        temp=$(basename $clusterf)
        # append cluster 
        Clusters="$Clusters ${temp/%".run"}"
    done
else
    # test all cluster by default
    Clusters="center share dcss"
fi

basename_no_extension(){
    base=$(basename $1)
    echo ${base/%"$2"}
}


for config in $(ls $DIR/*.yaml);do
    echo $config
    configname=$(basename_no_extension $config '.yaml')
    echo "start test from $config for $Clusters"
    for c in $Clusters;do
        echo "start test from $config for $c"
        make ${c}Test Config=$config TargetFolder="$1/target/${configname}_${c}" > /tmp/test_compose.log 2>&1
        if [ $? -eq 0 ];then echo "successd";else cat /tmp/test_compose.log;exit  1 ;fi
    done
done

mkdir -p $DIR/target/all
cp $DIR/*.yaml $DIR/target/all
#for config in $(ls $DIR/*.yaml);do
#    echo $config
#    configname=$(basename_no_extension $config '.yaml')
#    for c in $Clusters;do
#        cp $DIR/target/${configname}_${c}/cluster_status.png $DIR/target/all/${configname}_${c}.png
#    done
#done

