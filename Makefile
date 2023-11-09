.PHONY:centerTest dcssTest analyse fmt testCompose

Config=./config.yaml
timeNow:=$(shell date '+%m_%d_%H_%M_%S')
TargetFolder=./target/$(timeNow)


centerTest: preDeal
	@mkdir -p $(TargetFolder)
	go run . -c $(Config) --OutputDir $(TargetFolder) --Cluster Center >  $(TargetFolder)/components.log
	@make analyse TargetFolder=$(TargetFolder)
dcssTest: preDeal
	@mkdir -p $(TargetFolder)
	go run . -c $(Config) --OutputDir $(TargetFolder) --Cluster Dcss >  $(TargetFolder)/components.log
	@make analyse TargetFolder=$(TargetFolder)

shareTest: preDeal
	@mkdir -p $(TargetFolder)
	go run .   -c $(Config) --OutputDir $(TargetFolder) --Cluster ShareState > $(TargetFolder)/components.log
	@make analyse TargetFolder=$(TargetFolder)

preDeal:
	if [ -d $(TargetFolder) ];then echo "target folder is not empty";exit 1;fi
	@mkdir -p $(TargetFolder)



analyse:
	go run ./analyse -logFile $(TargetFolder)/tasks_event.log  -verbose -outputDir $(TargetFolder)
	cp ./py/draw.py $(TargetFolder)
	cd $(TargetFolder) && python3 draw.py 
	#rm $(TargetFolder)/components.log

ComposeFolder = ./test_compose
testCompose:
	bash ./test_compose.sh $(ComposeFolder)

fmt:
	gofmt -l -w .
	golangci-lint run
