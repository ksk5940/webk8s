package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
)

type MetricsResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
	Raw       any    `json:"raw,omitempty"`
}

// hits: /apis/metrics.k8s.io/v1beta1/namespaces/{ns}/pods/{pod}
func GetPodMetrics(ns, pod string) ([]byte, error) {
	cfg := RestConfig()

	rc, err := rest.RESTClientFor(&rest.Config{
		Host:    cfg.Host,
		APIPath: "",
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: cfg.NegotiatedSerializer,
		},
		BearerToken:     cfg.BearerToken,
		TLSClientConfig: cfg.TLSClientConfig,
	})
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/apis/metrics.k8s.io/v1beta1/namespaces/%s/pods/%s", ns, pod)
	return rc.Get().AbsPath(path).Do(context.TODO()).Raw()
}
