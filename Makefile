.PHONY:centerTest dcssTest analyse fmt

centerTest:
	go run . >  ./componets.log
	@make analyse 
dcssTest:
	go run . -dcss >  ./componets.log
	@make analyse

shareTest:
	go run . -share > ./componets.log
	@make analyse

Folder= target/$(shell date '+%m_%d_%H_%M_%S')

analyse:
	@mkdir -p $(Folder)
	@cp ./config.yaml $(Folder)
	go run ./analyse -logFile ./tasks_event.log  -verbose -outputDir $(Folder)

fmt:
	gofmt -l -w .
