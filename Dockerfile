FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/filesync-server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/filesync-server /filesync-server
VOLUME ["/var/lib/filesync"]
ENV FILESYNC_SERVER_ADDR=:8080
ENV FILESYNC_SERVER_STORAGE=/var/lib/filesync
EXPOSE 8080
ENTRYPOINT ["/filesync-server"]
