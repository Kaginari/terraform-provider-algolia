ifndef VERBOSE
	MAKEFLAGS += --no-print-directory
endif

default: install

.PHONY: install lint unit

#OS_ARCH=linux_amd64
OS_ARCH=windows_amd64
HOSTNAME=registry.terraform.io
NAMESPACE=Kaginari
NAME=algolia
VERSION=0.0.1
## on linux base os
#TERRAFORM_PLUGINS_DIRECTORY=~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
## windows base os
TERRAFORM_PLUGINS_DIRECTORY=C:\plugins\${HOSTNAME}\${NAMESPACE}\${NAME}\${VERSION}\${OS_ARCH}

install: unit
#	mkdir -p ${TERRAFORM_PLUGINS_DIRECTORY}
	go build -o ${TERRAFORM_PLUGINS_DIRECTORY}/terraform-provider-${NAME}.exe
#	cd examples && rm -rf .terraform
#	cd examples && terraform init

lint:
	 golangci-lint run

unit:
	go test ./algolia