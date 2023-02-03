ARG BUILDER
FROM ${BUILDER} as builder

FROM python:3.7-alpine3.15

ENV OPERATOR=/usr/local/bin/spinnaker-operator \
    USER_UID=1001 \
    USER_NAME=spinnaker-operator \
    AWS_AIM_AUTHENTICATOR_VERSION=0.4.0 \
    KUBECTL_RELEASE=1.17.7 \
    AWS_CLI_VERSION=1.27.63 \
    OPERATOR_HOME=/opt/spinnaker-operator \
    GOOGLE_CLOUD_SDK_VERSION=313.0.1 \
    PATH="$PATH:/usr/local/bin/:/opt/google-cloud-sdk/bin/:/usr/local/bin/aws-iam-authenticator"

EXPOSE 8383
RUN apk update                        \
	&& apk add ca-certificates bash curl wget unzip \
	&& adduser -D -u ${USER_UID} ${USER_NAME} \
	&& apk upgrade

# Google cloud SDK with anthos removed for CVE and because we don't need it
RUN wget -nv https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz \
  && mkdir -p /opt && cd /opt \
  && tar -xzf /google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz \
  && rm /google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz \
  && CLOUDSDK_PYTHON="python3" /opt/google-cloud-sdk/install.sh --usage-reporting=false --bash-completion=false  --additional-components app-engine-java app-engine-go \
  && rm -rf ~/.config/gcloud \
  && gcloud components remove --quiet anthoscli \
  && rm -rf /opt/google-cloud-sdk/.install/.backup

# kubectl + AWS IAM authenticator
RUN wget https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_RELEASE}/bin/linux/amd64/kubectl \
  && chmod +x kubectl \
  && mv ./kubectl /usr/local/bin/kubectl \
  && wget -O aws-iam-authenticator https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${AWS_AIM_AUTHENTICATOR_VERSION}/aws-iam-authenticator_${AWS_AIM_AUTHENTICATOR_VERSION}_linux_amd64 \
  && chmod +x ./aws-iam-authenticator \
  && mv ./aws-iam-authenticator /usr/local/bin/aws-iam-authenticator

# Install aws-cli
RUN pip install --upgrade awscli==${AWS_CLI_VERSION}  \
    && pip uninstall -y pip

USER ${USER_NAME}

# Everything after this line is never cached
ARG CACHE_DATE

RUN echo "CACHE_DATE: ${CACHE_DATE}"
COPY --from=builder /opt/spinnaker-operator/build/build/bin/linux_amd64/spinnaker-operator ${OPERATOR}
COPY --from=builder /opt/spinnaker-operator/build/build/bin/linux_amd64/MANIFEST ${OPERATOR_HOME}/MANIFEST
COPY --from=builder /opt/spinnaker-operator/build/build-tools/entrypoint /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]
