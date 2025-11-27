module github.com/t4RG3T21/GoBigTech/services/notification

go 1.25.1

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/segmentio/kafka-go v0.4.47
	github.com/t4RG3T21/GoBigTech/platform v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.42.0 // indirect
)

replace github.com/t4RG3T21/GoBigTech/platform => ../../platform
