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
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"

	authnv1 "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/k8stopologyawareschedwg/podfingerprint"

	"github.com/openshift-kni/debug-tools/pkg/pfpstatus/record"
)

const (
	PFPStatusDumpEnvVar string = "PFP_STATUS_DUMP"
	PFPStatusHostEnvVar string = "PFP_STATUS_HOST"
	PFPStatusPortEnvVar string = "PFP_STATUS_PORT"
)

const (
	DefaultHTTPServePort int    = 33445
	DefaultDumpDirectory string = "/run/pfpstatus"
)

const (
	defaultMaxNodes          = 5000
	defaultMaxSamplesPerNode = 10
	defaultDumpPeriod        = 10 * time.Second
)

type Middleware struct {
	Name string
	Link func(http.Handler) http.Handler
}

type HTTPParams struct {
	Enabled     bool
	Host        string
	Port        int
	Middlewares []Middleware
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
			Enabled: true,
			Host:    "", // all interfaces
			Port:    DefaultHTTPServePort,
		},
		Storage: StorageParams{
			Enabled:   false,
			Directory: DefaultDumpDirectory,
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

	// the port setting is deemed more important than the host setting;
	// we don't control the enable toggle from the host, by design.
	dumpHost, ok := os.LookupEnv(PFPStatusHostEnvVar)
	if ok && params.HTTP.Enabled {
		params.HTTP.Host = dumpHost
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /pfpstatus", func(w http.ResponseWriter, r *http.Request) {
		pfpstatusMainHandler(env, w, r)
	})
	mux.HandleFunc("GET /pfpstatus/{nodeName}", func(w http.ResponseWriter, r *http.Request) {
		pfpstatusNodeHandler(env, w, r)
	})

	var handle http.Handler = mux
	for _, mw := range params.Middlewares {
		env.lh.Info("Linking middleware", "name", mw.Name)
		handle = mw.Link(handle)
	}
	host := expandHost(params.Host)
	env.lh.Info("Starting PFP server", "host", host, "port", params.Port)
	addr := fmt.Sprintf("%s:%d", host, params.Port)
	err := http.ListenAndServe(addr, handle)
	if err != nil {
		env.lh.Error(err, "cannot serve PFP status")
	}
}

func expandHost(host string) string {
	if host == "" {
		return "0.0.0.0"
	}
	return host
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

type TokenBearerAuthenticator struct {
	cs   kubernetes.Interface
	lh   logr.Logger
	next http.Handler
}

func (tba *TokenBearerAuthenticator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tba.lh.V(6).Info("auth middleware start")
	defer tba.lh.V(6).Info("auth middleware stop")

	if isRequestFromLoopback(tba.lh, r.RemoteAddr) {
		tba.lh.V(2).Info("auth bypass for loopback request", "remoteAddr", r.RemoteAddr)
		tba.next.ServeHTTP(w, r)
		return
	}

	err, code := canTokenBearerDebugPFPs(context.Background(), tba.cs, tba.lh, r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	tba.next.ServeHTTP(w, r)
}
func (tba *TokenBearerAuthenticator) WithLink(next http.Handler) *TokenBearerAuthenticator {
	tba.next = next
	return tba
}

func (tba *TokenBearerAuthenticator) Link(next http.Handler) http.Handler {
	tba.next = next
	return tba
}

func NewTokenBearerAuth(cs kubernetes.Interface, lh logr.Logger) *TokenBearerAuthenticator {
	return &TokenBearerAuthenticator{
		cs: cs,
		lh: lh,
	}
}

func isRequestFromLoopback(lh logr.Logger, remoteAddr string) bool {
	// Get the remote host IP, splitting off the port
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// this is so unexpected that we want to be loud
		lh.Error(err, "cannot parse address", "remoteAddr", remoteAddr)
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func canTokenBearerDebugPFPs(ctx context.Context, cs kubernetes.Interface, lh logr.Logger, authHeader string) (error, int) {
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return errors.New("Unauthorized: Missing Authorization Header"), http.StatusUnauthorized
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	tr, err := cs.AuthenticationV1().TokenReviews().Create(ctx, &authnv1.TokenReview{
		Spec: authnv1.TokenReviewSpec{
			Token: token,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		lh.V(2).Error(err, "performing TokenReview")
		return fmt.Errorf("Internal Server Error: %w", err), http.StatusInternalServerError
	}

	if !tr.Status.Authenticated {
		lh.V(2).Info("token authentication failed")
		return errors.New("Unauthorized: Invalid token"), http.StatusUnauthorized
	}

	lh.V(4).Info("token authenticated", "user", tr.Status.User.Username)

	sar := &authzv1.SubjectAccessReview{
		Spec: authzv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authzv1.ResourceAttributes{
				Group:    "topology.node.k8s.io",
				Resource: "noderesourcetopologies",
				Verb:     "sched.openshift-kni.io/debug",
			},
			User:   tr.Status.User.Username,
			Groups: tr.Status.User.Groups,
			Extra:  convertExtra(tr.Status.User.Extra),
		},
	}

	sarResult, err := cs.AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		lh.V(2).Error(err, "checking SubjectAccessReview")
		return fmt.Errorf("Internal Server Error: %w", err), http.StatusInternalServerError
	}

	if !sarResult.Status.Allowed {
		lh.V(2).Info("request denied", "username", tr.Status.User.Username, "reason", sarResult.Status.Reason)
		return errors.New("Forbidden: " + sarResult.Status.Reason), http.StatusForbidden
	}

	lh.V(4).Info("request allowed", "username", tr.Status.User.Username)
	return nil, http.StatusOK
}

func convertExtra(extra map[string]authnv1.ExtraValue) map[string]authzv1.ExtraValue {
	if extra == nil {
		return nil
	}
	newExtra := make(map[string]authzv1.ExtraValue)
	for k, v := range extra {
		newExtra[k] = authzv1.ExtraValue(v)
	}
	return newExtra
}
