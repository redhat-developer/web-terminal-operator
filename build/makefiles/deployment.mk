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
