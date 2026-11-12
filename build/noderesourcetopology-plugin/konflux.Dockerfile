FROM registry.redhat.io/openshift/golang-builder:golang-builder-v1.24-rhel9@sha256:3bc38203ab4627012320b3b1da5a0a81cc6987afd0977e731b583db222e31d55 as builder

ARG COMMIT_SHA
ARG OCP_MAJOR_VERSION=4
ARG OCP_MINOR_VERSION=21

WORKDIR /app

COPY . .

RUN GOEXPERIMENT=strictfipsruntime GOOS=linux CGO_ENABLED=1 go build -ldflags "-X k8s.io/component-base/version.gitMajor=${OCP_MAJOR_VERSION} -X k8s.io/component-base/version.gitMinor=${OCP_MINOR_VERSION} -X k8s.io/component-base/version.gitCommit=${COMMIT_SHA}  -w" -tags strictfipsruntime -o bin/noderesourcetopology-plugin cmd/noderesourcetopology-plugin/main.go

FROM registry.redhat.io/ubi9/ubi-minimal@sha256:4b79c93106c14ad1c234907f83441ff0e4f8d24e78426a36e159098e9f7e6fa5

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
      cpe="cpe:/a:redhat:openshift:4.21::el9"
