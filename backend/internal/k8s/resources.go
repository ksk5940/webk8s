package k8s

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceRow struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	CreationTimestamp string            `json:"creationTimestamp"`
	Status            map[string]any    `json:"status"`
	Labels            map[string]string `json:"labels"`
}

func ListResources(namespace, rtype string) ([]ResourceRow, error) {
	cs := Clientset()
	rtype = strings.ToLower(rtype)

	switch rtype {

	// -----------------------------
	// Core resources
	// -----------------------------
	case "pods":
		list, err := cs.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		out := make([]ResourceRow, 0, len(list.Items))
		for i := range list.Items {
			pod := &list.Items[i]
			ready, total := PodReadyCount(pod)
			restarts := PodRestarts(pod)
			reason := PodWaitingReason(pod)

			out = append(out, ResourceRow{
				Name:              pod.Name,
				Namespace:         pod.Namespace,
				CreationTimestamp: pod.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
				Labels:            pod.Labels,
				Status: map[string]any{
					"phase":    string(pod.Status.Phase),
					"ready":    fmt.Sprintf("%d/%d", ready, total),
					"restarts": restarts,
					"nodeName": pod.Spec.NodeName,
					"reason":   reason,
				},
			})
		}
		return out, nil

	case "services":
		list, err := cs.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		out := make([]ResourceRow, 0, len(list.Items))
		for i := range list.Items {
			svc := &list.Items[i]
			out = append(out, ResourceRow{
				Name:              svc.Name,
				Namespace:         svc.Namespace,
				CreationTimestamp: svc.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
				Labels:            svc.Labels,
				Status: map[string]any{
					"type":      string(svc.Spec.Type),
					"clusterIP": svc.Spec.ClusterIP,
				},
			})
		}
		return out, nil

	case "configmaps":
		list, err := cs.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		out := make([]ResourceRow, 0, len(list.Items))
		for i := range list.Items {
			cm := &list.Items[i]
			out = append(out, ResourceRow{
				Name:              cm.Name,
				Namespace:         cm.Namespace,
				CreationTimestamp: cm.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
				Labels:            cm.Labels,
				Status: map[string]any{
					"keys": len(cm.Data),
				},
			})
		}
		return out, nil

	case "secrets":
		list, err := cs.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		out := make([]ResourceRow, 0, len(list.Items))
		for i := range list.Items {
			s := &list.Items[i]
			out = append(out, ResourceRow{
				Name:              s.Name,
				Namespace:         s.Namespace,
				CreationTimestamp: s.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
				Labels:            s.Labels,
				Status: map[string]any{
					"type": string(s.Type),
				},
			})
		}
		return out, nil

	// -----------------------------
	// Apps resources
	// -----------------------------
	case "deployments":
		return listDeployments(namespace)

	case "replicasets":
		return listReplicaSets(namespace)

	case "statefulsets":
		return listStatefulSets(namespace)

	case "daemonsets":
		return listDaemonSets(namespace)

	// -----------------------------
	// Batch resources
	// -----------------------------
	case "jobs":
		return listJobs(namespace)

	case "cronjobs":
		return listCronJobs(namespace)
	}

	return nil, errors.New("unsupported resource type: " + rtype)
}

// Used by logs stream (kept here in case you later want container selection)
func defaultLogOptions() *v1.PodLogOptions {
	tail := int64(50)
	return &v1.PodLogOptions{
		Follow:    true,
		TailLines: &tail,
	}
}
