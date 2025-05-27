# pbin

- https://gitlab.com/reedrichards/pbin

try it out at [p.jjk.is](https://p.jjk.is)

## Quickstart


if you have [just](https://github.com/casey/just) and [docker](https://docs.docker.com/get-docker/) installed, you can
start the project with `just run`. 

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
