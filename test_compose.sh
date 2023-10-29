for config in $(ls ./test_compose/*.yaml);do
    echo $config
    cp $config ./config.yaml
    make centerTest
    make shareTest
    make dcssTest
done

