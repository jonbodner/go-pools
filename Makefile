default: build


.PHONY: deps
deps:
	glide cc
	glide up

.PHONY: test
test: 
	go tool vet -v -all ./buffers
	go test -covermode=count -v `glide nv`

.PHONY: cover
cover:
	go test -covermode=count  -coverprofile=buffers.out  -v ./buffers/...
	go test -covermode=count  -coverprofile=workers.out  -v ./workers/...

.PHONY: build
build: test
	
