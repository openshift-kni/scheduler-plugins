hack/install-envtest.sh
hack/unit-test.sh 
+++ [0415 19:45:14] Configuring envtest
?   	sigs.k8s.io/scheduler-plugins/cmd/controller	[no test files]
?   	sigs.k8s.io/scheduler-plugins/cmd/controller/app	[no test files]
FAIL	sigs.k8s.io/scheduler-plugins/cmd/scheduler [build failed]
FAIL	sigs.k8s.io/scheduler-plugins/pkg/capacityscheduling [build failed]
FAIL	sigs.k8s.io/scheduler-plugins/pkg/coscheduling [build failed]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/applyconfiguration	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/applyconfiguration/scheduling/v1alpha1	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/applyconfiguration/internal	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/clientset/versioned	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/clientset/versioned/fake	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/clientset/versioned/scheme	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/clientset/versioned/typed/scheduling/v1alpha1	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/clientset/versioned/typed/scheduling/v1alpha1/fake	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/informers/externalversions	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/informers/externalversions/scheduling	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/informers/externalversions/scheduling/v1alpha1	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/listers/scheduling/v1alpha1	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/generated/informers/externalversions/internalinterfaces	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/networkaware/util	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/logging	[no test files]
?   	sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/podprovider	[no test files]
W0415 19:45:16.052201  248841 registry.go:345] setting componentGlobalsRegistry in SetFallback. We recommend calling componentGlobalsRegistry.Set() right after parsing flags to avoid using feature gates before their final values are set by the flags.
I0415 19:45:16.320908  248841 serving.go:386] Generated self-signed cert in-memory
W0415 19:45:16.495281  248841 authentication.go:368] No authentication-kubeconfig provided in order to lookup client-ca-file in configmap/extension-apiserver-authentication in kube-system, so client certificate authentication won't work.
W0415 19:45:16.495300  248841 authentication.go:392] No authentication-kubeconfig provided in order to lookup requestheader-client-ca-file in configmap/extension-apiserver-authentication in kube-system, so request-header client certificate authentication won't work.
W0415 19:45:16.495311  248841 authorization.go:193] No authorization-kubeconfig provided, so SubjectAccessReview of authorization tokens won't work.
W0415 19:45:16.499671  248841 registry.go:256] calling componentGlobalsRegistry.AddFlags more than once, the registry will be set by the latest flags
I0415 19:45:16.875937  248841 serving.go:386] Generated self-signed cert in-memory
W0415 19:45:17.235623  248841 authentication.go:368] No authentication-kubeconfig provided in order to lookup client-ca-file in configmap/extension-apiserver-authentication in kube-system, so client certificate authentication won't work.
W0415 19:45:17.235638  248841 authentication.go:392] No authentication-kubeconfig provided in order to lookup requestheader-client-ca-file in configmap/extension-apiserver-authentication in kube-system, so request-header client certificate authentication won't work.
W0415 19:45:17.235647  248841 authorization.go:193] No authorization-kubeconfig provided, so SubjectAccessReview of authorization tokens won't work.
--- FAIL: TestSetup (1.19s)
    --- FAIL: TestSetup/single_profile_config_-_NodeResourceTopologyMatch_with_args (0.74s)
        main_test.go:209: unexpected plugins diff (-want, +got):   map[string]*config.Plugins{
              	"default-scheduler": &{
              		PreEnqueue: config.PluginSet{
              			Enabled: []config.Plugin{
              				{Name: "SchedulingGates"},
            + 				{Name: "DefaultPreemption"},
              			},
              			Disabled: nil,
              		},
              		QueueSort: {Enabled: {{Name: "PrioritySort"}}},
              		PreFilter: {Enabled: {{Name: "NodeAffinity"}, {Name: "NodePorts"}, {Name: "NodeResourcesFit"}, {Name: "VolumeRestrictions"}, ...}},
              		... // 10 identical fields
              	},
              }
