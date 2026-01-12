package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
)

// GetNodeMetrics hits: /apis/metrics.k8s.io/v1beta1/nodes/{node}
func GetNodeMetrics(nodeName string) ([]byte, error) {
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

	path := fmt.Sprintf("/apis/metrics.k8s.io/v1beta1/nodes/%s", nodeName)
	return rc.Get().AbsPath(path).Do(context.TODO()).Raw()
}
