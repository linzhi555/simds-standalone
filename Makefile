.PHONY:centerTest dcssTest analyse fmt testCompose

Config=./config.yaml
timeNow:=$(shell date '+%m_%d_%H_%M_%S')
TargetFolder=./target/$(timeNow)
Cluster=Center

test: preDeal
	@mkdir -p $(TargetFolder)
	go run . -c $(Config) --OutputDir $(TargetFolder) --Cluster $(Cluster) >  $(TargetFolder)/components.log
	@make analyse TargetFolder=$(TargetFolder)

preDeal:
	if [ -d $(TargetFolder) ];then echo "target folder is not empty";exit 1;fi
	@mkdir -p $(TargetFolder)



analyse:
	go run ./analyse -taskLog $(TargetFolder)/tasks_event.log -netLog $(TargetFolder)/network_event.log -verbose -outputDir $(TargetFolder)
	cp ./py/draw.py $(TargetFolder)
	cd $(TargetFolder) && grep  'TaskGen : send task to' ./components.log > ./task_speed.log \
	&& grep  'Info network.*sended' ./components.log > ./net.log \
	&& python3 draw.py 
	#rm $(TargetFolder)/components.log

ComposeFolder = ./test_compose
testCompose:
	bash ./test_compose.sh $(ComposeFolder)

fmt:
	gofmt -l -w .
	golangci-lint run
