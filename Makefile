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
deps: ./frontend/node_modules
	go mod tidy
watch_frontend:
	@cd frontend && bun dev
watch_buf:
	fd -e proto | entr -rc make gen
watch_go:
	fd --exclude frontend --exclude gen --exclude cache | entr -rc go run ./cmd/api
watch:
	${MAKE} -j 3 watch_go watch_frontend watch_buf
dep_bins:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest




