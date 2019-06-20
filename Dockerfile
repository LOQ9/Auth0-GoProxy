FROM alpine:3.9
RUN apk add --no-cache ca-certificates
ADD . /build/
CMD ["/auth0-goproxy"]
