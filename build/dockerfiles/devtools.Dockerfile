# Copyright (c) 2021-2024 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#
FROM quay.io/devfile/base-developer-image:ubi9-latest

USER root

# Install gettext
RUN dnf install -y gettext make jq python3-pip && \
    pip3 install yq

# Install the Operator SDK
ENV OPERATOR_SDK_VERSION="v0.17.2"
ENV OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}
RUN curl -sSLO ${OPERATOR_SDK_DL_URL}/operator-sdk_linux_amd64 && \
    chmod +x operator-sdk_linux_amd64 && \
    mv operator-sdk_linux_amd64 /usr/local/bin/operator-sdk

# Install opm CLI
ENV OPM_VERSION="v1.13.1"
RUN curl -sSLO https://github.com/operator-framework/operator-registry/releases/download/${OPM_VERSION}/linux-amd64-opm && \
    chmod +x linux-amd64-opm && \
    mv linux-amd64-opm /usr/local/bin/opm

ENV OC_VERSION=4.19
RUN curl -L https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable-${OC_VERSION}/openshift-client-linux.tar.gz | \
    tar -C /usr/local/bin -xz --no-same-owner && \
    chmod +x /usr/local/bin/oc

USER 1001

