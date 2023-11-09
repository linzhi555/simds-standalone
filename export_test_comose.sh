outpath=./target/test_compose
rm -r $outpath
mkdir -p $outpath
for i in ./test_compose/* ;do 
    cp -r $i/target/all $outpath/$(basename $i)_all
done
