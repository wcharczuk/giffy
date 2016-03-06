NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
BLUE_COLOR=\033[94;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

all: test-bot test-server

run-bot:
	@echo "$(OK_COLOR)==> Running Bot$(NO_COLOR)"
	@genv -f="./_config/config.json" go run ./bot/main.go

run-server:
	@echo "$(OK_COLOR)==> Running Bot$(NO_COLOR)"
	@genv -f="./_config/config.json" go run ./server/main.go

test-bot:
	@echo "$(OK_COLOR)==> Running Server Tests$(NO_COLOR)"
	@util/test.bash --root="./bot" --package=$(package) --short --filter=${filter}
	@echo "$(OK_COLOR)==> Running Server Tests Done!$(NO_COLOR)"

test-server:
	@echo "$(OK_COLOR)==> Running Server Tests$(NO_COLOR)"
	@bash ./_util/test.bash --root="./server" --package=$(package) --short --filter=${filter}
	@echo "$(OK_COLOR)==> Running Server Tests Done!$(NO_COLOR)"

cover-bot:
	@echo "$(OK_COLOR)==> Running Coverage$(NO_COLOR)"
	@sh ./_util/coverage.sh --root="./bot"
	@echo "$(OK_COLOR)==> Running Coverage Done!$(NO_COLOR)"

cover-server:
	@echo "$(OK_COLOR)==> Running Coverage$(NO_COLOR)"
	@sh ./_util/coverage.sh --root="./server"
	@echo "$(OK_COLOR)==> Running Coverage Done!$(NO_COLOR)"

db:
	@echo "$(OK_COLOR)==> Initializing Database with configuration from config.json file$(NO_COLOR)"
	@genv -f="./_config/config.json" sh ./server/_db/init.sh
	@echo "$(OK_COLOR)==> Initializing Database Done!$(NO_COLOR)"

migrate:
	@echo "$(OK_COLOR)==> Migrating Database with configuration from config.json file$(NO_COLOR)"
	@genv -f="./_config/config.json" sh ./server/_db/migrate.sh
	@echo "$(OK_COLOR)==> Migrating Database Done!$(NO_COLOR)"