# syntax=docker/dockerfile:1.7

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/mind-daemon-dmz ./cmd/mind-daemon-dmz

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/mind-daemon-dmz /mind-daemon-dmz
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/mind-daemon-dmz"]
