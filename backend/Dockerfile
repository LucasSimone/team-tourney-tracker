FROM golang:1.21-bullseye AS build
WORKDIR /app
COPY . .
# Install Debian build deps for cgo and sqlite
RUN apt-get update && apt-get install -y build-essential libsqlite3-dev git ca-certificates && rm -rf /var/lib/apt/lists/*
ENV CGO_ENABLED=1
RUN go mod tidy && go build -o server .

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]