FAIL
FAIL	sigs.k8s.io/scheduler-plugins/cmd/noderesourcetopology-plugin	1.219s
ok  	sigs.k8s.io/scheduler-plugins/pkg/controllers	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/coscheduling/core	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/crossnodepreemption	(cached) [no tests to run]
ok  	sigs.k8s.io/scheduler-plugins/pkg/networkaware/networkoverhead	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/networkaware/topologicalsort	(cached)
--- FAIL: TestNodeResourcesAllocatable (0.00s)
    --- FAIL: TestNodeResourcesAllocatable/nothing_scheduled,_nothing_requested (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0xcc8cef]

goroutine 31 [running]:
testing.tRunner.func1.2({0x220ab20, 0x3faaf40})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1632 +0x230
testing.tRunner.func1()
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1635 +0x35e
panic({0x220ab20?, 0x3faaf40?})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/runtime/panic.go:785 +0x132
k8s.io/component-base/metrics.(*CounterVec).WithLabelValues(0x0, {0xc000557840, 0x3, 0x3})
	/home/shajmakh/go/pkg/mod/k8s.io/component-base@v0.32.0/metrics/counter.go:212 +0x2f
k8s.io/kubernetes/pkg/scheduler/framework/runtime.(*frameworkImpl).setInstrumentedPlugins(0xc0001f0008)
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:430 +0x595
k8s.io/kubernetes/pkg/scheduler/framework/runtime.NewFramework({0x2920d60, 0xc0004f00a0}, 0xc00055bd70, 0xc00053af00, {0xc000557da8, 0x3, 0x410d25?})
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:400 +0x11f3
k8s.io/kubernetes/pkg/scheduler/testing/framework.NewFramework({0x2920d60, 0xc0004f00a0}, {0xc000557e28, 0x3, 0x3f498c0?}, {0x25e2bfe, 0x11}, {0xc000557da8, 0x3, 0x3})
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/testing/framework/framework_helpers.go:43 +0x15e
sigs.k8s.io/scheduler-plugins/pkg/noderesources.TestNodeResourcesAllocatable.func1(0xc0002741a0)
	/home/shajmakh/ghrepo/scheduler-plugins/pkg/noderesources/allocatable_test.go:255 +0x4e5
testing.tRunner(0xc0002741a0, 0xc0006820d0)
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1690 +0xf4
created by testing.(*T).Run in goroutine 12
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1743 +0x390
FAIL	sigs.k8s.io/scheduler-plugins/pkg/noderesources	0.029s
ok  	sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/cache	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/nodeconfig	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/resourcerequests	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/stringify	(cached)
FAIL	sigs.k8s.io/scheduler-plugins/pkg/preemptiontoleration [build failed]
--- FAIL: TestPodState (0.00s)
    --- FAIL: TestPodState/node_has_more_terminating_pods_will_be_scored_with_higher_score,_node_has_regular_pods_only_will_be_scored_with_the_lowest_score. (0.00s)
        framework.go:399: I0415 19:45:17.371758] the scheduler starts to work with those plugins Plugins={"PreEnqueue":{"Enabled":null,"Disabled":null},"QueueSort":{"Enabled":[{"Name":"PrioritySort","Weight":0}],"Disabled":null},"PreFilter":{"Enabled":null,"Disabled":null},"Filter":{"Enabled":null,"Disabled":null},"PostFilter":{"Enabled":null,"Disabled":null},"PreScore":{"Enabled":null,"Disabled":null},"Score":{"Enabled":[{"Name":"PodState","Weight":1}],"Disabled":null},"Reserve":{"Enabled":null,"Disabled":null},"Permit":{"Enabled":null,"Disabled":null},"PreBind":{"Enabled":null,"Disabled":null},"Bind":{"Enabled":[{"Name":"DefaultBinder","Weight":0}],"Disabled":null},"PostBind":{"Enabled":null,"Disabled":null},"MultiPoint":{"Enabled":null,"Disabled":null}}
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0xcc27af]

