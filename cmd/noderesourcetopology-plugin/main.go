/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"sigs.k8s.io/scheduler-plugins/pkg-kni/knidebug"
	"sigs.k8s.io/scheduler-plugins/pkg-kni/pfpstatus"
	"sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology"

	// Ensure scheme package is initialized.
	_ "sigs.k8s.io/scheduler-plugins/apis/config/scheme"

	knifeatures "sigs.k8s.io/scheduler-plugins/pkg-kni/features"
	kniinformer "sigs.k8s.io/scheduler-plugins/pkg-kni/podinformer"

	"github.com/k8stopologyawareschedwg/podfingerprint"
)

const (
	PFPStatusDumpEnvVar string = "PFP_STATUS_DUMP"
)

func setupPFPStatusDump() {
	dumpDir, ok := os.LookupEnv(PFPStatusDumpEnvVar)
	if !ok || dumpDir == "" {
		klog.InfoS("PFP Status dump disabled", "variableFound", ok, "valueGiven", dumpDir != "")
		return
	}

	klog.InfoS("PFP Status dump enabled", "statusDirectory", dumpDir)

	ch := make(chan podfingerprint.Status)
	logh := klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog))

	podfingerprint.SetCompletionSink(ch)
	go pfpstatus.RunForever(context.Background(), logh, dumpDir, ch)
}

func main() {
	utilfeature.DefaultMutableFeatureGate.SetFromMap(knifeatures.Desired())

	rand.Seed(time.Now().UnixNano())

	kniinformer.Setup()

	// Register custom plugins to the scheduler framework.
	// Later they can consist of scheduler profile(s) and hence
	// used by various kinds of workloads.
	command := app.NewSchedulerCommand(
		app.WithPlugin(noderesourcetopology.Name, noderesourcetopology.New),
		app.WithPlugin(knidebug.Name, knidebug.New),
	)

	// TODO: once we switch everything over to Cobra commands, we can go back to calling
	// utilflag.InitFlags() (by removing its pflag.Parse() call). For now, we have to set the
	// normalize func and add the go flag set by hand.
	// utilflag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	setupPFPStatusDump()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
