FROM registry.redhat.io/openshift/golang-builder:golang-builder-v1.19-rhel8@sha256:cdb6fe437dcbdab2f4d5faee06d67b39401f484d0234f9adcccbc5af37402cfb as builder

WORKDIR /app

COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -o bin/noderesourcetopology-plugin cmd/noderesourcetopology-plugin/main.go

FROM registry.redhat.io/ubi8/ubi-minimal:latest@sha256:bc7e42f085daa77b4cc51968ae9afffd0a96d802b710cdba6d75630200591e05

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
