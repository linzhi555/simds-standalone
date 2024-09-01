.PHONY:centerTest dcssTest analyse fmt testCompose

Cluster=Center
Config=./config.yaml
timeNow:=$(shell date '+%m_%d_%H_%M_%S')
TargetFolder:=./target/$(timeNow)
RootPath=$(CURDIR)

test: preDeal
	@mkdir -p $(TargetFolder)
	go run ./standalone/main -c $(Config) --OutputDir $(TargetFolder) --Cluster $(Cluster) > $(TargetFolder)/stdout.log
	@make analyse TargetFolder=$(TargetFolder)

debug: preDeal
	@mkdir -p $(TargetFolder)
	go run ./standalone/main -c $(Config) --OutputDir $(TargetFolder) --Cluster $(Cluster) --Debug
	@make analyse TargetFolder=$(TargetFolder)

preDeal:
	if [ -d $(TargetFolder) ];then echo "target folder is not empty";exit 1;fi
	@mkdir -p $(TargetFolder)

analyse:
	go run ./tracing/main  --OutputDir $(TargetFolder) -c $(Config)
	cd $(TargetFolder) && python3 $(RootPath)/test/draw.py

fmt:
	gofmt -l -w .
	golangci-lint run

k8sTest:preDeal
	@mkdir -p $(TargetFolder)
	go run ./simctl -c $(Config) --OutputDir $(TargetFolder) --Cluster $(Cluster) 
	@make analyse TargetFolder=$(TargetFolder)
k8sClean:
	go run ./simctl --CleanMode
