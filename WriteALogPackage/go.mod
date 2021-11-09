module github.com/manofthelionarmy/prolog

go 1.16

require (
	github.com/casbin/casbin v1.9.1
	github.com/golang/protobuf v1.5.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/stretchr/testify v1.7.0
	github.com/tysonmote/gommap v0.0.0-20210506040252-ef38c88b18e1
	google.golang.org/genproto v0.0.0-20200423170343-7949de9c1215
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.27.1
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 => github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
