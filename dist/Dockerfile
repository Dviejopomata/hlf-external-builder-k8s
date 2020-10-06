FROM hyperledger/fabric-peer:amd64-2.2.0
RUN mkdir -p /builders/golang/bin
COPY launcher /builders/golang/bin/externalcc
RUN ln -s /builders/golang/bin/externalcc /builders/golang/bin/detect
RUN ln -s /builders/golang/bin/externalcc /builders/golang/bin/build
RUN ln -s /builders/golang/bin/externalcc /builders/golang/bin/release
RUN ln -s /builders/golang/bin/externalcc /builders/golang/bin/run

COPY k8scc.yaml /builders/golang/bin/k8scc.yaml
