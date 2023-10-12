centerTest:
	go run .
	make analyse
dcssTest:
	go run . -dcss
	make analyse

Folder= target/$(shell date '+%m_%d_%H_%M_%S')

.PHONY:analyse
analyse:
	mkdir -p $(Folder)
	cp ./config.yaml $(Folder)
	go run ./analyse -verbose -outputDir $(Folder)

fmt:
	gofmt -l -w .
