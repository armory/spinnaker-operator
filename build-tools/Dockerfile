ARG BUILDER
FROM ${BUILDER} as builder

FROM alpine:3.11

ENV OPERATOR=/usr/local/bin/spinnaker-operator \
    USER_UID=1001 \
    USER_NAME=spinnaker-operator \
    AWS_BINARY_RELEASE_DATE=2019-08-22 \
    KUBECTL_RELEASE=1.14.6 \
    AWS_CLI_RELEASE=1.18.58 \
    OPERATOR_HOME=/opt/spinnaker-operator

EXPOSE 8383
RUN apk update                        \
	&& apk add ca-certificates bash curl wget unzip python3 py-pip \
	&& adduser -D -u ${USER_UID} ${USER_NAME} \
	&& apk upgrade

# Install aws iam authenticator
RUN curl -o aws-iam-authenticator https://amazon-eks.s3-us-west-2.amazonaws.com/${KUBECTL_RELEASE}/${AWS_BINARY_RELEASE_DATE}/bin/linux/amd64/aws-iam-authenticator && \
  chmod +x ./aws-iam-authenticator && \
  mv ./aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
ENV PATH "$PATH:/usr/local/bin/aws-iam-authenticator"

# Install aws-cli
RUN pip install --upgrade awscli==${AWS_CLI_RELEASE}

# Install gcloud
RUN wget -nv https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.zip && \
    unzip -qq google-cloud-sdk.zip -d /opt && \
    rm google-cloud-sdk.zip && \
    CLOUDSDK_PYTHON="python3" /opt/google-cloud-sdk/install.sh --usage-reporting=false --bash-completion=false && \
    rm -rf ~/.config/gcloud
ENV PATH=$PATH:/opt/google-cloud-sdk/bin/

USER ${USER_NAME}

# Everything after this line is never cached
ARG CACHE_DATE

RUN echo "CACHE_DATE: ${CACHE_DATE}"
COPY --from=builder /opt/spinnaker-operator/build/build/bin/linux_amd64/spinnaker-operator ${OPERATOR}
COPY --from=builder /opt/spinnaker-operator/build/build/bin/linux_amd64/MANIFEST ${OPERATOR_HOME}/MANIFEST
COPY --from=builder /opt/spinnaker-operator/build/build-tools/entrypoint /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]
