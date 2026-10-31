# follow https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70135
FROM registry.redhat.io/openshift/golang-builder:golang-builder-v1.26-rhel9@sha256:0a93368c78345f50b59d0cd1733505015b9b6d532d060bf097b59c8a2c0b85dc as builder

ARG COMMIT_SHA
ARG OCP_MAJOR_VERSION=5
ARG OCP_MINOR_VERSION=1

WORKDIR /app

COPY . .

RUN GOFIPS140=v1.26.0 GOEXPERIMENT=strictfipsruntime GOOS=linux CGO_ENABLED=0 go build -ldflags "-X k8s.io/component-base/version.gitMajor=${OCP_MAJOR_VERSION} -X k8s.io/component-base/version.gitMinor=${OCP_MINOR_VERSION} -X k8s.io/component-base/version.gitCommit=${COMMIT_SHA}  -w" -tags strictfipsruntime,no_openssl -o bin/noderesourcetopology-plugin cmd/noderesourcetopology-plugin/main.go

FROM registry.redhat.io/ubi9/ubi-minimal-pqc@sha256:58d09489b1d9123e24d691e47af6c390fe5e08f50c7f4f2967bb64df7ab72aa0

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
