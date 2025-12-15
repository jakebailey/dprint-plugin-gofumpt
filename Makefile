.PHONY: build clean test test-dprint vendor licenses docker-build

WASM_FILE = plugin.wasm
DOCKER_IMAGE = dprint-plugin-gofumpt-builder

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

LICENSES_FILE = licenses/LICENSES

licenses:
	@echo "Generating combined license file..."
	@mkdir -p licenses
	@cp LICENSE $(LICENSES_FILE)
	@echo "" >> $(LICENSES_FILE)
	@echo "================================================================================" >> $(LICENSES_FILE)
	@echo "THIRD PARTY LICENSES" >> $(LICENSES_FILE)
	@echo "================================================================================" >> $(LICENSES_FILE)
	@GOFLAGS="-tags=tinygo -mod=mod" go run github.com/google/go-licenses/v2@latest report . --ignore=github.com/jakebailey/dprint-plugin-gofumpt --template=licenses/licenses.tpl 2>/dev/null >> $(LICENSES_FILE)
	@echo "Generated $(LICENSES_FILE)"

docker-build:
	docker build -t $(DOCKER_IMAGE) .
	docker run --rm -v "$$(pwd)":/workspace $(DOCKER_IMAGE)
