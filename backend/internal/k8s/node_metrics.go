package k8s

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

// GetNodeMetrics hits: /apis/metrics.k8s.io/v1beta1/nodes/{node}
func GetNodeMetrics(nodeName string) ([]byte, error) {
	cfg := RestConfig()

	// Create a copy of the config with metrics.k8s.io settings
	metricsCfg := *cfg
	metricsCfg.APIPath = "/apis"
	metricsCfg.GroupVersion = &schema.GroupVersion{Group: "metrics.k8s.io", Version: "v1beta1"}
	metricsCfg.NegotiatedSerializer = cfg.NegotiatedSerializer

	rc, err := rest.RESTClientFor(&metricsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %v", err)
	}

	result := rc.Get().
		Resource("nodes").
		Name(nodeName).
		Do(context.TODO())

	return result.Raw()
}
