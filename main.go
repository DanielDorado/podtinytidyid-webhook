package main

// Based on https://github.com/kubernetes/kubernetes/blob/release-1.24/test/images/agnhost/webhook/main.go

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"k8s.io/klog/v2"
)

func (c *Config) serveMutatePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, newDelegateToV1AdmitHandler(c.mutatePods))

}

func main() {
	if len(os.Args) != 2 {
		klog.Fatalf("Usage: %s <configuration-file>", os.Args[0])
	}
	configFile := os.Args[1]
	config, err := NewConfigFromFile(configFile)
	if err != nil {
		klog.Fatalf("Loading configuration: %e", err)
	}

	http.HandleFunc("/mutating-pod", config.serveMutatePods)

	sCert, err := tls.LoadX509KeyPair(config.Server.TLS.CertFile, config.Server.TLS.KeyFile)
	if err != nil {
		klog.Fatalf("Loading: CertFile: %s, CertKey: %s: Error: %s",
			config.Server.TLS.CertFile, config.Server.TLS.KeyFile, err)
	}
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", config.Server.Port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{sCert},
		},
	}
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		klog.Fatalf("Starting HTTP server: %e", err)
	}
}
