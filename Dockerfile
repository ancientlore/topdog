FROM golang:1.14.4 as builder
WORKDIR /go/src/topdog
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go install
WORKDIR /go/foo
RUN echo "root:x:0:0:user:/home:/bin/bash" > passwd && echo "nobody:x:65534:65534:user:/home:/bin/bash" >> passwd
RUN echo "root:x:0:" > group && echo "nobody:x:65534:" >> group

FROM gcr.io/distroless/static:latest
LABEL Description="Who's the top dog?"
COPY --from=builder /go/foo/group /etc/group
COPY --from=builder /go/foo/passwd /etc/passwd
  
COPY --from=builder /go/bin/topdog /topdog
COPY --from=builder /go/src/topdog/static /static
WORKDIR /

# Needed to know what port to listen on
ENV SERVICE_PORT=5000

EXPOSE 5000/tcp
USER nobody:nobody

ENTRYPOINT ["/topdog", "-static", "/static"]
