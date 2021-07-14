ifndef VERBOSE
	MAKEFLAGS += --no-print-directory
endif

default: install

.PHONY: install lint unit

OS_ARCH=linux_amd64
HOSTNAME=registry.terraform.io
NAMESPACE=Kaginari
NAME=algolia
VERSION=9.9.9
TERRAFORM_PLUGINS_DIRECTORY=${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

install:
	mkdir -p ${TERRAFORM_PLUGINS_DIRECTORY}
	go build -o ${TERRAFORM_PLUGINS_DIRECTORY}/terraform-provider-${NAME}
	cd example && rm -rf .terraform && rm -rf .terraform.lock.hcl
	cd example && make init

lint:
	 golangci-lint run

unit:
	go test ./algolia