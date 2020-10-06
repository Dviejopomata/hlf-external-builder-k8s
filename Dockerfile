FROM golang:1.15.2-alpine3.12

ADD . /go/src/github.com/kungfusoftware/externalbuilder
RUN go install github.com/kungfusoftware/externalbuilder/cmd/fileserver


# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/fileserver

# Document that the service listens on port 8080.
EXPOSE 8080