goroutine 26 [running]:
testing.tRunner.func1.2({0x2230760, 0x4032290})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1632 +0x230
testing.tRunner.func1()
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1635 +0x35e
panic({0x2230760?, 0x4032290?})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/runtime/panic.go:785 +0x132
k8s.io/component-base/metrics.(*CounterVec).WithLabelValues(0x0, {0xc000113808, 0x3, 0x3})
	/home/shajmakh/go/pkg/mod/k8s.io/component-base@v0.32.0/metrics/counter.go:212 +0x2f
k8s.io/kubernetes/pkg/scheduler/framework/runtime.(*frameworkImpl).setInstrumentedPlugins(0xc0004ad688)
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:430 +0x595
k8s.io/kubernetes/pkg/scheduler/framework/runtime.NewFramework({0x29828d8, 0xc0004deb70}, 0xc0004ded20, 0xc000369bc0, {0xc0004dbf40, 0x4, 0x410d25?})
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:400 +0x11f3
k8s.io/kubernetes/pkg/scheduler/testing/framework.NewFramework({0x29828d8, 0xc0004deb70}, {0xc0004dbe08, 0x3, 0x478b48?}, {0x2610040, 0x11}, {0xc0004dbf40, 0x4, 0x4})
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/testing/framework/framework_helpers.go:43 +0x15e
sigs.k8s.io/scheduler-plugins/pkg/podstate.TestPodState.func1(0xc000279380)
	/home/shajmakh/ghrepo/scheduler-plugins/pkg/podstate/pod_state_test.go:88 +0x4bf
testing.tRunner(0xc000279380, 0xc0005ce120)
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1690 +0xf4
created by testing.(*T).Run in goroutine 25
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1743 +0x390
FAIL	sigs.k8s.io/scheduler-plugins/pkg/podstate	0.058s
ok  	sigs.k8s.io/scheduler-plugins/pkg/qos	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/sysched	(cached)
ok  	sigs.k8s.io/scheduler-plugins/pkg/trimaran	(cached)
E0415 19:45:17.522567  249179 analysis.go:36] "Invalid resource capacity" capacity=0
--- FAIL: TestNew (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0xd302ef]

goroutine 93 [running]:
testing.tRunner.func1.2({0x267d380, 0x4911160})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1632 +0x230
testing.tRunner.func1()
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1635 +0x35e
panic({0x267d380?, 0x4911160?})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/runtime/panic.go:785 +0x132
k8s.io/component-base/metrics.(*CounterVec).WithLabelValues(0x0, {0xc0004f7860, 0x3, 0x3})
	/home/shajmakh/go/pkg/mod/k8s.io/component-base@v0.32.0/metrics/counter.go:212 +0x2f
k8s.io/kubernetes/pkg/scheduler/framework/runtime.(*frameworkImpl).setInstrumentedPlugins(0xc0000e58c8)
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:430 +0x595
k8s.io/kubernetes/pkg/scheduler/framework/runtime.NewFramework({0x2ed2890, 0xc000462140}, 0xc0005c6ea0, 0xc00029a280, {0xc0004f7ec0, 0x3, 0x410d25?})
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:400 +0x11f3
sigs.k8s.io/scheduler-plugins/test/util.NewFramework({0x2ed2890, 0xc000462140}, {0xc0004f7f28, 0x3, 0x54ce6d?}, {0xc00079e8a0, 0x1, 0x1}, {0x2b14173, 0x11}, ...)
	/home/shajmakh/ghrepo/scheduler-plugins/test/util/framework.go:41 +0x1ad
sigs.k8s.io/scheduler-plugins/pkg/trimaran/loadvariationriskbalancing.TestNew(0xc0000ffd40)
	/home/shajmakh/ghrepo/scheduler-plugins/pkg/trimaran/loadvariationriskbalancing/loadvariationriskbalancing_test.go:109 +0x5f5
