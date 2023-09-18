FROM --platform=$TARGETPLATFORM debian:12.1-slim
ARG TARGETOS
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates \
	&& rm -rf /var/lib/apt/lists/*
# RUN apt-get update 
# RUN apt search glibc
# CMD /lib/x86_64-linux-gnu/libc.so.6
COPY ./build/ddns-${TARGETOS}-${TARGETARCH} /app/ddns

WORKDIR /app
CMD ./ddns -t ali -d 