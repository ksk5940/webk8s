package k8s

import (
	"log"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	clientset *kubernetes.Clientset
	once      sync.Once
)

func Clientset() *kubernetes.Clientset {
	once.Do(func() {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("failed to get incluster config: %v", err)
		}

		cs, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			log.Fatalf("failed to create clientset: %v", err)
		}
		clientset = cs
	})
	return clientset
}

func RestConfig() *rest.Config {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to get incluster config: %v", err)
	}
	return cfg
}
