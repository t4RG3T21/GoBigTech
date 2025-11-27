module github.com/t4RG3T21/GoBigTech/services/iam

go 1.25.1

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/pressly/goose/v3 v3.26.0
	github.com/redis/go-redis/v9 v9.7.0
	github.com/stretchr/testify v1.11.1
	github.com/t4RG3T21/GoBigTech/platform v0.0.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.40.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/t4RG3T21/GoBigTech/platform => ../../platform
