ARG GO_VERSION
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

FROM docker.io/library/golang:${GO_VERSION:-1.25}-alpine AS builder
RUN apk add --no-cache make git
WORKDIR /go/src/github.com/byxorna/multiband
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make bin && ./bin/multiband version

FROM docker.io/library/alpine:latest

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch
COPY --from=builder /go/src/github.com/byxorna/multiband/bin/multiband /bin/multiband
ENTRYPOINT ["/bin/multiband"]

