#
# Copyright (c) 2019-2022 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#

SHELL := bash
.SHELLFLAGS = -ec
.ONESHELL:

ifndef VERBOSE
  MAKEFLAGS += --silent
endif

# Enable using Podman instead of Docker
export DOCKER ?= podman
export WEB_TERMINAL_EXEC_IMG ?= quay.io/wto/web-terminal-exec:next


### fmt: Formats code using goimports
fmt:
  ifneq ($(shell command -v goimports 2> /dev/null),)
	  find . -name '*.go' -exec goimports -w {} \;
  else
	  @echo "WARN: goimports is not installed -- formatting using go fmt instead."
	  @echo "      Please install goimports to ensure file imports are consistent."
	  go fmt -x ./...
  endif

### fmt_license: Ensures the license header is set on all files
fmt_license:
  ifneq ($(shell command -v addlicense 2> /dev/null),)
	  @echo 'addlicense -v -f license_header.txt **/*.go'
	  addlicense -v -f license_header.txt $$(find . -name '*.go')
  else
	  $(error addlicense must be installed for this rule: go get -u github.com/google/addlicense)
  endif

### vet: Runs go vet against code
vet:
	go vet ./...

### check_fmt: Checks the formatting on files in repo
check_fmt:
  ifeq ($(shell command -v goimports 2> /dev/null),)
	  $(error "goimports must be installed for this rule" && exit 1)
  endif
  ifeq ($(shell command -v addlicense 2> /dev/null),)
	  $(error "error addlicense must be installed for this rule: go get -u github.com/google/addlicense")
  endif
	@{
	  if [[ $$(find . -name '*.go' -exec goimports -l {} \;) != "" ]]; then \
	    echo "Files not formatted; run 'make fmt'"; exit 1 ;\
	  fi ;\
	  if ! addlicense -check -f license_header.txt $$(find . -name '*.go'); then \
	    echo "Licenses are not formatted; run 'make fmt_license'"; exit 1 ;\
	  fi \
	}

### compile: Compiles Web Terminal Exec
compile:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) GO111MODULE=on go build \
	  -a -o _output/bin/web-terminal-exec \
	  -gcflags all=-trimpath=/ \
	  -asmflags all=-trimpath=/ \
	  -ldflags "-X $(GO_PACKAGE_PATH)/version.Commit=$(GIT_COMMIT_ID) \
	  -X $(GO_PACKAGE_PATH)/version.BuildTime=$(BUILD_TIME)" \
	main.go

### docker: Builds and pushes Web Terminal Exec image
docker: docker-build docker-push

### docker-build: Builds the Web Terminal Exec image
docker-build:
	$(DOCKER) build . -t ${WEB_TERMINAL_EXEC_IMG} -f build/Dockerfile

### docker-push: Web Terminal Exec image
docker-push:
  ifneq ($(INITIATOR),CI)
    ifeq ($(WEB_TERMINAL_EXEC_IMG),quay.io/wto/web-terminal-exec:next)
	    @echo -n "Are you sure you want to push $(WEB_TERMINAL_EXEC_IMG)? [y/N] " && read ans && [ $${ans:-N} = y ]
    endif
  endif
	$(DOCKER) push ${WEB_TERMINAL_EXEC_IMG}
