// Copyright © 2020 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	webhook "github.com/openebs/cstor-operators/pkg/webhook"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	clientset "github.com/openebs/api/v3/pkg/client/clientset/versioned"
	//snapclientset "github.com/openebs/maya/pkg/client/generated/openebs.io/snapshot/v1alpha1/clientset/internalclientset"
)

var (
	kubeconfig string
)

func main() {
	var parameters webhook.Parameters

	// get command line parameters
	flag.IntVar(&parameters.Port, "port", 8443, "Webhook server port.")
	flag.StringVar(&parameters.CertFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.KeyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")

	klog.InitFlags(nil)
	err := flag.Set("logtostderr", "true")
	if err != nil {
		klog.Fatalf("Failed to set logtostderr flag: %s", err.Error())
	}
	flag.Parse()

	// Get in cluster config
	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	// Building Kubernetes Clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	// Building OpenEBS Clientset
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building openebs clientset: %s", err.Error())
	}

	// Building Snapshot Clientset
	// snapClient, err := snapclientset.NewForConfig(cfg)
	// if err != nil {
	// 	klog.Fatalf("Error building openebs snapshot clientset: %s", err.Error())
	// }

	// Fetch a reference to the admission server deployment object
	ownerReference, err := webhook.GetAdmissionReference(kubeClient)
	if err != nil {
		klog.Fatal(err, "failed to get a reference of the admission deployment object")
	}
	validatorErr := webhook.InitValidationServer(*ownerReference, kubeClient)
	if validatorErr != nil {
		klog.Fatal(validatorErr, "failed to initialize validation server")
	}

	wh, err := webhook.New(parameters, kubeClient, openebsClient)
	if err != nil {
		klog.Fatalf("failed to create validation webhook: %s", err.Error())
	}
	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", wh.Serve)
	wh.Server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := wh.Server.ListenAndServeTLS("", ""); err != nil {
			klog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	klog.Info("Webhook server started")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-signalChan

	klog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	err = wh.Server.Shutdown(context.Background())
	if err != nil {
		klog.Errorf("failed to shutdown server: error {%v}", err)
	}
}

// GetClusterConfig return the config for k8s.
func getClusterConfig(kubeconfig string) (*rest.Config, error) {
	var masterURL string
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("Failed to get k8s Incluster config. %+v", err)
		if kubeconfig == "" {
			return nil, fmt.Errorf("Kubeconfig is empty: %v", err.Error())
		}
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("Error building kubeconfig: %s", err.Error())
		}
	}
	return cfg, err
}
