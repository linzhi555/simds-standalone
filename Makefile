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
	nautilus ./target


fmt:
	gofmt -l -w .
