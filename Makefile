.PHONY:centerTest dcssTest analyse fmt testCompose

Config=./config.yaml
centerTest:
	go run . -c $(Config) --Center >  ./componets.log
	@make analyse 
dcssTest:
	go run . -c $(Config) --Dcss >  ./componets.log
	@make analyse

shareTest:
	go run .   -c $(Config) --ShareState > ./componets.log
	@make analyse


TargetFolder= ./target/$(shell date '+%m_%d_%H_%M_%S')
analyse:
	@mkdir -p $(TargetFolder)
	@cp ./config.log $(TargetFolder)
	go run ./analyse -logFile ./tasks_event.log  -verbose -outputDir $(TargetFolder)
	cp ./draw.py $(TargetFolder)
	cd $(TargetFolder) && python3 draw.py

ComposeFolder = ./test_compose
testCompose:
	bash ./test_compose.sh $(ComposeFolder)

fmt:
	gofmt -l -w .
	golangci-lint run
