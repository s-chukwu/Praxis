FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend .

FROM alpine:3.20

RUN adduser -D -H -u 10001 app
USER app
COPY --from=build /out/backend /usr/local/bin/backend

EXPOSE 8080
ENTRYPOINT ["backend"]
