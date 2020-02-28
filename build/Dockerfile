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
    USER_NAME=spinnaker-operator \
    AWS_BINARY_RELEASE_DATE=2019-08-22 \
    KUBECTL_RELEASE=1.14.6 \
    OPERATOR_HOME=/opt/spinnaker-operator

EXPOSE 8383
RUN apk update                        \
	&& apk add ca-certificates bash curl wget unzip 'python2>2.7.9' \
	&& adduser -D -u ${USER_UID} ${USER_NAME}

# Install aws iam authenticator
RUN curl -o aws-iam-authenticator https://amazon-eks.s3-us-west-2.amazonaws.com/${KUBECTL_RELEASE}/${AWS_BINARY_RELEASE_DATE}/bin/linux/amd64/aws-iam-authenticator && \
  chmod +x ./aws-iam-authenticator && \
  mv ./aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
ENV PATH "$PATH:/usr/local/bin/aws-iam-authenticator"

# Install gcloud
RUN wget -nv https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.zip && \
    unzip -qq google-cloud-sdk.zip -d /opt && \
    rm google-cloud-sdk.zip && \
    CLOUDSDK_PYTHON="python2.7" /opt/google-cloud-sdk/install.sh --usage-reporting=false --bash-completion=false && \
    rm -rf ~/.config/gcloud
ENV PATH=$PATH:/opt/google-cloud-sdk/bin/

COPY --from=builder /opt/spinnaker-operator/build/$OPERATOR_PATH ${OPERATOR}
COPY --from=builder /opt/spinnaker-operator/build/build/entrypoint /usr/local/bin/entrypoint
COPY --from=builder /opt/spinnaker-operator/build/build/MANIFEST ${OPERATOR_HOME}/MANIFEST

USER ${USER_NAME}
ENTRYPOINT ["/usr/local/bin/entrypoint"]



