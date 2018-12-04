FROM golang as builder
WORKDIR /go/src/topdog
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go install

FROM gcr.io/distroless/base
LABEL Description="Who's the top dog?"
  
COPY --from=builder /go/bin/topdog /topdog
COPY --from=builder /go/src/topdog/static /static
WORKDIR /

# Needed to know what port to listen on
ENV SERVICE_PORT=5000

EXPOSE 5000/tcp

ENTRYPOINT ["/topdog", "-static", "/static"]
