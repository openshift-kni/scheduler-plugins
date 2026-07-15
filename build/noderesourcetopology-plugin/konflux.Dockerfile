FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_8_golang_1.19@sha256:08316d1d9032a0440a1ba68b8dccaa8fdf65c4f3baa99e3718f5096774b4d7ea as builder

WORKDIR /app

COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -o bin/noderesourcetopology-plugin cmd/noderesourcetopology-plugin/main.go

FROM registry.redhat.io/ubi8/ubi-minimal:latest@sha256:7eedea3d199fd9b0e6a06cd1a22537c1b84fab65d6eec8a5049c12e68c57f692

COPY --from=builder /app/bin/noderesourcetopology-plugin /bin/kube-scheduler
WORKDIR /bin
CMD ["kube-scheduler"]

LABEL com.redhat.component="noderesourcetopology-scheduler-container" \
      name="openshift4/noderesourcetopology-scheduler-container-rhel8" \
      summary="node resource topology aware scheduler" \
      io.openshift.expose-services="" \
      io.openshift.tags="numa,topology,scheduler" \
      io.k8s.display-name="noderesourcetopology-scheduler" \
      description="kubernetes scheduler aware of node resource topology." \
      maintainer="openshift-operators@redhat.com" \
      io.openshift.maintainer.component="Node Resource Topology aware Scheduler" \
      io.openshift.maintainer.product="OpenShift Container Platform" \
      io.k8s.description="Node Resource Topology aware Scheduler" \
      cpe="cpe:/a:redhat:openshift:4.12::el8"
