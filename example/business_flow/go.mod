module business_flow

go 1.24

toolchain go1.24.3

require (
	github.com/google/uuid v1.6.0
	github.com/kslamph/teleflow v0.0.0
	golang.org/x/image v0.28.0
)

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1 // indirect
	golang.org/x/text v0.26.0 // indirect
)

replace github.com/kslamph/teleflow => ../..
