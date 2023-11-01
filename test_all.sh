#!/bin/bash
for i in ./test_compose/* ;do 
    if [ -d $i ];then
        echo run test $(basename $i) of [ $(ls ./test_compose) ]
        make testCompose ComposeFolder=$i  >/dev/null 2>&1
    fi
done
