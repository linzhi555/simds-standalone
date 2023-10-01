centerTest:
	go run .
	make analyse
dcssTest:
	go run . -dcss
	make analyse
analyse:
	mkdir -p target
	rm -f target/*.log && rm -f target/*.png 
	analyse 


fmt:
	gofmt -l -w .
