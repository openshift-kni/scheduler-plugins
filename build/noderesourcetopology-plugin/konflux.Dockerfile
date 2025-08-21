FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_golang_1.23@sha256:96cfceb50f5323efa1aa8569d4420cdbf1bb391225d5171ef72a0d0ecf028467 as builder

ARG COMMIT_SHA
ARG OCP_MAJOR_VERSION=4
ARG OCP_MINOR_VERSION=20

WORKDIR /app

COPY . .

RUN GOEXPERIMENT=strictfipsruntime GOOS=linux CGO_ENABLED=1 go build -ldflags "-X k8s.io/component-base/version.gitMajor=${OCP_MAJOR_VERSION} -X k8s.io/component-base/version.gitMinor=${OCP_MINOR_VERSION} -X k8s.io/component-base/version.gitCommit=${COMMIT_SHA}  -w" -tags strictfipsruntime -o bin/noderesourcetopology-plugin cmd/noderesourcetopology-plugin/main.go

FROM registry.redhat.io/rhel9-4-els/rhel-minimal:9.4@sha256:574decff4ac70ae46f9e6c562a20e429c3303b5173f2529b6e5f8a13827cdd7d

COPY --from=builder /app/bin/noderesourcetopology-plugin /bin/kube-scheduler
WORKDIR /bin
CMD ["kube-scheduler"]

LABEL com.redhat.component="noderesourcetopology-scheduler-container" \
      name="openshift4/noderesourcetopology-scheduler" \
      summary="node resource topology aware scheduler" \
      io.openshift.expose-services="" \
      io.openshift.tags="numa,topology,scheduler" \
      io.k8s.display-name="noderesourcetopology-scheduler" \
      description="kubernetes scheduler aware of node resource topology." \
      maintainer="openshift-operators@redhat.com" \
      io.openshift.maintainer.component="Node Resource Topology aware Scheduler" \
      io.openshift.maintainer.product="OpenShift Container Platform" \
      io.k8s.description="Node Resource Topology aware Scheduler"