testing.tRunner(0xc0000ffd40, 0x2c54228)
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1690 +0xf4
created by testing.(*T).Run in goroutine 1
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1743 +0x390
FAIL	sigs.k8s.io/scheduler-plugins/pkg/trimaran/loadvariationriskbalancing	0.028s
--- FAIL: TestLowRiskOverCommitment_New (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0xd9826f]

goroutine 87 [running]:
testing.tRunner.func1.2({0x2681640, 0x491b540})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1632 +0x230
testing.tRunner.func1()
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1635 +0x35e
panic({0x2681640?, 0x491b540?})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/runtime/panic.go:785 +0x132
k8s.io/component-base/metrics.(*CounterVec).WithLabelValues(0x0, {0xc00014d7d8, 0x3, 0x3})
	/home/shajmakh/go/pkg/mod/k8s.io/component-base@v0.32.0/metrics/counter.go:212 +0x2f
k8s.io/kubernetes/pkg/scheduler/framework/runtime.(*frameworkImpl).setInstrumentedPlugins(0xc0002dd688)
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:424 +0x405
k8s.io/kubernetes/pkg/scheduler/framework/runtime.NewFramework({0x2ed8970, 0xc000461130}, 0xc00080cb70, 0xc0006d0880, {0xc00014dd70, 0x3, 0x410d25?})
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:400 +0x11f3
sigs.k8s.io/scheduler-plugins/test/util.NewFramework({0x2ed8970, 0xc000461130}, {0xc00014dde0, 0x4, 0x39b970c?}, {0xc00062f5e0, 0x1, 0x1}, {0x2b19565, 0x11}, ...)
	/home/shajmakh/ghrepo/scheduler-plugins/test/util/framework.go:41 +0x1ad
sigs.k8s.io/scheduler-plugins/pkg/trimaran/lowriskovercommitment.TestLowRiskOverCommitment_New(0xc00002d040)
	/home/shajmakh/ghrepo/scheduler-plugins/pkg/trimaran/lowriskovercommitment/lowriskovercommitment_test.go:112 +0x74e
testing.tRunner(0xc00002d040, 0x2c59970)
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1690 +0xf4
created by testing.(*T).Run in goroutine 1
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1743 +0x390
FAIL	sigs.k8s.io/scheduler-plugins/pkg/trimaran/lowriskovercommitment	0.027s
?   	sigs.k8s.io/scheduler-plugins/apis/config	[no test files]
?   	sigs.k8s.io/scheduler-plugins/apis/scheduling	[no test files]
?   	sigs.k8s.io/scheduler-plugins/apis/scheduling/scheme	[no test files]
?   	sigs.k8s.io/scheduler-plugins/apis/scheduling/v1alpha1	[no test files]
args power model : map[]
--- FAIL: TestPeaksNew (0.01s)
    utils.go:190: Power model json data: map[node-1:map[k0:471.7412504314313 k1:-91.50493019588365 k2:-0.07186049052516227]]
    utils.go:200: NODE_POWER_MODEL: ./power_model/node_power_model
    utils.go:202: fileDir: ./power_model/ fileName: node_power_model
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0xdb1d4f]

goroutine 198 [running]:
testing.tRunner.func1.2({0x2b5f6a0, 0x520e230})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1632 +0x230
testing.tRunner.func1()
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1635 +0x35e
panic({0x2b5f6a0?, 0x520e230?})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/runtime/panic.go:785 +0x132
k8s.io/component-base/metrics.(*CounterVec).WithLabelValues(0x0, {0xc0007ef7c8, 0x3, 0x3})
	/home/shajmakh/go/pkg/mod/k8s.io/component-base@v0.32.0/metrics/counter.go:212 +0x2f
k8s.io/kubernetes/pkg/scheduler/framework/runtime.(*frameworkImpl).setInstrumentedPlugins(0xc00081cd88)
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:430 +0x595
k8s.io/kubernetes/pkg/scheduler/framework/runtime.NewFramework({0x34c63a8, 0xc0001a5d60}, 0xc000a06780, 0xc000688240, {0xc0007efd68, 0x3, 0x4110e5?})
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:400 +0x11f3
sigs.k8s.io/scheduler-plugins/test/util.NewFramework({0x34c63a8, 0xc0001a5d60}, {0xc0007efe20, 0x3, 0x6f4b9?}, {0xc0006570c0, 0x1, 0x1}, {0x30868f1, 0xe}, ...)
	/home/shajmakh/ghrepo/scheduler-plugins/test/util/framework.go:41 +0x1ad
