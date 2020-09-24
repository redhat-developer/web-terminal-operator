# Copyright (c) 2020 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#

FROM scratch
COPY manifests /manifests
COPY metadata /metadata

# These are three labels needed to control how the pipeline should handle this container image
# This first label tells the pipeline that this is a bundle image and should be
# delivered via an index image
LABEL com.redhat.delivery.operator.bundle=true

# This second label tells the pipeline which versions of OpenShift the operator supports (4.5+).
# This is used to control which index images should include this operator.
LABEL com.redhat.openshift.versions="v4.5"

# This third label tells the pipeline that this operator should *also* be supported on OCP 4.4 and
# earlier.  It is used to control whether or not the pipeline should attempt to automatically
# backport this content into the old appregistry format and upload it to the quay.io application
# registry endpoints.
LABEL com.redhat.delivery.backport=true

# The rest of these labels are copies of the same content in annotations.yaml and are needed by OLM
# Note the package name and channels which are very important!
LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=web-terminal
LABEL operators.operatorframework.io.bundle.channels.v1=alpha
LABEL operators.operatorframework.io.bundle.channel.default.v1=alpha

# append Brew metadata here
ENV SUMMARY="Web Terminal operator-metadata container" \
    DESCRIPTION="Web Terminal operator-metadata container" \
    PRODNAME="web-terminal" \
    COMPNAME="operator-metadata"
 
LABEL summary="$SUMMARY" \
      description="$DESCRIPTION" \
      io.k8s.description="$DESCRIPTION" \
      io.k8s.display-name="$DESCRIPTION" \
      io.openshift.tags="$PRODNAME,$COMPNAME" \
      com.redhat.component="$PRODNAME-$COMPNAME-container" \
      name="$PRODNAME/$COMPNAME" \
      version="1.0.2" \
      license="EPLv2" \
      maintainer="Joshua Pinkney <jpinkney@redhat.com>" \
      io.openshift.expose-services="" \
      com.redhat.delivery.operator.bundle=true \
      com.redhat.openshift.versions="v4.5" \
      usage=""
