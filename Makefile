.PHONY:centerTest dcssTest analyse fmt

centerTest:
	go run .
	make analyse
dcssTest:
	go run . -dcss
	make analyse

shareTest:
	go run . -share
	make analyse

Folder= target/$(shell date '+%m_%d_%H_%M_%S')

analyse:
	mkdir -p $(Folder)
	cp ./config.yaml $(Folder)
	go run ./analyse -verbose -outputDir $(Folder)

fmt:
	gofmt -l -w .
