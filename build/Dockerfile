FROM golang:1.12.1-alpine3.9 as builder

# vendor flags conflict with `go get`
# so we fetch golint before running make
# and setting the env variable
RUN apk update && apk add git make bash build-base gcc bc
RUN go get -u golang.org/x/lint/golint

ENV GO111MODULE=on GOOS=linux GOARCH=amd64
WORKDIR /opt/spinnaker-operator/build/
ADD ./ /opt/spinnaker-operator/build/
ARG OPERATOR_VERSION
ENV OPERATOR_VERSION=$OPERATOR_VERSION
RUN make

FROM alpine:3.9.3

ARG OPERATOR_PATH
ENV OPERATOR=/usr/local/bin/spinnaker-operator \
    USER_UID=1001 \
    USER_NAME=spinnaker-operator

EXPOSE 8383
RUN apk update                        \
	&& apk add ca-certificates bash   \
	&& adduser -D -u ${USER_UID} ${USER_NAME}

COPY --from=builder /opt/spinnaker-operator/build/$OPERATOR_PATH ${OPERATOR}
COPY --from=builder /opt/spinnaker-operator/build/build/entrypoint /usr/local/bin/entrypoint

USER ${USER_NAME}
ENTRYPOINT ["/usr/local/bin/entrypoint"]



