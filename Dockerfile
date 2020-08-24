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
ADD manifests /manifests

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
      com.redhat.component="$PRODNAME-rhel8-$COMPNAME-container" \
      name="$PRODNAME/$COMPNAME" \
      version="1.0.0" \
      license="MIT" \
      maintainer="Josh Pinkney <jpinkney@redhat.com>" \
      io.openshift.expose-services="" \
      com.redhat.delivery.appregistry="false" \
      com.redhat.delivery.operator.bundle=true \
      com.redhat.openshift.versions="v4.5" \
      usage=""
