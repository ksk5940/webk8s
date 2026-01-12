package k8s

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

// GetPodMetrics hits: /apis/metrics.k8s.io/v1beta1/namespaces/{ns}/pods/{pod}
func GetPodMetrics(ns, pod string) ([]byte, error) {
	cfg := RestConfig()

	// Create scheme and serializer for metrics API
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)

	// Create a new config for metrics API
	metricsCfg := rest.CopyConfig(cfg)
	metricsCfg.GroupVersion = &schema.GroupVersion{Group: "metrics.k8s.io", Version: "v1beta1"}
	metricsCfg.APIPath = "/apis"
	metricsCfg.NegotiatedSerializer = codecs.WithoutConversion()

	rc, err := rest.RESTClientFor(metricsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %v", err)
	}

	result := rc.Get().
		Namespace(ns).
		Resource("pods").
		Name(pod).
		Do(context.TODO())

	return result.Raw()
}

// GetNodeMetrics hits: /apis/metrics.k8s.io/v1beta1/nodes/{node}
func GetNodeMetrics(nodeName string) ([]byte, error) {
	cfg := RestConfig()

	// Create scheme and serializer for metrics API
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)

	// Create a new config for metrics API
	metricsCfg := rest.CopyConfig(cfg)
	metricsCfg.GroupVersion = &schema.GroupVersion{Group: "metrics.k8s.io", Version: "v1beta1"}
	metricsCfg.APIPath = "/apis"
	metricsCfg.NegotiatedSerializer = codecs.WithoutConversion()

	rc, err := rest.RESTClientFor(metricsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %v", err)
	}

	result := rc.Get().
		Resource("nodes").
		Name(nodeName).
		Do(context.TODO())

	return result.Raw()
}
