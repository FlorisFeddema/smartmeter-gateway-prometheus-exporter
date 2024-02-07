FROM golang:1.22 AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app

FROM scratch as runtime

LABEL org.opencontainers.image.source="https://github.com/FlorisFeddema/smartmeter-gateway-prometheus-exporter"
LABEL org.opencontainers.image.description="Image for smartmeter-gateway-prometheus-exporter application"
LABEL org.opencontainers.image.licenses=Apache

USER app
COPY --from=build /build/app /app
ENTRYPOINT ["/app"]
