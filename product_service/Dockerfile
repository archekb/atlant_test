# Application builder 
FROM debian:10 as builder

WORKDIR /root/product_service

RUN apt update \
    && apt install --no-install-recommends --no-install-suggests -y ca-certificates curl git tar protobuf-compiler \
    && rm -rf /var/cache/apk/* /var/lib/apt/lists/*

RUN curl https://dl.google.com/go/go1.15.2.linux-amd64.tar.gz -o go.tar.gz \
	&& tar -C /usr/local -xzf go.tar.gz \
	&& rm go.tar.gz 

ENV PATH=$PATH:/usr/local/go/bin:/root/go/bin

COPY . .

# v1
RUN go get -u github.com/golang/protobuf/protoc-gen-go \
    && go get -u google.golang.org/grpc 

RUN protoc -I proto proto/product.proto --go_out=plugins=grpc:proto/ \
    && cd server \
    && go build

# v2
# RUN go get -u google.golang.org/protobuf/cmd/protoc-gen-go \
#     && go install google.golang.org/protobuf/cmd/protoc-gen-go \
#     && go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc \
#     && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

# RUN protoc proto/product.proto --go-grpc_out=proto/ \ 
#     && cd server \
#     && go build
    
# Final container
FROM debian:10

WORKDIR /product_service
COPY --from=builder /root/product_service/server/server .

ENTRYPOINT ./server