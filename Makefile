NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
BLUE_COLOR=\033[94;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

all: test

run:
	@echo "$(OK_COLOR)==> Running$(NO_COLOR)"
	@genv -f="./_config/config.json" go run main.go

test:
	@echo "$(OK_COLOR)==> Tests$(NO_COLOR)"
	@bash ./_util/test.bash --root="./server/" --package=$(package) --short --filter=${filter} ${flags}
	@echo "$(OK_COLOR)==> Tests Done!$(NO_COLOR)"

cover:
	@echo "$(OK_COLOR)==> Coverage$(NO_COLOR)"
	@sh ./_util/coverage.sh --root="./server/"
	@echo "$(OK_COLOR)==> Coverage Done!$(NO_COLOR)"

db:
	@echo "$(OK_COLOR)==> Initializing Database with configuration from config.json file$(NO_COLOR)"
	@genv -f="./_config/config.json" sh ./_db/init.sh
	@echo "$(OK_COLOR)==> Initializing Database Done!$(NO_COLOR)"

db-dev:
	@echo "$(OK_COLOR)==> Initializing Database with configuration from config.json file$(NO_COLOR)"
	@GIFFY_APP=giffy-dev-db GIFFY_HOST=45.33.5.126 sh ./_db/init.sh
	@echo "$(OK_COLOR)==> Initializing Database Done!$(NO_COLOR)"

migrate:
	@echo "$(OK_COLOR)==> Migrating Database with configuration from config.json file$(NO_COLOR)"
	@genv -f="./_config/config.json" sh ./_db/migrate.sh
	@echo "$(OK_COLOR)==> Migrating Database Done!$(NO_COLOR)"

migrate-dev:
	@echo "$(OK_COLOR)==> Migrating Database with configuration from config.json file$(NO_COLOR)"
	@GIFFY_APP=giffy-dev-db GIFFY_HOST=45.33.5.126 sh ./_db/migrate.sh
	@echo "$(OK_COLOR)==> Migrating Database Done!$(NO_COLOR)"

migrate-prod:
	@echo "$(OK_COLOR)==> Migrating Database with configuration from config.json file$(NO_COLOR)"
	@GIFFY_APP=giffy-db GIFFY_HOST=45.33.5.126 sh ./_db/migrate.sh
	@echo "$(OK_COLOR)==> Migrating Database Done!$(NO_COLOR)"