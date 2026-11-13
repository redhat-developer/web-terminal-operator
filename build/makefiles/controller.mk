### fmt: Runs go fmt against code
fmt:
ifneq ($(shell command -v goimports 2> /dev/null),)
	find . -name '*.go' -exec goimports -w {} \;
else
	@echo "WARN: goimports is not installed -- formatting using go fmt instead."
	@echo "      Please install goimports to ensure file imports are consistent."
	go fmt -x ./...
endif

### check_fmt: Checks the formatting on go files in repo
check_fmt:
ifeq ($(shell command -v goimports 2> /dev/null),)
	$(error "goimports must be installed for this rule" && exit 1)
endif
	@{
		if [[ $$(find . -name '*.go' -exec goimports -l {} \;) != "" ]]; then \
			echo "Files not formatted; run 'make fmt'"; exit 1 ;\
		fi ;\
	}

### compile: Build binary for Web Terminal Operator
compile:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) GO111MODULE=on go build \
	-a -o _output/bin/web-terminal-controller \
	-gcflags all=-trimpath=/ \
	-asmflags all=-trimpath=/ \
	-ldflags "-X $(GO_PACKAGE_PATH)/version.Commit=$(GIT_COMMIT_ID) \
	-X $(GO_PACKAGE_PATH)/version.BuildTime=$(BUILD_TIME)" \
	main.go

### build_controller_image: Build container image for Web Terminal Operator
build_controller_image:
	$(DOCKER) build -t $(WTO_IMG) -f build/dockerfiles/controller.Dockerfile .
ifneq ($(INITIATOR),CI)
ifeq ($(WTO_IMG),quay.io/wto/web-terminal-operator:next)
	@echo -n "Are you sure you want to push $(WTO_IMG)? [y/N] " && read ans && [ $${ans:-N} = y ]
endif
endif
	$(DOCKER) push $(WTO_IMG)
