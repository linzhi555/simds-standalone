centerTest:
	go run .
	make analyse
dcssTest:
	go run . -dcss
	make analyse

Folder= target/$(shell date '+%m_%d_%H_%M_%S')
analyse:
	mkdir -p $(Folder)
	cp ./config.yaml $(Folder)
	analyse -outputDir $(Folder)



fmt:
	gofmt -l -w .
