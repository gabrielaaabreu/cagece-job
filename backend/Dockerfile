FROM golang:1.25-alpine AS build
RUN apk add --no-cache git build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
RUN go mod download
COPY . .
RUN go build -o /out/app ./

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=build /out/app /usr/local/bin/app
EXPOSE 3000
ENV PORT=3000
ENTRYPOINT ["/usr/local/bin/app"]
