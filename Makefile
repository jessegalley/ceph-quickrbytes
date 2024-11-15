export CGO_ENABLED=0

build:
	go build -o bin/ceph-quickrbytes

run: build
	./bin/ceph-quickrbytes
