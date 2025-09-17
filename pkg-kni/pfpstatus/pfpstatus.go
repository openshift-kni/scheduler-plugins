/*
 * Copyright 2023 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pfpstatus

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"

	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/k8stopologyawareschedwg/podfingerprint"
	"github.com/k8stopologyawareschedwg/podfingerprint/record"
)

const (
	PFPStatusDumpEnvVar string = "PFP_STATUS_DUMP"
	PFPStatusPortEnvVar string = "PFP_STATUS_PORT"
)

const (
	defaultMaxNodes          = 5000
	defaultMaxSamplesPerNode = 10
	defaultDumpPeriod        = 10 * time.Second
)

type HTTPParams struct {
	Enabled     bool
	Port        int
	RequireAuth bool
}

type StorageParams struct {
	Enabled   bool
	Directory string
	Period    time.Duration
}

type Params struct {
	HTTP    HTTPParams
	Storage StorageParams
}

type environ struct {
	mu  sync.Mutex
	rec *record.Recorder
	lh  logr.Logger
	cs  kubernetes.Interface
}

func DefaultParams() Params {
	return Params{
		HTTP: HTTPParams{
			Enabled:     true,
			Port:        33445,
			RequireAuth: true,
		},
		Storage: StorageParams{
			Enabled:   false,
			Directory: "/run/pfpstatus",
			Period:    10 * time.Second,
		},
	}
}

func ParamsFromEnv(lh logr.Logger, params *Params) {
	dumpDir, ok := os.LookupEnv(PFPStatusDumpEnvVar)
	if !ok || dumpDir == "" {
		params.Storage.Enabled = false
	} else {
		params.Storage.Enabled = true
		params.Storage.Directory = dumpDir
	}

	// let's try to keep the amount of code we do in init() at minimum.
	// This may happen if the container didn't have the directory mounted
	if !existsBaseDirectory(dumpDir) {
		lh.Info("base directory not found, will discard everything", "baseDirectory", dumpDir)
		params.Storage.Enabled = false
	}

	dumpPort, ok := os.LookupEnv(PFPStatusPortEnvVar)
	if !ok || dumpPort == "" {
		params.HTTP.Enabled = false
	} else {
		port, err := strconv.Atoi(dumpPort)
		if err != nil {
			lh.Error(err, "parsing dump port %q", dumpPort)
			params.HTTP.Enabled = false
		} else {
			params.HTTP.Enabled = true
			params.HTTP.Port = port
		}
	}
}

func Setup(logh logr.Logger, params Params) {
	if !params.Storage.Enabled && !params.HTTP.Enabled {
		logh.Info("no backend enabled, nothing to do")
		return
	}

	logh.Info("Setup in progress", "params", fmt.Sprintf("%+#v", params))

	rec, err := record.NewRecorder(defaultMaxNodes, defaultMaxSamplesPerNode, time.Now)
	if err != nil {
		logh.Error(err, "cannot create a status recorder")
		return
	}

	ctx := context.Background()
	env := environ{
		rec: rec,
		lh:  logh,
	}

	ch := make(chan podfingerprint.Status)
	podfingerprint.SetCompletionSink(ch)
	go collectLoop(ctx, &env, ch)
	if params.Storage.Enabled {
		go dumpLoop(ctx, &env, params.Storage)
	}
	if params.HTTP.Enabled {
		go serveLoop(ctx, &env, params.HTTP)
	}
}

func collectLoop(ctx context.Context, env *environ, updates <-chan podfingerprint.Status) {
	env.lh.V(4).Info("collect loop started")
	defer env.lh.V(4).Info("collect loop finished")
	for {
		select {
		case <-ctx.Done():
			return
		case st := <-updates:
			env.mu.Lock()
			_ = env.rec.Push(st) // intentionally ignore error
			env.mu.Unlock()
		}
	}
}

func dumpLoop(ctx context.Context, env *environ, params StorageParams) {
	env.lh.V(4).Info("dump loop started")
	defer env.lh.V(4).Info("dump loop finished")
	ticker := time.NewTicker(params.Period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			env.mu.Lock()
			snapshot := env.rec.Content()
			env.mu.Unlock()

			for nodeName, statuses := range snapshot {
				record.DumpToFile(params.Directory, nodeName, statuses)
			}
		}
	}
}

func serveLoop(ctx context.Context, env *environ, params HTTPParams) {
	var err error
	var cs kubernetes.Interface
	if params.RequireAuth {
		cs, err = createClient()
		if err != nil {
			env.lh.Error(err, "cannot create the authentication client")
			return
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /pfpstatus", func(w http.ResponseWriter, r *http.Request) {
		pfpstatusMainHandler(env, w, r)
	})
	mux.HandleFunc("GET /pfpstatus/{nodeName}", func(w http.ResponseWriter, r *http.Request) {
		pfpstatusNodeHandler(env, w, r)
	})
	env.lh.Info("Starting PFP server", "port", params.Port, "auth", params.RequireAuth)
	var handle http.Handler = mux
	if params.RequireAuth {
		handle = authMiddleware(mux, cs, env.lh)
	}
	err = http.ListenAndServe(fmt.Sprintf(":%d", params.Port), handle)
	if err != nil {
		env.lh.Error(err, "cannot serve PFP status")
	}
}

func pfpstatusMainHandler(env *environ, w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Nodes int `json:"nodes"`
	}
	env.mu.Lock()
	resp := Response{
		Nodes: env.rec.CountNodes(),
	}
	env.mu.Unlock()

	sendContent(env.lh, w, &resp, "generic")
}

func pfpstatusNodeHandler(env *environ, w http.ResponseWriter, r *http.Request) {
	nodeName := r.PathValue("nodeName")
	if nodeName == "" {
		env.lh.Info("requested pfpstatus for empty node")
		http.Error(w, "missing node name", http.StatusUnprocessableEntity)
		return
	}
	env.mu.Lock()
	content, ok := env.rec.ContentForNode(nodeName)
	env.mu.Unlock()

	if !ok {
		http.Error(w, "unknown node name", http.StatusUnprocessableEntity)
		return
	}
	sendContent(env.lh, w, content, nodeName)
}

func sendContent(lh logr.Logger, w http.ResponseWriter, content any, endpoint string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(content)
	if err != nil {
		lh.Error(err, "sending back content for endpoint %q", endpoint)
	}
}

func existsBaseDirectory(baseDir string) bool {
	info, err := os.Stat(baseDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isRequestFromLoopback(r *http.Request) bool {
	// Get the remote host IP, splitting off the port
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// TODO: log
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func authMiddleware(next http.Handler, cs kubernetes.Interface, lh logr.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isRequestFromLoopback(r) {
			lh.V(2).Info("auth bypass for loopback request", "remoteAddr", r.RemoteAddr)
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Forbidden: Missing Authorization Header", http.StatusForbidden)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		sar := &authv1.SubjectAccessReview{
			Spec: authv1.SubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Group:    "sched.openshift-kni.io", // Must match the apiGroup in the ClusterRole
					Resource: "pfpstatus",              // Must match the resource in the ClusterRole
					Verb:     "get",                    // Must match the verb in the ClusterRole
				},
				Extra: map[string]authv1.ExtraValue{"token": {token}},
			},
		}

		sar, err := cs.AuthorizationV1().SubjectAccessReviews().Create(context.Background(), sar, metav1.CreateOptions{})
		if err != nil {
			lh.V(4).Error(err, "checking SubjectAccessReview")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !sar.Status.Allowed {
			lh.V(4).Info("request denied for token", "reason", sar.Status.Reason)
			http.Error(w, fmt.Sprintf("Forbidden: %s", sar.Status.Reason), http.StatusForbidden)
			return
		}

		lh.V(4).Info("request allowed", "reason", sar.Status.Reason)
		next.ServeHTTP(w, r)
	})
}

func createClient() (kubernetes.Interface, error) {
	var err error
	var kubeconfig *rest.Config

	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath != "" {
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		kubeconfig, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("unable to load kubeconfig (path=%s)", kubeconfigPath)
	}

	return kubernetes.NewForConfig(kubeconfig)
}
