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
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/go-logr/logr"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2/klogr"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"sigs.k8s.io/scheduler-plugins/pkg-kni/knidebug"
	"sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology"

	// Ensure scheme package is initialized.
	_ "sigs.k8s.io/scheduler-plugins/apis/config/scheme"

	knifeatures "sigs.k8s.io/scheduler-plugins/pkg-kni/features"

	"github.com/openshift-kni/debug-tools/pkg/k8sclient"
	"github.com/openshift-kni/debug-tools/pkg/pfpstatus"
)

func main() {
	logh := klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog))
	printVersion(logh) // this must be the first thing logged ever. Note: we can't use V() yet - no flags parsed

	utilfeature.DefaultMutableFeatureGate.SetFromMap(knifeatures.Desired())

	rand.Seed(time.Now().UnixNano())

	if err := setupPFPStatus(logh); err != nil {
		logh.Error(err, "failed to setup PFP Status repoorting")
		os.Exit(1)
	}

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

	if err := command.Execute(); err != nil {
		logh.Error(err, "failed to execute the scheduler command")
		os.Exit(1)
	}
}

func printVersion(logh logr.Logger) {
	ver := version.Get()
	logh.Info("starting noderesourcetopology scheduler", "version", fmt.Sprintf("%s.%s", ver.Major, ver.Minor), "gitcommit", ver.GitCommit, "goversion", ver.GoVersion, "platform", ver.Platform)
}

func setupPFPStatus(logh logr.Logger) error {
	params := pfpstatus.DefaultParams()
	// TODO: uncomment once ready
	// pfpstatus.ParamsFromEnv(logh, &params)
	if params.HTTP.Enabled {
		cs, err := k8sclient.Create()
		if err != nil {
			return err
		}
		tba := pfpstatus.NewTokenBearerAuth(cs, logh)
		params.HTTP.Middlewares = append(params.HTTP.Middlewares, pfpstatus.Middleware{
			Name: "TokenAuth",
			Link: tba.Link,
		})
	}
	pfpstatus.Setup(logh, params)
	return nil
}
