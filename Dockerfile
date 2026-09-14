FROM golang:1.27-alpine AS toolchain
WORKDIR /src
RUN apk add --no-cache git curl
RUN go install github.com/air-verse/air@v1.61.7
COPY go.mod ./
RUN go mod download
COPY . .

FROM toolchain AS dev
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

FROM toolchain AS build
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/lantern ./cmd/lantern \
	&& mkdir -p /out/data/cache/icons

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/lantern /lantern
COPY --from=build /out/data /data
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/lantern"]
