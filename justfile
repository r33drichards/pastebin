tidy:
    go mod tidy

build: tidy
     docker build -t pbin:latest .

run: build
    docker run --env-file .env -p 8000:8000 pbin:latest