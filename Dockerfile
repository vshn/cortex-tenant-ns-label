FROM golang:stretch as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0

WORKDIR /build

# Cache modules
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build

WORKDIR /dist

RUN cp /build/cortex-tenant ./cortex-tenant

RUN ldd cortex-tenant | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname ./%); cp % ./%;'
RUN mkdir -p lib64 && cp /lib64/ld-linux-x86-64.so.2 lib64/


FROM busybox
COPY --from=builder /build/cortex-tenant /
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/bin/sh", "-c", "/bin/echo \"${CONFIG}\" > /tmp/cortex-tenant.yml; /cortex-tenant -config /tmp/cortex-tenant.yml"]
