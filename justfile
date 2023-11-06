tidy:
    go mod tidy
    gomod2nix 

build: tidy 
    nix build .#dockerImage
    docker load -i ./result

run: build
    docker run --env-file .env -p 8000:8000 pbin:latest

lint:
    go fmt
    npx prettier --parser html --write templates/*
    npx prettier --write README.md

clean:
    rm -rf result