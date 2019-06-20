FROM alpine:3.9
RUN apk add --no-cache ca-certificates
COPY build/auth0-goproxy .
RUN ls -la
ENTRYPOINT ["/auth0-goproxy"]
