NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
BLUE_COLOR=\033[94;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

CONFIG_PATH ?= _config/local.yml
export CONFIG_PATH

NAMESPACE ?= giffy
export NAMESPACE

all: test

new-install: init-db db

run:
	@echo "$(OK_COLOR)==> Running$(NO_COLOR)"
	@go run main.go

test:
	@echo "$(OK_COLOR)==> Tests$(NO_COLOR)"
	@go test -timeout 5s "./server/..."
	@echo "$(OK_COLOR)==> Tests Done!$(NO_COLOR)"

db-init: init-db

init-db:
	@echo "$(OK_COLOR)==> Fist Time Database Setup$(NO_COLOR)"
	@sh ./_config/init_db.sh
	@echo "$(OK_COLOR)==> Fist Time Database Setup Done!$(NO_COLOR)"

db:
	@echo "$(OK_COLOR)==> Initializing Database$(NO_COLOR)"
	@go run ./database/initialize.go
	@echo "$(OK_COLOR)==> Initializing Database Done!$(NO_COLOR)"

list-packages:
	@go list ./... | grep -v /vendor/

migrate:
	@echo "$(OK_COLOR)==> Migrating Database$(NO_COLOR)"
	@go run ./database/migrate.go
	@echo "$(OK_COLOR)==> Migrating Database Done!$(NO_COLOR)"

build:
	@docker build -t giffy:latest -t wcharczuk/giffy:latest -f Dockerfile .

push: build
	@docker push wcharczuk/giffy:latest

kube-init:
	@kubectl create namespace $(NAMESPACE)

provision: #push
	@kubectl --namespace=$(NAMESPACE) create secret generic web-config --from-file=config.yml=$(CONFIG_PATH)
	@kubectl --namespace=$(NAMESPACE) create -f _kube/deployment.yml
	@kubectl --namespace=$(NAMESPACE) create -f _kube/service.yml

deprecate: 
	@kubectl --namespace=$(NAMESPACE) delete --ignore-not-found --grace-period=0 -f _kube/service.yml
	@kubectl --namespace=$(NAMESPACE) delete --ignore-not-found --grace-period=0 -f _kube/deployment.yml
	@kubectl --namespace=$(NAMESPACE) delete secret web-config --ignore-not-found
