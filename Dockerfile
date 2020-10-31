FROM --platform=$TARGETPLATFORM debian:10-slim
ARG TARGETOS
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

COPY ./build/ddns-${TARGETOS}-${TARGETARCH} /app/ddns

WORKDIR /app
EXPOSE 8080
CMD ./ddns -t ali -d 