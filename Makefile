.PHONY:centerTest dcssTest analyse fmt testCompose

Config=./config.yaml
centerTest:
	go run . -c $(Config) --Center >  ./components.log
	@make analyse 
dcssTest:
	go run . -c $(Config) --Dcss >  ./components.log
	@make analyse

shareTest:
	go run .   -c $(Config) --ShareState > ./components.log
	@make analyse


timeNow := $(shell date '+%m_%d_%H_%M_%S')
TargetFolder= ./target/$(timeNow)

analyse:
	@mkdir -p $(TargetFolder)
	@cp ./config.log $(TargetFolder)
	go run ./analyse -logFile ./tasks_event.log  -verbose -outputDir $(TargetFolder)
	cp ./py/draw.py $(TargetFolder)
	cp ./components.log $(TargetFolder)
	cd $(TargetFolder) && python3 draw.py 
	rm $(TargetFolder)/components.log

ComposeFolder = ./test_compose
testCompose:
	bash ./test_compose.sh $(ComposeFolder)

fmt:
	gofmt -l -w .
	golangci-lint run
