SHELL=/bin/bash
.DEFAULT_GOAL=setup
CURRENTDIR=$(shell dirname `pwd`)
TIMESTAMP := $(shell date +"%Y%m%d%H%M%S")

# Setup application
setup: go.mod
	@echo "`tput bold`#### Verifying configuration files and server certificates ####`tput sgr0`"
	@test -f .env || cp .env.example .env
	@echo "## Configuration files are now ready to use ##"

	@sleep 2

	@echo "`tput bold`#### Installing dependencies to your project ####`tput sgr0`"
	go mod tidy

	go get -u golang.org/x/lint/golint
	go install golang.org/x/lint/golint
	go get -u github.com/mgechev/revive@latest
	go install github.com/mgechev/revive@latest

	go install golang.org/x/tools/gopls@latest

	go install github.com/securego/gosec/v2/cmd/gosec@latest

	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/automation-co/husky@latest

	go install github.com/daixiang0/gci@latest

	test -f .husky/hooks/pre-commit || husky init
	test -f .husky/hooks/pre-commit && husky add pre-commit "go fmt . && golangci-lint run  --go=1.24 --enable-all --disable tagliatelle,wsl,godox,lll,gochecknoglobals,exhaustruct,wrapcheck,depguard,ireturn,misspell && fieldalignment ./... && go test -race -count=1 ./... && gosec ./..." 

	@echo "## All dependencies installed successfully ##"
	@sleep 2

	@echo ""
	@echo "`tput bold``tput setaf 1`## Verify .env and fill it according to your params ##`tput sgr0`"
	@echo ""

# Run local server
run: .env
	set -a; source .env; set +a; go run .

# Lint application
lint:
	@printf "\e[34m Running golangci-lint. ## \n"

	golangci-lint run $(file) --go=1.24 --enable-all --disable tagliatelle,wsl,godox,lll,gochecknoglobals,exhaustruct,wrapcheck,depguard,ireturn,misspell --timeout=5m

	@printf "\e[34m No issues found with golangci-lint. ## \n"
	@sleep 2

	fieldalignment ./...

	@printf "\e[34m No issues found ## \n"

	@printf "\e[34m All error checks passed! ## \n"


# format go files to avoid gofumpt linting errors
format:
  ifndef file
	$(error file is not defined)
  else
	gofumpt -w $(file)
  endif

.PHONY: help
help:
	@echo "List of Makefile commands"
	@echo ""
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t

alignment:
	@printf "\e[34m Fixing data alignment... ## \n"

	fieldalignment -fix ./...

	@printf "\e[34m## Passed! ##\e[0m\n"

update_deps:
	go get -u ./...
	go mod tidy

import: 
  ifndef file
	$(error file is not defined)
  else
	gci write --skip-generated -s standard -s default $(file)
  endif

  test:
	@printf "\e[34m Running tests... ## \n"

	go test -race -count=1 ./...

	@printf "\e[34m## All tests passed! ##\e[0m\n"

run-docker:
	docker-compose up

new-hazelcast:
	docker run -d \
	--hostname hazelcast-node \
	--name hazelcast \
	-p 5701:5701 \
	-e HZ_CLUSTERNAME=dev \
	-e JAVA_OPTS="-Dhazelcast.rest.enabled=true -Dhazelcast.http.health-check.enabled=true" \
	hazelcast/hazelcast:5.5

start-hazelcast:
	docker start hazelcast

stop-hazelcast:
	docker stop hazelcast

update-docker-image:
	docker build --no-cache -t=ghcr.io/marincor/rinha-backend-2025-golang:latest .

	make push-docker-image

push-docker-image:
	set -a; source .env; set +a; docker login ghcr.io -u marincor -p $$GITHUB_TOKEN;
	docker push ghcr.io/marincor/rinha-backend-2025-golang:latest

profile-cpu:
	set -a; source .env; set +a; \
	if command -v dot &> /dev/null; then \
		echo "Graphviz installed"; \
	else \
		echo "Graphviz not found. Installing..."; \
		sudo apt update && sudo apt install -y graphviz; \
	fi; \
	curl -o profile.out http://localhost:$${SERVER_PORT}/debug/pprof/profile; \
	go tool pprof -http=:9595 profile.out

profile-mem:
	set -a; source .env; set +a; \
	if command -v dot &> /dev/null; then \
		echo "Graphviz installed"; \
	else \
		echo "Graphviz not found. Installing..."; \
		sudo apt update && sudo apt install -y graphviz; \
	fi; \
	curl -o profile.out http://localhost:$${SERVER_PORT}/debug/pprof/heap; \
	go tool pprof -http=:9595 profile.out