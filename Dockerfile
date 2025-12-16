FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    wget \
    && rm -rf /var/lib/apt/lists/*

ARG GO_VERSION

RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

ARG TINYGO_VERSION

RUN wget -q https://github.com/tinygo-org/tinygo/releases/download/v${TINYGO_VERSION}/tinygo_${TINYGO_VERSION}_amd64.deb \
    && dpkg -i tinygo_${TINYGO_VERSION}_amd64.deb \
    && rm tinygo_${TINYGO_VERSION}_amd64.deb

ENV PATH="/usr/local/bin:${PATH}"

RUN useradd -m builder
USER builder

ENV GOPATH="/home/builder/go"
ENV GOCACHE="/home/builder/.cache/go-build"

WORKDIR /workspace
