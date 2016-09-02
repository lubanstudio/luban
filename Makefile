build-dev:
	go install -v
	mv $(GOPATH)/bin/luban .

clean:
	go clean -i ./...
