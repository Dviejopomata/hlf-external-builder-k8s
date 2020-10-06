set -e
if [ -f "/chaincode/input/src/go.mod" ] && [ -d "/chaincode/input/src/vendor" ]; then
    cd /chaincode/input/src
    GO111MODULE=on go build -v -mod=vendor -ldflags "-linkmode external -extldflags '-static'" -o /chaincode/output/chaincode %[2]s
elif [ -f "/chaincode/input/src/go.mod" ]; then
    cd /chaincode/input/src
    GO111MODULE=on go build -v -mod=readonly -ldflags "-linkmode external -extldflags '-static'" -o /chaincode/output/chaincode /chaincode/input/src
elif [ -f "/chaincode/input/src/%[2]s/go.mod" ] && [ -d "/chaincode/input/src/%[2]s/vendor" ]; then
    cd /chaincode/input/src/%[2]s
    GO111MODULE=on go build -v -mod=vendor -ldflags "-linkmode external -extldflags '-static'" -o /chaincode/output/chaincode .
elif [ -f "/chaincode/input/src/%[2]s/go.mod" ]; then
    cd /chaincode/input/src/%[2]s
    GO111MODULE=on go build -v -mod=readonly -ldflags "-linkmode external -extldflags '-static'" -o /chaincode/output/chaincode .
else
    GOPATH=/chaincode/input:$GOPATH go build -v -ldflags "-linkmode external -extldflags '-static'" -o /chaincode/output/chaincode %[2]s
fi
echo Done!