sigs.k8s.io/scheduler-plugins/pkg/trimaran/peaks.TestPeaksNew(0xc00038f520)
	/home/shajmakh/ghrepo/scheduler-plugins/pkg/trimaran/peaks/peaks_test.go:119 +0x628
testing.tRunner(0xc00038f520, 0x31e5168)
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1690 +0xf4
created by testing.(*T).Run in goroutine 1
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1743 +0x390
FAIL	sigs.k8s.io/scheduler-plugins/pkg/trimaran/peaks	0.047s
E0415 19:45:17.891721  249219 collector.go:136] "Load watcher client failed" err="Get \"http://deadbeef:2020/watcher\": dial tcp: lookup deadbeef on 10.47.242.10:53: no such host"
E0415 19:45:17.891821  249219 collector.go:79] "Unable to populate metrics initially" err="Get \"http://deadbeef:2020/watcher\": dial tcp: lookup deadbeef on 10.47.242.10:53: no such host"
--- FAIL: TestNew (0.06s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0xd9766f]

goroutine 14 [running]:
testing.tRunner.func1.2({0x267c280, 0x49101a0})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1632 +0x230
testing.tRunner.func1()
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1635 +0x35e
panic({0x267c280?, 0x49101a0?})
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/runtime/panic.go:785 +0x132
k8s.io/component-base/metrics.(*CounterVec).WithLabelValues(0x0, {0xc000807858, 0x3, 0x3})
	/home/shajmakh/go/pkg/mod/k8s.io/component-base@v0.32.0/metrics/counter.go:212 +0x2f
k8s.io/kubernetes/pkg/scheduler/framework/runtime.(*frameworkImpl).setInstrumentedPlugins(0xc0003566c8)
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:430 +0x595
k8s.io/kubernetes/pkg/scheduler/framework/runtime.NewFramework({0x2ed17d8, 0xc00061c640}, 0xc00061e270, 0xc000b0c800, {0xc000631f38, 0x3, 0x410d25?})
	/home/shajmakh/go/pkg/mod/k8s.io/kubernetes@v1.32.0/pkg/scheduler/framework/runtime/framework.go:400 +0x11f3
sigs.k8s.io/scheduler-plugins/test/util.NewFramework({0x2ed17d8, 0xc00061c640}, {0xc000631da8, 0x3, 0x2fa4645?}, {0xc0005ccd40, 0x1, 0x1}, {0x2b0dfeb, 0xe}, ...)
	/home/shajmakh/ghrepo/scheduler-plugins/test/util/framework.go:41 +0x1ad
sigs.k8s.io/scheduler-plugins/pkg/trimaran/targetloadpacking.TestNew(0xc0004dcb60)
	/home/shajmakh/ghrepo/scheduler-plugins/pkg/trimaran/targetloadpacking/targetloadpacking_test.go:105 +0x525
testing.tRunner(0xc0004dcb60, 0x2c53260)
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1690 +0xf4
created by testing.(*T).Run in goroutine 1
	/home/shajmakh/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.23.0.linux-amd64/src/testing/testing.go:1743 +0x390
FAIL	sigs.k8s.io/scheduler-plugins/pkg/trimaran/targetloadpacking	0.082s
ok  	sigs.k8s.io/scheduler-plugins/pkg/util	(cached)
ok  	sigs.k8s.io/scheduler-plugins/apis/config/scheme	(cached)
ok  	sigs.k8s.io/scheduler-plugins/apis/config/v1	(cached)
ok  	sigs.k8s.io/scheduler-plugins/apis/config/validation	(cached)
FAIL
