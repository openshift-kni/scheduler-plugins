FROM registry.redhat.io/openshift/golang-builder:golang-builder-v1.26-rhel9@sha256:78bd08b0aa5520dec6f388f5d17c5da6e44ee9b8f826aebc8f3264738e9532e2 as builder

ARG COMMIT_SHA
ARG OCP_MAJOR_VERSION=5
ARG OCP_MINOR_VERSION=0

WORKDIR /app

COPY . .

RUN GOEXPERIMENT=strictfipsruntime GOOS=linux CGO_ENABLED=1 go build -ldflags "-X k8s.io/component-base/version.gitMajor=${OCP_MAJOR_VERSION} -X k8s.io/component-base/version.gitMinor=${OCP_MINOR_VERSION} -X k8s.io/component-base/version.gitCommit=${COMMIT_SHA}  -w" -tags strictfipsruntime -o bin/noderesourcetopology-plugin cmd/noderesourcetopology-plugin/main.go

FROM registry.redhat.io/ubi9/ubi-minimal-pqc@sha256:eb807555968d40a35c815329eec6ef52b56a8bb3e6f06c25f7440ead905f8c74

COPY --from=builder /app/bin/noderesourcetopology-plugin /bin/kube-scheduler
WORKDIR /bin
CMD ["kube-scheduler"]

LABEL com.redhat.component="noderesourcetopology-scheduler-container" \
      name="openshift5/noderesourcetopology-scheduler-rhel9" \
      summary="node resource topology aware scheduler" \
      io.openshift.expose-services="" \
      io.openshift.tags="numa,topology,scheduler" \
      io.k8s.display-name="noderesourcetopology-scheduler" \
      description="kubernetes scheduler aware of node resource topology." \
      maintainer="openshift-operators@redhat.com" \
      io.openshift.maintainer.component="Node Resource Topology aware Scheduler" \
      io.openshift.maintainer.product="OpenShift Container Platform" \
      io.k8s.description="Node Resource Topology aware Scheduler" \
      cpe="cpe:/a:redhat:openshift:5.0::el9" \
      url="https://github.com/openshift-kni/scheduler-plugins"
