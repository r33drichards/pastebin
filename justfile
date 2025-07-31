tidy:
    go mod tidy
    gomod2nix 

# Install frontend dependencies
install:
    pnpm install

# Run frontend in development mode
dev-frontend:
    pnpm dev

# Run backend in development mode
dev-backend:
    go run .

# Run both frontend and backend in development
dev:
    #!/usr/bin/env bash
    pnpm dev &
    go run .

# Build frontend
build-frontend:
    pnpm build

# Build everything with Nix
build: tidy 
    nix build .#default

# Build Docker image with Nix
build-docker: tidy
    nix build .#dockerImage
    docker load -i ./result

# Build and run with Docker
run: build-docker
    docker run --env-file .env -p 8000:8000 pbin:latest

# Run linting and formatting
lint:
    go fmt
    npx prettier --parser html --write templates/*
    npx prettier --write README.md
    npx prettier --write "src/**/*.{ts,tsx,js,jsx,css}"

clean:
    rm -rf result
    rm -rf static
    rm -rf node_modules