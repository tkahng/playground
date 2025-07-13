#react stage
FROM node:slim AS react-build
WORKDIR /app
COPY . .
RUN npm --prefix ./ui install
RUN npm --prefix ./ui run build

#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY --from=react-build /app .
RUN go get -d -v ./...
RUN go build -o /go/bin/app -v .

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
ENTRYPOINT ["/app", "serve"]
LABEL Name=playground Version=0.0.1
EXPOSE 8080
