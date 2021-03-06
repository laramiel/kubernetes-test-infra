# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This file creates a build environment for building and running kubernetes
# unit and integration tests

ARG GO_VERSION=1.9.1
FROM golang:${GO_VERSION}
LABEL maintainer="Benjamin Elder <bentheelder@google.com>"

# Setup workspace and symlink to gopath
WORKDIR /workspace
RUN mkdir -p /go/src/k8s.io/kubernetes /workspace/k8s.io \
    && ln -s /go/src/k8s.io/kubernetes /workspace/k8s.io/kubernetes
ENV WORKSPACE=/workspace \
    TERM=xterm

# Install linux packages
# bc is needed by shell2junit
# dnsutils is needed by federation cluster scripts.
# file is used when uploading test artifacts to GCS.
# jq is used by hack/verify-godep-licenses.sh
# python-pip is needed to install the AWS cli.
# netcat is used by integration test scripts.
RUN apt-get -o Acquire::Check-Valid-Until=false update && apt-get install -y \
        bc \
	dnsutils \
	file \
	jq \
	python-pip \
	netcat-openbsd \
	rsync \
	--no-install-recommends \
	&& rm -rf /var/lib/apt/lists/*

# Install any go packages
# TODO(fejta): migrate this to a unit test image
RUN go get \
        github.com/golang/lint/golint

# install gcloud
ENV GCLOUD_VERSION 163.0.0
RUN wget https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-$GCLOUD_VERSION-linux-x86_64.tar.gz && \
    tar xf google-cloud-sdk-$GCLOUD_VERSION-linux-x86_64.tar.gz -C / && \
    rm google-cloud-sdk-$GCLOUD_VERSION-linux-x86_64.tar.gz && \
    /google-cloud-sdk/install.sh
ENV PATH "/google-cloud-sdk/bin:${PATH}"


#
# BEGIN: DOCKER IN DOCKER SETUP
#

# docker deps, some of these are already installed in the base image but
# that's fine since they won't re-install and we can reuse the code below
# for another image someday
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg2 \
        software-properties-common \
        lsb-release

RUN curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg | apt-key add -

RUN add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
    $(lsb_release -cs) \
    stable"

RUN apt-get update && \
    apt-get install -y --no-install-recommends docker-ce

# Move Docker's storage location
RUN echo 'DOCKER_OPTS="${DOCKER_OPTS} --graph=/docker-graph"' | \
        tee --append /etc/default/docker
# NOTE this should be mounted in from the host ideally (!)
# We will make a fake one now just in case
RUN mkdir /docker-graph

# Add a runner that runs `service docker start` and bootstrap
ADD ["runner", \
    "/workspace/"]

ENTRYPOINT ["/workspace/runner"]

#
# END: DOCKER IN DOCKER SETUP
#
