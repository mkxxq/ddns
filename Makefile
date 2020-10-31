.PHONY: default build
app = ddns
image = xiaoqiang321/$(app)
compile = go build -a -ldflags '-s -w --extldflags "-static -fpic"'
src = ./cmd/ddns

default:
	$(compile) -o ./build/${app} $(src)
build:

	GOOS=linux GOARCH=arm64 $(compile) -o ./build/${app}-linux-arm64 $(src)
	docker buildx build --platform linux/arm64 -t $(image) . --push