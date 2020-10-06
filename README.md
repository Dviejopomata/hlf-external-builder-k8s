Build

```bash
export CGO_ENABLED=0 # this is needed to execute it in the alpine docker container
go build -o ./dist/launcher ./cmd/launcher
docker build -t dviejo/fabric-peer:amd64-2.2.0 ./dist
docker push dviejo/fabric-peer:amd64-2.2.0

#go build -o ./dist/build ./cmd/build
#go build -o ./dist/detect ./cmd/detect
#go build -o ./dist/release ./cmd/release
#go build -o ./dist/run ./cmd/run


docker run --rm -it dviejo/fabric-peer:amd64-2.2.0 sh

docker build -t dviejo/fs-peer:amd64-2.2.0 ./
docker push dviejo/fs-peer:amd64-2.2.0

docker build -t dviejo/fabric-init:amd64-2.2.0 -f init.Dockerfile ./
docker push dviejo/fabric-init:amd64-2.2.0


kubectl scale deployment peer-0-org1-mdcamq --replicas=0
#kubectl get deployment | grep peer-0-org1-mdcamq  
kubectl get pod | grep peer-0-org1-mdcamq | wc -l
watch -n1 'kubectl get pod | grep peer-0-org1-mdcamq | wc -l'

kubectl wait --for=replicas=available --timeout=600s deployment/peer-0-org1-mdcamq -n default

kubectl scale deployment peer-0-org1-mdcamq --replicas=1

export CHAINCODE_PATH=/disco-grande/euipo/hyperledger-git-repos/fabric-samples/chaincode/fabcar/external
docker run --name ccenv -it -v $(pwd)/build.sh:/chaincode/build.sh -v $CHAINCODE_PATH:/chaincode/input/src --rm hyperledger/fabric-ccenv:2.2 sh

docker rm ccenv -f

```

curl -s --upload-file ./Dockerfile 'http://localhost:8080/Dockerfile'

Detect ->

curl -X POST -s --upload-file  /tmp/ec4235d137.026685681/chaincode-source.tar.gz http://172.18.0.19:8080/58fefcd24f/chaincode-source.tar.gz