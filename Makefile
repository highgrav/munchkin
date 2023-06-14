clean:
	- rm ./walcompactor
	- rm ./munchkin
	- rm ./grpcwalclient
all:
	protoc api/v1/*.proto --go_out=. --go-grpc_out=.  --go_opt=paths=source_relative  --go-grpc_opt=paths=source_relative --proto_path=.
	go build -o munchkin ./cmd/main
	go build -o walcompactor ./cmd/walcompactor
	go build -o grpcwalclient ./cmd/grpcwalclient
munchkin:
	 go build -o munchkin ./cmd/main
walcompactor:
	go build -o walcompactor ./cmd/walcompactor
grpcwalclient:
	go build -o grpcwalclient ./cmd/grpcwalclient
protos:
	protoc api/v1/*.proto --go_out=. --go_opt=paths=source_relative --proto_path=. --go-grpc_out=.  --go-grpc_opt=paths=source_relative
test:
	go test -v -race ./...
certs:
	- rm ./certs/*.pem
	- rm ./certs/*.csr
	cfssl gencert -initca ./test/ca-csr.json | cfssljson -bare ca
	cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=./test/ca-config.json -profile=server ./test/server-csr.json | cfssljson -bare server
	mv *.pem ./certs
	mv *.csr ./certs