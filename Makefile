.PHONY:centerTest dcssTest analyse fmt testCompose

Config=./config.yaml
timeNow:=$(shell date '+%m_%d_%H_%M_%S')
TargetFolder:=./target/$(timeNow)

Cluster=Center
test: preDeal
	@mkdir -p $(TargetFolder)
	go run ./standalone -c $(Config) --OutputDir $(TargetFolder) --Cluster $(Cluster) > $(TargetFolder)/stdout.log
	@make analyse TargetFolder=$(TargetFolder)

centerTest: 
	make test Cluster=Center

dcssTest:
	make test Cluster=Dcss

shareTest:
	make test Cluster=ShareState

preDeal:
	if [ -d $(TargetFolder) ];then echo "target folder is not empty";exit 1;fi
	@mkdir -p $(TargetFolder)

analyse:
	go run ./analyse  --OutputDir $(TargetFolder) -c $(Config)
	cp ./py/draw.py $(TargetFolder)
	cd $(TargetFolder) && python3 draw.py 
	#rm $(TargetFolder)/components.log

ComposeFolder = ./test_compose
testCompose:
	bash ./test_compose.sh $(ComposeFolder)

fmt:
	gofmt -l -w .
	golangci-lint run
