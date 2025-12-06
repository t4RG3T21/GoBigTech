module github.com/t4RG3T21/GoBigTech/services/order

go 1.25.1

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/jackc/pgx/v5 v5.7.6
	github.com/oapi-codegen/runtime v1.1.2
	github.com/pressly/goose/v3 v3.26.0
	github.com/prometheus/client_golang v1.23.2
	github.com/segmentio/kafka-go v0.4.47
	github.com/stretchr/testify v1.11.1
	github.com/t4RG3T21/GoBigTech/platform v0.0.0
	github.com/t4RG3T21/GoBigTech/services/inventory v0.0.0-00010101000000-000000000000
	github.com/t4RG3T21/GoBigTech/services/payment v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.77.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/t4RG3T21/GoBigTech/platform => ../../platform

replace github.com/t4RG3T21/GoBigTech/services/inventory => ../inventory

replace github.com/t4RG3T21/GoBigTech/services/payment => ../payment
