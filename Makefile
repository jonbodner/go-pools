default: build

.PHONY: prof
prof:
	rm -rf prof.fixed prof.norm
	go build -a -o go-pools ./prof/main.go
	./go-pools -profile.func norm
	./go-pools -profile.func fixed
	./go-pools -profile.func fixed2
	echo "prof.norm" > mem_pprof.txt
	go tool pprof -text ./go-pools prof.norm/mem.pprof >> mem_pprof.txt

	echo "" >> mem_pprof.txt
	echo "prof.fixed" >> mem_pprof.txt
	go tool pprof -text ./go-pools prof.fixed/mem.pprof >> mem_pprof.txt
	echo "" >> mem_pprof.txt
	echo "prof.fixed2" >> mem_pprof.txt
	go tool pprof -text ./go-pools prof.fixed2/mem.pprof >> mem_pprof.txt


	echo "prof.norm" > block_pprof.txt
	go tool pprof -text ./go-pools prof.norm/block.pprof >> block_pprof.txt

	echo "" >> block_pprof.txt
	echo "prof.fixed" >> block_pprof.txt
	go tool pprof -text ./go-pools prof.fixed/block.pprof >> block_pprof.txt

	echo "" >> block_pprof.txt
	echo "prof.fixed2" >> block_pprof.txt
	go tool pprof -text ./go-pools prof.fixed2/block.pprof >> block_pprof.txt


	echo "prof.norm" > cpu_pprof.txt
	go tool pprof -text ./go-pools prof.norm/cpu.pprof >> cpu_pprof.txt

	echo "" >> cpu_pprof.txt
	echo "prof.fixed" >> cpu_pprof.txt
	go tool pprof -text ./go-pools prof.fixed/cpu.pprof >> cpu_pprof.txt

	echo "" >> cpu_pprof.txt
	echo "prof.fixed2" >> cpu_pprof.txt
	go tool pprof -text ./go-pools prof.fixed2/cpu.pprof >> cpu_pprof.txt
	
	


.PHONY: deps
deps:
	glide up

.PHONY: test
test: deps
	go tool vet -v -all ./buffers
	go test -covermode=count -v `glide nv`
