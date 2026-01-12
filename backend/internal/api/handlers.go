package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"webk8s/internal/k8s"
)

var resourceTypes = []map[string]string{
	{"key": "pods", "label": "Pods"},
	{"key": "nodes", "label": "Nodes"},
	{"key": "deployments", "label": "Deployments"},
	{"key": "replicasets", "label": "ReplicaSets"},
	{"key": "statefulsets", "label": "StatefulSets"},
	{"key": "daemonsets", "label": "DaemonSets"},
	{"key": "jobs", "label": "Jobs"},
	{"key": "cronjobs", "label": "CronJobs"},
	{"key": "configmaps", "label": "ConfigMaps"},
	{"key": "services", "label": "Services"},
}

func GetResourceTypes(c *gin.Context) {
	c.JSON(200, resourceTypes)
}

func GetNamespaces(c *gin.Context) {
	client := k8s.Clientset()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Error listing namespaces: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	out := []string{}
	for _, ns := range list.Items {
		out = append(out, ns.Name)
	}

	log.Printf("Found %d namespaces: %v", len(out), out)
	c.JSON(200, out)
}

func ListResources(c *gin.Context) {
	ns := c.Query("namespace")
	rtype := c.Query("type")

	if ns == "" && rtype != "nodes" {
		c.JSON(400, gin.H{"error": "namespace parameter is required"})
		return
	}
	if rtype == "" {
		c.JSON(400, gin.H{"error": "type parameter is required"})
		return
	}

	rows, err := k8s.ListResources(ns, rtype)
	if err != nil {
		log.Printf("Error listing resources (ns=%s, type=%s): %v", ns, rtype, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, rows)
}

func GetPodDetails(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")

	if ns == "" || podName == "" {
		c.JSON(400, gin.H{"error": "namespace and pod parameters are required"})
		return
	}

	client := k8s.Clientset()
	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting pod details (ns=%s, pod=%s): %v", ns, podName, err)
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

// NEW: Get Node Details with pods running on it
func GetNodeDetails(c *gin.Context) {
	nodeName := c.Query("node")

	if nodeName == "" {
		c.JSON(400, gin.H{"error": "node parameter is required"})
		return
	}

	client := k8s.Clientset()

	// Get node info
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting node details (node=%s): %v", nodeName, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Get pods running on this node
	pods, err := client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		log.Printf("Error listing pods on node %s: %v", nodeName, err)
	}

	podList := []map[string]string{}
	if pods != nil {
		for _, pod := range pods.Items {
			podList = append(podList, map[string]string{
				"name":      pod.Name,
				"namespace": pod.Namespace,
				"status":    string(pod.Status.Phase),
			})
		}
	}

	// Determine node status
	ready := "NotReady"
	for _, cond := range node.Status.Conditions {
		if cond.Type == v1.NodeReady {
			if cond.Status == v1.ConditionTrue {
				ready = "Ready"
			}
			break
		}
	}

	c.JSON(200, gin.H{
		"name":        node.Name,
		"status":      ready,
		"labels":      node.Labels,
		"podCount":    len(podList),
		"pods":        podList,
		"capacity":    node.Status.Capacity,
		"allocatable": node.Status.Allocatable,
	})
}

// NEW: Get Node Metrics
func GetNodeMetrics(c *gin.Context) {
	nodeName := c.Query("node")

	if nodeName == "" {
		c.JSON(400, gin.H{"error": "node parameter is required"})
		return
	}

	raw, err := k8s.GetNodeMetrics(nodeName)
	if err != nil {
		log.Printf("Node metrics not available (node=%s): %v", nodeName, err)
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

// NEW: Get Service Details with endpoints
func GetServiceDetails(c *gin.Context) {
	ns := c.Query("namespace")
	svcName := c.Query("service")

	if ns == "" || svcName == "" {
		c.JSON(400, gin.H{"error": "namespace and service parameters are required"})
		return
	}

	client := k8s.Clientset()

	// Get service
	svc, err := client.CoreV1().Services(ns).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting service details (ns=%s, svc=%s): %v", ns, svcName, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Get endpoints
	endpoints, err := client.CoreV1().Endpoints(ns).Get(context.TODO(), svcName, metav1.GetOptions{})
	endpointsList := []string{}
	if err == nil && endpoints != nil {
		for _, subset := range endpoints.Subsets {
			// Get all addresses (ready endpoints)
			for _, addr := range subset.Addresses {
				// Add each port combination
				if len(subset.Ports) > 0 {
					for _, port := range subset.Ports {
						endpointsList = append(endpointsList, fmt.Sprintf("%s:%d", addr.IP, port.Port))
					}
				} else {
					// No ports specified, just add IP
					endpointsList = append(endpointsList, addr.IP)
				}
			}
			// Also check NotReadyAddresses
			for _, addr := range subset.NotReadyAddresses {
				if len(subset.Ports) > 0 {
					for _, port := range subset.Ports {
						endpointsList = append(endpointsList, fmt.Sprintf("%s:%d (not ready)", addr.IP, port.Port))
					}
				} else {
					endpointsList = append(endpointsList, fmt.Sprintf("%s (not ready)", addr.IP))
				}
			}
		}
	} else if err != nil {
		log.Printf("Error getting endpoints for service %s: %v", svcName, err)
	}

	// If no endpoints found, add a note
	if len(endpointsList) == 0 {
		log.Printf("No endpoints found for service %s in namespace %s", svcName, ns)
	}

	c.JSON(200, gin.H{
		"name":      svc.Name,
		"namespace": svc.Namespace,
		"type":      string(svc.Spec.Type),
		"clusterIP": svc.Spec.ClusterIP,
		"ports":     svc.Spec.Ports,
		"selector":  svc.Spec.Selector,
		"labels":    svc.Labels,
		"endpoints": endpointsList,
	})
}

// NEW: Get ConfigMap Details with keys
func GetConfigMapDetails(c *gin.Context) {
	ns := c.Query("namespace")
	cmName := c.Query("configmap")

	if ns == "" || cmName == "" {
		c.JSON(400, gin.H{"error": "namespace and configmap parameters are required"})
		return
	}

	client := k8s.Clientset()
	cm, err := client.CoreV1().ConfigMaps(ns).Get(context.TODO(), cmName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting configmap details (ns=%s, cm=%s): %v", ns, cmName, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	keys := []string{}
	for k := range cm.Data {
		keys = append(keys, k)
	}
	for k := range cm.BinaryData {
		keys = append(keys, k)
	}

	c.JSON(200, gin.H{
		"name":      cm.Name,
		"namespace": cm.Namespace,
		"labels":    cm.Labels,
		"keys":      keys,
		"keyCount":  len(keys),
	})
}

func GetPodContainers(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")

	if ns == "" || podName == "" {
		c.JSON(400, gin.H{"error": "namespace and pod parameters are required"})
		return
	}

	client := k8s.Clientset()
	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting pod containers (ns=%s, pod=%s): %v", ns, podName, err)
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

	if ns == "" || podName == "" {
		c.JSON(400, gin.H{"error": "namespace and pod parameters are required"})
		return
	}

	client := k8s.Clientset()
	ev, err := client.CoreV1().Events(ns).List(context.TODO(), metav1.ListOptions{
		FieldSelector: "involvedObject.name=" + podName,
	})
	if err != nil {
		log.Printf("Error getting pod events (ns=%s, pod=%s): %v", ns, podName, err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, ev.Items)
}

func GetPodMetrics(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")

	if ns == "" || podName == "" {
		c.JSON(400, gin.H{"error": "namespace and pod parameters are required"})
		return
	}

	raw, err := k8s.GetPodMetrics(ns, podName)
	if err != nil {
		log.Printf("Metrics not available (ns=%s, pod=%s): %v", ns, podName, err)
		c.JSON(200, gin.H{
			"available": false,
			"message":   fmt.Sprintf("metrics not available: %v", err),
		})
		return
	}

	var obj any
	if err := json.Unmarshal(raw, &obj); err != nil {
		c.JSON(200, gin.H{"available": false, "message": "failed to parse metrics"})
		return
	}

	// Add available flag for easier frontend handling
	if result, ok := obj.(map[string]any); ok {
		result["available"] = true
		c.JSON(200, result)
		return
	}

	c.JSON(200, obj)
}

func StreamPodLogsSSE(c *gin.Context) {
	ns := c.Query("namespace")
	podName := c.Query("pod")
	container := c.Query("container")

	if ns == "" || podName == "" {
		c.SSEvent("message", "ERROR: namespace and pod parameters are required\n")
		return
	}

	log.Printf("Starting log stream: ns=%s, pod=%s, container=%s", ns, podName, container)

	client := k8s.Clientset()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Pod not found (ns=%s, pod=%s): %v", ns, podName, err)
		c.SSEvent("message", fmt.Sprintf("ERROR: Pod not found: %v\n", err))
		return
	}

	if container != "" {
		found := false
		for _, ct := range pod.Spec.Containers {
			if ct.Name == container {
				found = true
				break
			}
		}
		if !found {
			for _, ct := range pod.Spec.InitContainers {
				if ct.Name == container {
					found = true
					break
				}
			}
		}
		if !found {
			c.SSEvent("message", fmt.Sprintf("ERROR: Container '%s' not found in pod\n", container))
			return
		}
	}

	tail := int64(100)
	req := client.CoreV1().Pods(ns).GetLogs(podName, &v1.PodLogOptions{
		Container: container,
		Follow:    true,
		TailLines: &tail,
	})

	stream, err := req.Stream(context.TODO())
	if err != nil {
		log.Printf("Error opening log stream (ns=%s, pod=%s, container=%s): %v", ns, podName, container, err)
		c.SSEvent("message", fmt.Sprintf("ERROR: Cannot open log stream: %v\n", err))
		return
	}
	defer stream.Close()

	log.Printf("Log stream opened successfully")

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			logLine := string(buf[:n])
			c.SSEvent("message", logLine)
			c.Writer.Flush()
		}
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Log stream read error: %v", err)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}
