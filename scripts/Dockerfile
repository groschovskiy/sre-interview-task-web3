# Prepare the build environment and build the image
FROM golang:1.22.4-alpine as builder
RUN mkdir -p /build
ADD * /build/
WORKDIR /build
RUN apk add git && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -o cloud-lb cmd/main.go

# Prepare the image for the final run
FROM alpine:3.20.0
COPY --from=builder /build/cloud-lb .

# Run the image with entrypoint
ENTRYPOINT [ "./cloud-lb" ]