.PHONY: build clean test test-dprint vendor

WASM_FILE = plugin.wasm

build:
	tinygo build -o $(WASM_FILE).tmp -target=wasm-unknown -scheduler=none -no-debug -opt=2 .
	wasm2wat $(WASM_FILE).tmp -o $(WASM_FILE).wat
	sed -i 's/(export "_initialize" (func $$_initialize))/(export "_initialize" (func $$_initialize))\n  (start $$_initialize)/' $(WASM_FILE).wat
	wat2wasm $(WASM_FILE).wat -o $(WASM_FILE)
	rm -f $(WASM_FILE).tmp $(WASM_FILE).wat

clean:
	rm -f $(WASM_FILE)

test: build
	dprint clear-cache
	cd testdata/basic && cp input.go.txt test.go && dprint fmt --log-level=debug --incremental=false && diff -u expected.go test.go && rm test.go

vendor:
	go mod vendor
	@# Comment out goroutine calls from go-cmp (required for -scheduler=none)
	sed -i 's/go detectRaces/\/\/ go detectRaces/' vendor/github.com/google/go-cmp/cmp/compare.go
