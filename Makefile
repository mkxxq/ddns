.PHONY: default build
app = ddns
image = $(app)
compile = go build -a -ldflags '-s -w --extldflags "-static -fpic"'
src = ./cmd/ddns

default:
	$(compile) -o ./build/${app} $(src)
build_arm:
	GOOS=linux GOARCH=arm64 $(compile) -o ./build/${app}-linux-arm64 $(src)
	docker buildx build --platform linux/arm64 -t $(image)-arm64 .
build:
	GOOS=linux $(compile) -o ./build/${app}-linux-amd64 $(src)
	docker buildx build --platform linux/amd64 -t $(image)-amd64 .