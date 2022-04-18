# syntax=docker/dockerfile:1
# Using a multi-stage build to reduce image size (483MB -> 27MB)

##
## Build
##
FROM golang:1.17-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /wipercheck-loader ./cmd/loader

##
## Deploy
##
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /wipercheck-loader /wipercheck-loader

USER nonroot:nonroot

ENTRYPOINT ["/wipercheck-loader"]