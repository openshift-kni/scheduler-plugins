/*
 * Copyright 2024 Red Hat, Inc.
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

package zpages

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"

	knifeatures "sigs.k8s.io/scheduler-plugins/pkg-kni/features"
)

func Serve(logh logr.Logger) {
	host := ""
	port := 13579
	addr := fmt.Sprintf("%s:%d", host, port)

	logh.Info("KNI zpages serving", "addr", addr)

	pfps := NewPFPStatus(logh)

	rt := mux.NewRouter()
	rt.HandleFunc("/kniz/featuregates/", func(w http.ResponseWriter, _ *http.Request) {
		fgs := knifeatures.Desired()
		processObject(w, fgs, "featureGate", "", true)
	})
	rt.HandleFunc("/kniz/pfpstatus/", func(w http.ResponseWriter, _ *http.Request) {
		nodes := pfps.List()
		processObject(w, nodes, "node", "", true)
	})
	rt.HandleFunc("/kniz/pfpstatus/{nodeName}/", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["nodeName"]
		node, ok := pfps.Get(vars["nodeName"])
		processObject(w, node, "node", key, ok)
	})

	http.ListenAndServe(addr, rt)
}

func processObject(w http.ResponseWriter, obj any, kind, key string, ok bool) {
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		http.Error(w, fmt.Sprintf("no data for %v %v", kind, key), http.StatusInternalServerError)
		return
	}
	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		http.Error(w, fmt.Sprintf("error building the response, %v", err), http.StatusInternalServerError)
		return
	}
}
