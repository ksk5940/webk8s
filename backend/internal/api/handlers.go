package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"webk8s/internal/k8s"
)

var resourceTypes = []map[string]string{
	{"key": "pods", "label": "Pods"},
	{"key": "deployments", "label": "Deployments"},
	{"key": "replicasets", "label": "ReplicaSets"},
	{"key": "statefulsets", "label": "StatefulSets"},
	{"key": "daemonsets", "label": "DaemonSets"},
	{"key": "jobs", "label": "Jobs"},
	{"key": "cronjobs", "label": "CronJobs"},
	{"key": "configmaps", "label": "ConfigMaps"},
	{"key": "secrets", "label": "Secrets"},
	{"key": "services", "label": "Services"},
}

func GetResourceTypes(c *gin.Context) {
	c.JSON(200, resourceTypes)
}

func GetNamespaces(c *gin.Context) {
	client := k8s.Clientset()
	list, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	out := []string{}
	for _, ns := range list.Items {
		out = append(out, ns.Name)
	}
	c.JSON(200, out)
}

func ListResources(c *gin.Context) {
	ns := c.Query("namespace")
	rtype := c.Query("type")

	rows, err := k8s.ListResources(ns, rtype)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, rows)
}

func GetPodDetails(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")

	client := k8s.Clientset()
	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ready, total := k8s.PodReadyCount(pod)
	restarts := k8s.PodRestarts(pod)
	reason := k8s.PodWaitingReason(pod)

	containers := []map[string]any{}
	for _, ct := range pod.Spec.Containers {
		containers = append(containers, map[string]any{
			"name":  ct.Name,
			"image": ct.Image,
		})
	}

	c.JSON(200, gin.H{
		"name":       pod.Name,
		"namespace":  pod.Namespace,
		"node":       pod.Spec.NodeName,
		"podIP":      pod.Status.PodIP,
		"phase":      string(pod.Status.Phase),
		"reason":     reason,
		"startTime":  pod.Status.StartTime,
		"ready":      fmt.Sprintf("%d/%d", ready, total),
		"restarts":   restarts,
		"containers": containers,
	})
}

// ✅ NEW: Pod containers list (containers + initContainers)
func GetPodContainers(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")

	client := k8s.Clientset()
	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	containers := []string{}
	for _, ct := range pod.Spec.Containers {
		containers = append(containers, ct.Name)
	}

	initContainers := []string{}
	for _, ct := range pod.Spec.InitContainers {
		initContainers = append(initContainers, ct.Name)
	}

	c.JSON(200, gin.H{
		"containers":     containers,
		"initContainers": initContainers,
	})
}

func GetPodEvents(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")

	client := k8s.Clientset()
	ev, err := client.CoreV1().Events(ns).List(context.TODO(), metav1.ListOptions{
		FieldSelector: "involvedObject.name=" + podName,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, ev.Items)
}

func GetPodMetrics(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")

	raw, err := k8s.GetPodMetrics(ns, podName)
	if err != nil {
		c.JSON(200, gin.H{
			"available": false,
			"message":   "metrics not available (metrics-server missing or RBAC)",
		})
		return
	}

	var obj any
	if err := json.Unmarshal(raw, &obj); err != nil {
		c.JSON(200, gin.H{"available": false, "message": "failed to parse metrics"})
		return
	}
	c.JSON(200, obj)
}

// ✅ UPDATED: logs stream supports container
func StreamPodLogsSSE(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")

	// ✅ new query param
	container := c.Query("container")

	tail := int64(50)
	client := k8s.Clientset()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	req := client.CoreV1().Pods(ns).GetLogs(podName, &v1.PodLogOptions{
		Container: container, // ✅ support specific container
		Follow:    true,
		TailLines: &tail,
	})

	stream, err := req.Stream(context.TODO())
	if err != nil {
		c.SSEvent("message", fmt.Sprintf("ERROR: %v", err))
		return
	}
	defer stream.Close()

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			c.SSEvent("message", string(buf[:n]))
			c.Writer.Flush()
		}
		if err != nil {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
}
