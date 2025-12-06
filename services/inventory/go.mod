module github.com/t4RG3T21/GoBigTech/services/inventory

go 1.25.1

require (
	github.com/stretchr/testify v1.11.1
	github.com/t4RG3T21/GoBigTech/platform v0.0.0
	github.com/t4RG3T21/GoBigTech/services/iam v0.0.0-00010101000000-000000000000
	go.mongodb.org/mongo-driver v1.17.6
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/t4RG3T21/GoBigTech/platform => ../../platform

replace github.com/t4RG3T21/GoBigTech/services/iam => ../iam
