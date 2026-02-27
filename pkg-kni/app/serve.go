/*
 * Copyright 2026 Red Hat, Inc.
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

package app

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"k8s.io/apiserver/pkg/endpoints/metrics"
	apiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/klog/v2"

	configclientset "github.com/openshift/client-go/config/clientset/versioned"
)

// serveFunc is the callback signature used by run() to start the HTTPS server.
// It replaces the direct cc.SecureServing.Serve() call, giving us full control
// over the *tls.Config without modifying any vendored code.
type serveFunc func(
	s *apiserver.SecureServingInfo,
	configClient configclientset.Interface,
	handler http.Handler,
	shutdownTimeout time.Duration,
	stopCh <-chan struct{},
) (<-chan struct{}, <-chan struct{}, error)

// customServe is our serveFunc implementation. It builds an http.Server with a
// custom *tls.Config (from buildCustomTLSConfig) and delegates the
// goroutine/shutdown lifecycle to the exported server.RunServer().
func customServe(
	s *apiserver.SecureServingInfo,
	configClient configclientset.Interface,
	handler http.Handler,
	shutdownTimeout time.Duration,
	stopCh <-chan struct{},
) (<-chan struct{}, <-chan struct{}, error) {
	if s.Listener == nil {
		return nil, nil, fmt.Errorf("listener must not be nil")
	}

	tlsConfig, err := buildCustomTLSConfig(configClient)
	if err != nil {
		return nil, nil, err
	}

	secureServer := &http.Server{
		Addr:           s.Listener.Addr().String(),
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tlsConfig,

		IdleTimeout:       90 * time.Second, // matches http.DefaultTransport keep-alive timeout
		ReadHeaderTimeout: 32 * time.Second, // just shy of requestTimeoutUpperBound
	}

	if !s.DisableHTTP2 {
		// At least 99% of serialized resources in surveyed clusters were smaller than 256kb.
		// This should be big enough to accommodate most API POST requests in a single frame,
		// and small enough to allow a per connection buffer of this size multiplied by `MaxConcurrentStreams`.
		const resourceBody99Percentile = 256 * 1024

		http2Options := &http2.Server{
			IdleTimeout: 90 * time.Second, // matches http.DefaultTransport keep-alive timeout
			// shrink the per-stream buffer and max framesize from the 1MB default while still accommodating most API POST requests in a single frame
			MaxUploadBufferPerStream: resourceBody99Percentile,
			MaxReadFrameSize:         resourceBody99Percentile,
		}

		// use the overridden concurrent streams setting or make the default of 250 explicit so we can size MaxUploadBufferPerConnection appropriately
		if s.HTTP2MaxStreamsPerConnection > 0 {
			http2Options.MaxConcurrentStreams = uint32(s.HTTP2MaxStreamsPerConnection)
		} else {
			// match http2.initialMaxConcurrentStreams used by clients
			// this makes it so that a malicious client can only open 400 streams before we forcibly close the connection
			// https://github.com/golang/net/commit/b225e7ca6dde1ef5a5ae5ce922861bda011cfabd
			http2Options.MaxConcurrentStreams = 100
		}

		// increase the connection buffer size from the 1MB default to handle the specified number of concurrent streams
		http2Options.MaxUploadBufferPerConnection = http2Options.MaxUploadBufferPerStream * int32(http2Options.MaxConcurrentStreams)
		// apply settings to the server
		if err := http2.ConfigureServer(secureServer, http2Options); err != nil {
			return nil, nil, fmt.Errorf("error configuring http2: %v", err)
		}
	}

	// use tlsHandshakeErrorWriter to handle messages of tls handshake error
	tlsErrorWriter := &tlsHandshakeErrorWriter{os.Stderr}
	tlsErrorLogger := log.New(tlsErrorWriter, "", 0)
	secureServer.ErrorLog = tlsErrorLogger

	klog.Infof("Serving securely on %s", secureServer.Addr)
	return apiserver.RunServer(secureServer, s.Listener, shutdownTimeout, stopCh)
}

func buildCustomTLSConfig(configClient configclientset.Interface) (*tls.Config, error) {
	// Fetch TLS config from API Server
	tlsConfig, err := getTLSConfigFromAPIServer(configClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get TLS config: %w", err)
	}

	return tlsConfig, nil
}

// tlsHandshakeErrorWriter writes TLS handshake errors to klog with
// trace level - V(5), to avoid flooding of tls handshake errors.
type tlsHandshakeErrorWriter struct {
	out io.Writer
}

const tlsHandshakeErrorPrefix = "http: TLS handshake error"

func (w *tlsHandshakeErrorWriter) Write(p []byte) (int, error) {
	if strings.Contains(string(p), tlsHandshakeErrorPrefix) {
		klog.V(5).Info(string(p))
		metrics.TLSHandshakeErrors.Inc()
		return len(p), nil
	}

	// for non tls handshake error, log it as usual
	return w.out.Write(p)
}
