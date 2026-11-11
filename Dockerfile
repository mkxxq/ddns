FROM --platform=$TARGETPLATFORM debian:12.1-slim
ARG TARGETOS
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates \
	&& rm -rf /var/lib/apt/lists/*
COPY ./build/ddns-${TARGETOS}-${TARGETARCH} /app/ddns

WORKDIR /app
CMD ["./ddns", "-t", "cloudflare"]
