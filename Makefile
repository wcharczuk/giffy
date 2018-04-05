CONFIG_PATH ?= _config/local.yml
export CONFIG_PATH

NAMESPACE ?= giffy
export NAMESPACE

CURRENT_REF := $(shell git log --pretty=format:'%h' -n 1)
export CURRENT_REF

all: migrate test

new-install: init-db db

run:
	@echo "==> Running"
	@go run main.go

test:
	@echo "==> Tests"
	@go test -timeout 5s "./server/..."
	@echo "==> Tests Done!"

db-init: init-db

init-db:
	@echo "==> Fist Time Database Setup"
	@createdb giffy
	@echo "==> Fist Time Database Setup Done!"

db:
	@echo "==> Initializing Database"
	@go run ./database/main.go init
	@echo "==> Initializing Database Done!"

migrate:
	@echo "==> Migrating Database"
	@go run ./database/main.go migrate
	@echo "==> Migrating Database Done!"

list-packages:
	@go list ./... | grep -v /vendor/

build:
	@docker build --build-arg CURRENT_REF=$(CURRENT_REF) -t giffy:latest -t wcharczuk/giffy:latest -t wcharczuk/giffy:$(CURRENT_REF) -f Dockerfile .

push-image: push

push: build
	@docker push wcharczuk/giffy:latest
	@docker push wcharczuk/giffy:$(CURRENT_REF)

kube-init:
	@kubectl create namespace $(NAMESPACE)

provision:
	@kubectl --namespace=$(NAMESPACE) create secret generic web-config --from-file=config.yml=$(CONFIG_PATH)
	@kubectl --namespace=$(NAMESPACE) create -f _kube/deployment.yml
	@kubectl --namespace=$(NAMESPACE) create -f _kube/service.yml

deprecate: 
	@kubectl --namespace=$(NAMESPACE) delete --ignore-not-found --grace-period=0 -f _kube/service.yml
	@kubectl --namespace=$(NAMESPACE) delete --ignore-not-found --grace-period=0 -f _kube/deployment.yml
	@kubectl --namespace=$(NAMESPACE) delete secret web-config --ignore-not-found

recreate-deployment:
	@kubectl --namespace=$(NAMESPACE) delete --ignore-not-found --grace-period=0 -f _kube/deployment.yml
	@kubectl --namespace=$(NAMESPACE) create -f _kube/deployment.yml

recreate-config:
	@kubectl --namespace=$(NAMESPACE) delete secret web-config --ignore-not-found
	@kubectl --namespace=$(NAMESPACE) create secret generic web-config --from-file=config.yml=$(CONFIG_PATH)

deploy:
	@kubectl --namespace=$(NAMESPACE) set image deployment/web-server web-server=docker.io/wcharczuk/giffy:$(CURRENT_REF)
