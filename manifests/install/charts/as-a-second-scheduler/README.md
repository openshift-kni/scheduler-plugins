# Scheduler-plugins as a second scheduler in cluster

## Table of Contents

<!-- toc -->
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Installing the chart](#installing-the-chart)
    - [Install chart using Helm v3.0+](#install-chart-using-helm-v30)
    - [Verify that scheduler and plugin-controller pod are running properly.](#verify-that-scheduler-and-plugin-controller-pod-are-running-properly)
  - [Configuration](#configuration)
<!-- /toc -->

## Installation

Quick start instructions for the setup and configuration of as-a-second-scheduler using Helm.

### Prerequisites

- [Helm](https://helm.sh/docs/intro/quickstart/#install-helm)

### Installing the chart

#### Install chart using Helm v3.0+

> 🆕 Starting v0.28, Helm charts are hosted on https://scheduler-plugins.sigs.k8s.io

```bash
$ git clone git@github.com:kubernetes-sigs/scheduler-plugins.git
$ cd scheduler-plugins/manifests/install/charts
$ helm install --repo https://scheduler-plugins.sigs.k8s.io scheduler-plugins scheduler-plugins
```

#### Verify that scheduler and plugin-controller pod are running properly.

```bash
$ kubectl get deploy -n scheduler-plugins
NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
scheduler-plugins-controller   1/1     1            1           7s
scheduler-plugins-scheduler    1/1     1            1           7s
```

### Configuration

The following table lists the configurable parameters of the as-a-second-scheduler chart and their default values.

| Parameter                      | Description                  | Default                                                                                         |
|--------------------------------|------------------------------|-------------------------------------------------------------------------------------------------|
| `scheduler.name`               | Scheduler name               | `scheduler-plugins-scheduler`                                                                   |
| `scheduler.image`              | Scheduler image              | `registry.k8s.io/scheduler-plugins/kube-scheduler:v0.31.8`                                      |
| `scheduler.command`            | Scheduler command            | `["/bin/kube-scheduler"]`                                                                       |
| `scheduler.leaderElect`        | Scheduler leaderElection     | `false`                                                                                         |
| `scheduler.replicaCount`       | Scheduler replicaCount       | `1`                                                                                             |
| `scheduler.priorityClassName`  | Scheduler priorityClassName  | `""`                                                                                            |
| `scheduler.resources`          | Scheduler resources          | `{}`                                                                                            |
| `scheduler.nodeSelector`       | Scheduler nodeSelector       | `{}`                                                                                            |
| `scheduler.affinity`           | Scheduler affinity           | `{}`                                                                                            |
| `scheduler.tolerations`        | Scheduler tolerations        | `[]`                                                                                            |
| `controller.name`              | Controller name              | `scheduler-plugins-controller`                                                                  |
| `controller.image`             | Controller image             | `registry.k8s.io/scheduler-plugins/controller:v0.29.7`                                          |
| `controller.replicaCount`      | Controller replicaCount      | `1`                                                                                             |
| `controller.priorityClassName` | Controller priorityClassName | `""`                                                                                            |
| `controller.resources`         | Controller resources         | `{}`                                                                                            |
| `controller.nodeSelector`      | Controller nodeSelector      | `{}`                                                                                            |
| `controller.affinity`          | Controller affinity          | `{}`                                                                                            |
| `controller.tolerations`       | Controller tolerations       | `[]`                                                                                            |
| `plugins.enabled`              | Plugins enabled by default   | `["Coscheduling","CapacityScheduling","NodeResourceTopologyMatch", "NodeResourcesAllocatable"]` |
| `plugins.disabled`             | Plugins disabled by default  | `["PrioritySort"]`                                                                              |
