echo $1
rm -rf  $1/target/
for config in $(ls $1/*.yaml);do
    echo $config
    base=$(basename $config)
    configname=${base/%".yaml"}
    echo "run three test , test : $configname"

    # if dcss_only.txt is in this folder then just run dcss test
    if [ -f $1/dcss_only.txt ];then
        echo " just run dcss test only"
        make dcssTest   Config=$config TargetFolder="$1/target/${configname}_dcss"
        continue
    fi

    make centerTest Config=$config TargetFolder="$1/target/${configname}_center"
    make shareTest  Config=$config TargetFolder="$1/target/${configname}_share"
    make dcssTest   Config=$config TargetFolder="$1/target/${configname}_dcss"
done
