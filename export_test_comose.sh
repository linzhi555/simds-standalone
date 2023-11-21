#!/bin/bash
outpath=./target/test_compose
rm -r $outpath
mkdir -p $outpath
for i in ./test_compose/* ;do 
    cd $i && python3 all.py ; cd -
    cp -r $i/target/all $outpath/$(basename $i)_all
done
