echo $1
rm -rf  $1/target/
for config in $(ls $1/*.yaml);do
    echo $config
    cp $config ./config.yaml
    base=$(basename $config)
    configname=${base/%".yaml"}
    echo "run three test , test : $configname"
    make centerTest Config=$config TargetFolder="$1/target/${configname}_center"
    make shareTest  Config=$config TargetFolder="$1/target/${configname}_share"
    make dcssTest   Config=$config TargetFolder="$1/target/${configname}_dcss"
done
