FROM golang:1.17.3

WORKDIR /app/cmd/appCode
# File changes must be added at the very end, to avoid the installation of dependencies again and again
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
COPY go.mod .
COPY go.sum .
# can not find this file in the directory structure
RUN go mod download

COPY . .

CMD ["air"]

#FROM golang:1.16.5 as builder
## Define build env
#ENV GOOS linux
#ENV CGO_ENABLED 0
## Add a work directory
#WORKDIR /app
## Cache and install dependencies
#COPY go.mod go.sum ./
#RUN go mod download
## Copy app files
#COPY . .
## Build app
#RUN go build -o app
#
#FROM alpine:3.14 as production
## Add certificates
#RUN apk add --no-cache ca-certificates
## Copy built binary from builder
#COPY --from=builder app .
## Expose port
#EXPOSE 4000
## Exec built binary
#CMD ./app