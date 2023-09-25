test:
	mkdir -p target
	go run .
	rm -f target/*.log && rm -f target/*.png 
	analyse 
	nautilus ./target


fmt:
	gofmt -l -w .
