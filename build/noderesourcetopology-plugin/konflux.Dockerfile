# follow https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70135
FROM registry.redhat.io/openshift/golang-builder:golang-builder-v1.26-rhel9@sha256:7e9c3b1b1d35bfc006b18fc5fdff7e0db0ab3136dbba77c2b7f4bd933c0a3235 as builder

ARG COMMIT_SHA
ARG OCP_MAJOR_VERSION=5
ARG OCP_MINOR_VERSION=1

WORKDIR /app

COPY . .

RUN GOEXPERIMENT=strictfipsruntime GOOS=linux CGO_ENABLED=1 go build -ldflags "-X k8s.io/component-base/version.gitMajor=${OCP_MAJOR_VERSION} -X k8s.io/component-base/version.gitMinor=${OCP_MINOR_VERSION} -X k8s.io/component-base/version.gitCommit=${COMMIT_SHA}  -w" -tags strictfipsruntime -o bin/noderesourcetopology-plugin cmd/noderesourcetopology-plugin/main.go

FROM registry.redhat.io/ubi9/ubi-minimal-pqc@sha256:22c70d6d7fa920df140c5facdce1f42178642ca7613590da5466f69418ddf57a

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
      cpe="cpe:/a:redhat:openshift:5.1::el9" \
      url="https://github.com/openshift-kni/scheduler-plugins"
