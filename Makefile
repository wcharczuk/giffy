NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
BLUE_COLOR=\033[94;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

all: test

run:
	@echo "$(OK_COLOR)==> Running$(NO_COLOR)"
	@genv -file="./_config/config.json" go run main.go

test:
	@echo "$(OK_COLOR)==> Tests$(NO_COLOR)"
	@genv -file="./_config/config.json" go test -timeout 5s "./server/..."
	@echo "$(OK_COLOR)==> Tests Done!$(NO_COLOR)"

db:
	@echo "$(OK_COLOR)==> Initializing Database$(NO_COLOR)"
	@genv -file=_config/config.json go run ./database/initialize.go
	@echo "$(OK_COLOR)==> Initializing Database Done!$(NO_COLOR)"

list-packages:
	@go list ./... | grep -v /vendor/

migrate:
	@echo "$(OK_COLOR)==> Migrating Database$(NO_COLOR)"
	@genv -file=_config/config.json go run ./database/migrate.go
	@echo "$(OK_COLOR)==> Migrating Database Done!$(NO_COLOR)"