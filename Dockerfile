#######################################################
## MailHog Dockerfile
## First Stage: builder
FROM golang:1.13-alpine AS builder

### Building MailHog:
COPY . /code
WORKDIR /code
RUN go install .

## Second Stage: final
FROM golang:1.13-alpine

COPY --from=builder /go/bin/mailhog /usr/local/bin/mailhog

### Add mailhog user/group with uid/gid 1000.
### This is a workaround for boot2docker issue #581, see
### https://github.com/boot2docker/boot2docker/issues/581
RUN adduser -D -u 1000 mailhog

USER mailhog

WORKDIR /home/mailhog

ENTRYPOINT ["/usr/local/bin/mailhog"]

### Expose the SMTP and HTTP ports:
EXPOSE 1025 8025
