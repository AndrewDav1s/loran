ARG IMG_TAG=latest

# Compile the loran binary
FROM golang:1.17-alpine AS loran-builder
WORKDIR /src/app/
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ENV PACKAGES make git libc-dev bash gcc linux-headers
RUN apk add --no-cache $PACKAGES
RUN make install

# Fetch hilod binary
FROM golang:1.17-alpine AS hilod-builder
ARG HILO_VERSION=bez/gb-module-poc
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev
RUN apk add --no-cache $PACKAGES
WORKDIR /downloads/
RUN git clone https://github.com/cicizeo/hilo.git
RUN cd hilo && git checkout ${HILO_VERSION} && make build && cp ./build/hilod /usr/local/bin/

# Add to a distroless container
FROM gcr.io/distroless/cc:$IMG_TAG
ARG IMG_TAG
COPY --from=loran-builder /go/bin/loran /usr/local/bin/
COPY --from=hilod-builder /usr/local/bin/hilod /usr/local/bin/
EXPOSE 26656 26657 1317 9090
