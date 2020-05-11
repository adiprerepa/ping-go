clean:
	@rm -rf bin/
	@mkdir -p bin/

build: clean
	go build -o bin/ping src/ping.go;


