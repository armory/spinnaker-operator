# avoid emulation at build step: https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/
FROM --platform=$BUILDPLATFORM golang:1.18.0-alpine3.15  as builder

ARG VERSION
ARG BUILDARCH

# vendor flags conflict with `go get`
# so we fetch golint before running make
# and setting the env variable
RUN apk update && apk add git make bash build-base gcc bc
RUN go install golang.org/x/lint/golint@v0.0.0-20210508222113-6edffad5e616

ENV GO111MODULE=on GOOS=linux GOARCH=${BUILDARCH}
WORKDIR /opt/spinnaker-operator/build/
ADD ./ /opt/spinnaker-operator/build/
RUN go mod vendor && go mod tidy

# Build both architectures without emulation
RUN make build VERSION=${VERSION} OS=linux ARCH=amd64
RUN make build VERSION=${VERSION} OS=linux ARCH=arm64


FROM --platform=$BUILDPLATFORM builder as tests
ARG BUILDARCH

ENV GO111MODULE=on GOOS=linux GOARCH=${TARGETARCH}
WORKDIR /opt/spinnaker-operator/build/
RUN make test OS=linux ARCH=${BUILDARCH}



FROM python:3.7-alpine3.15
ARG TARGETARCH

ENV OPERATOR=/usr/local/bin/spinnaker-operator \
    USER_UID=1001 \
    USER_NAME=spinnaker-operator \
    AWS_AIM_AUTHENTICATOR_VERSION=0.5.5 \
    KUBECTL_RELEASE=1.17.7 \
    AWS_CLI_VERSION=1.18.109 \
    OPERATOR_HOME=/opt/spinnaker-operator \
    GOOGLE_CLOUD_SDK_VERSION=379.0.0 \
    PATH="$PATH:/usr/local/bin/:/opt/google-cloud-sdk/bin/:/usr/local/bin/aws-iam-authenticator"

EXPOSE 8383

RUN apk update                        \
	&& apk add ca-certificates bash curl wget unzip \
	&& adduser -D -u ${USER_UID} ${USER_NAME} \
	&& apk upgrade

# Google cloud SDK with anthos removed for CVE and because we don't need it
RUN [ $TARGETARCH == 'amd64' ] &&  export GCP_ARCH="x86_64" || export GCP_ARCH="arm" \
  && wget -nv https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-${GCP_ARCH}.tar.gz \
  && mkdir -p /opt && cd /opt \
  && tar -xzf /google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-${GCP_ARCH}.tar.gz \
  && rm /google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-${GCP_ARCH}.tar.gz \
  && CLOUDSDK_PYTHON="python3" /opt/google-cloud-sdk/install.sh --usage-reporting=false --bash-completion=false  --additional-components app-engine-java app-engine-go \
  && rm -rf ~/.config/gcloud \
  && gcloud components remove --quiet anthoscli \
  && rm -rf /opt/google-cloud-sdk/.install/.backup

 # kubectl + AWS IAM authenticator
RUN wget https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_RELEASE}/bin/linux/${TARGETARCH}/kubectl \
  && chmod +x kubectl \
  && mv ./kubectl /usr/local/bin/kubectl \
  && wget -O aws-iam-authenticator https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${AWS_AIM_AUTHENTICATOR_VERSION}/aws-iam-authenticator_${AWS_AIM_AUTHENTICATOR_VERSION}_linux_${TARGETARCH} \
  && chmod +x ./aws-iam-authenticator \
  && mv ./aws-iam-authenticator /usr/local/bin/aws-iam-authenticator

# Install aws-cli
RUN pip install --upgrade awscli==${AWS_CLI_VERSION}  \
    && pip uninstall -y pip

USER ${USER_NAME}

# Everything after this line is never cached
ARG CACHE_DATE

RUN echo "CACHE_DATE: ${CACHE_DATE}"
COPY --from=builder /opt/spinnaker-operator/build/build/bin/linux_${TARGETARCH}/spinnaker-operator ${OPERATOR}
COPY --from=builder /opt/spinnaker-operator/build/build/bin/linux_${TARGETARCH}/MANIFEST ${OPERATOR_HOME}/MANIFEST
COPY --from=builder /opt/spinnaker-operator/build/build-tools/entrypoint /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]
