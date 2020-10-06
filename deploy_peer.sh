CGO_ENABLED=0 go build -o ./dist/launcher ./cmd/launcher
docker build -t "dviejo/fabric-peer:amd64-2.2.0" ./dist
docker push 'dviejo/fabric-peer:amd64-2.2.0'