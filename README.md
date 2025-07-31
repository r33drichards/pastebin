# pbin

- https://gitlab.com/reedrichards/pbin

try it out at [p.jjk.is](https://p.jjk.is)

## Quickstart

if you have [just](https://github.com/casey/just) and [docker](https://docs.docker.com/get-docker/) installed, you can
start the project with `just run`. 

## Development with Protocol Buffers

This project uses Protocol Buffers for API communication between the Go backend and React frontend.

### Prerequisites

Enter the development shell to get all required tools:
```bash
nix develop
```

This provides:
- `protoc` (Protocol Buffer compiler)
- `protoc-gen-go` (Go protobuf plugin)
- `protoc-gen-go-grpc` (Go gRPC plugin)
- `nodejs` and `pnpm` (for frontend development)

### Generate Go Code

Generate Go server and client code from protobuf definitions:
```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/pastebin.proto
```

### TypeScript Development

For TypeScript development, the project uses manually defined interfaces based on the protobuf definitions. The interfaces are located in `src/types/pastebin.ts` and provide type-safe access to the API.

The API client is implemented in `src/services/pastebinApi.ts` and provides a simple HTTP-based client for the Pastebin service.

## build and run binary 
```
$ nix build
$ ./result/bin/pbin
{"level":"info","ts":1745557900.438304,"caller":"pbin/main.go:453","msg":"starting_server","port":"8000"}
```


### build docker image 

```
 nix build .#dockerImage
 ```

```
docker load -i ./result
```

```
docker images
```

```
docker run --env-file .env -p 8000:8000 pbin:latest
# local port 3000
docker run --env-file .env -p 3000:8000 pbin:latest
```
