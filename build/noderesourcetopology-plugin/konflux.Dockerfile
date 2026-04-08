# follow https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70135
FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_golang_1.25@sha256:bd531796aacb86e4f97443797262680fbf36ca048717c00b6f4248465e1a7c0c as builder

ARG COMMIT_SHA
ARG OCP_MAJOR_VERSION=4
ARG OCP_MINOR_VERSION=22

WORKDIR /app

COPY . .

RUN GOEXPERIMENT=strictfipsruntime GOOS=linux CGO_ENABLED=1 go build -ldflags "-X k8s.io/component-base/version.gitMajor=${OCP_MAJOR_VERSION} -X k8s.io/component-base/version.gitMinor=${OCP_MINOR_VERSION} -X k8s.io/component-base/version.gitCommit=${COMMIT_SHA}  -w" -tags strictfipsruntime -o bin/noderesourcetopology-plugin cmd/noderesourcetopology-plugin/main.go

FROM registry.redhat.io/ubi9/ubi-minimal@sha256:d91be7cea9f03a757d69ad7fcdfcd7849dba820110e7980d5e2a1f46ed06ea3b

COPY --from=builder /app/bin/noderesourcetopology-plugin /bin/kube-scheduler
WORKDIR /bin
CMD ["kube-scheduler"]

LABEL com.redhat.component="noderesourcetopology-scheduler-container" \
      name="openshift4/noderesourcetopology-scheduler-rhel9" \
      summary="node resource topology aware scheduler" \
      io.openshift.expose-services="" \
      io.openshift.tags="numa,topology,scheduler" \
      io.k8s.display-name="noderesourcetopology-scheduler" \
      description="kubernetes scheduler aware of node resource topology." \
      maintainer="openshift-operators@redhat.com" \
      io.openshift.maintainer.component="Node Resource Topology aware Scheduler" \
      io.openshift.maintainer.product="OpenShift Container Platform" \
      io.k8s.description="Node Resource Topology aware Scheduler" \
      cpe="cpe:/a:redhat:openshift:4.22::el9" \
      url="https://github.com/konflux-io/noderesourcetopology-plugin"
