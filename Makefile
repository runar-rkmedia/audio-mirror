watch: ./prepare
	${MAKE} -j 4 watch_go watch_go_tests watch_frontend watch_buf
gen: ./api/v1/pods.proto ./buf.yaml ./buf.gen.yaml
	buf lint
	buf generate
	go mod tidy
./frontend/node_modules: ./frontend/bun.lockb ./frontend/package.json
	cd frontend && bun install
	@touch ./frontend/node_modules
./frontend/src/gen/: ./api/v1/pods.proto ./buf.yaml ./buf.gen.yaml
	buf lint
	buf generate
	go mod tidy
deps:
	go mod tidy
prepare: deps dep_bins ./frontend/node_modules
watch_frontend:
	@cd frontend && bun dev
watch_buf:
	fd -e proto | entr -rc make gen
watch_go:
	fd --exclude frontend --exclude gen --exclude .cache --exclude ./db.sqlite3 | entr -rc go run ./cmd/api
watch_go_tests:
	fd --exclude frontend --exclude gen --exclude .cache --exclude ./db.sqlite3 | entr -rc gotestsum --format-hide-empty-pkg --format testdox
test_go:
	gotestsum --format-hide-empty-pkg --format testdox
dep_bins: buf_installed grpcurl_installed protoc_gen_go_installed protoc_gen_connect_go_installed

buf_installed:
	@if ! command -v buf &> /dev/null; then \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi

grpcurl_installed:
	@if ! command -v grpcurl &> /dev/null; then \
		go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest; \
	fi

protoc_gen_go_installed:
	@if ! command -v protoc-gen-go &> /dev/null; then \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi

protoc_gen_connect_go_installed:
	@if ! command -v protoc-gen-connect-go &> /dev/null; then \
		go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest; \
	fi




