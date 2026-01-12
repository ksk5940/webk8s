package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// -----------------------------
// Apps Workloads
// -----------------------------

func listDeployments(namespace string) ([]ResourceRow, error) {
	cs := Clientset()

	list, err := cs.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]ResourceRow, 0, len(list.Items))
	for i := range list.Items {
		d := &list.Items[i]

		ready := int32(0)
		if d.Status.ReadyReplicas > 0 {
			ready = d.Status.ReadyReplicas
		}

		replicas := int32(0)
		if d.Spec.Replicas != nil {
			replicas = *d.Spec.Replicas
		}

		out = append(out, ResourceRow{
			Name:              d.Name,
			Namespace:         d.Namespace,
			CreationTimestamp: d.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
			Labels:            d.Labels,
			Status: map[string]any{
				"readyReplicas": ready,
				"replicas":      replicas,
				"updated":       d.Status.UpdatedReplicas,
				"available":     d.Status.AvailableReplicas,
			},
		})
	}
	return out, nil
}

func listReplicaSets(namespace string) ([]ResourceRow, error) {
	cs := Clientset()

	list, err := cs.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]ResourceRow, 0, len(list.Items))
	for i := range list.Items {
		rs := &list.Items[i]

		replicas := int32(0)
		if rs.Spec.Replicas != nil {
			replicas = *rs.Spec.Replicas
		}

		out = append(out, ResourceRow{
			Name:              rs.Name,
			Namespace:         rs.Namespace,
			CreationTimestamp: rs.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
			Labels:            rs.Labels,
			Status: map[string]any{
				"readyReplicas": rs.Status.ReadyReplicas,
				"replicas":      replicas,
				"available":     rs.Status.AvailableReplicas,
			},
		})
	}
	return out, nil
}

func listStatefulSets(namespace string) ([]ResourceRow, error) {
	cs := Clientset()

	list, err := cs.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]ResourceRow, 0, len(list.Items))
	for i := range list.Items {
		sts := &list.Items[i]

		replicas := int32(0)
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}

		out = append(out, ResourceRow{
			Name:              sts.Name,
			Namespace:         sts.Namespace,
			CreationTimestamp: sts.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
			Labels:            sts.Labels,
			Status: map[string]any{
				"readyReplicas": sts.Status.ReadyReplicas,
				"replicas":      replicas,
				"updated":       sts.Status.UpdatedReplicas,
				"current":       sts.Status.CurrentReplicas,
			},
		})
	}
	return out, nil
}

func listDaemonSets(namespace string) ([]ResourceRow, error) {
	cs := Clientset()

	list, err := cs.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]ResourceRow, 0, len(list.Items))
	for i := range list.Items {
		ds := &list.Items[i]

		out = append(out, ResourceRow{
			Name:              ds.Name,
			Namespace:         ds.Namespace,
			CreationTimestamp: ds.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
			Labels:            ds.Labels,
			Status: map[string]any{
				"readyReplicas": ds.Status.NumberReady,
				"replicas":      ds.Status.DesiredNumberScheduled,
				"current":       ds.Status.CurrentNumberScheduled,
				"available":     ds.Status.NumberAvailable,
			},
		})
	}
	return out, nil
}

// -----------------------------
// Batch Workloads
// -----------------------------

func listJobs(namespace string) ([]ResourceRow, error) {
	cs := Clientset()

	list, err := cs.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]ResourceRow, 0, len(list.Items))
	for i := range list.Items {
		j := &list.Items[i]

		desired := int32(1)
		if j.Spec.Parallelism != nil {
			desired = *j.Spec.Parallelism
		}

		completions := int32(0)
		if j.Spec.Completions != nil {
			completions = *j.Spec.Completions
		}

		out = append(out, ResourceRow{
			Name:              j.Name,
			Namespace:         j.Namespace,
			CreationTimestamp: j.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
			Labels:            j.Labels,
			Status: map[string]any{
				"active":      j.Status.Active,
				"succeeded":   j.Status.Succeeded,
				"failed":      j.Status.Failed,
				"parallelism": desired,
				"completions": completions,
			},
		})
	}
	return out, nil
}

func listCronJobs(namespace string) ([]ResourceRow, error) {
	cs := Clientset()

	list, err := cs.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make([]ResourceRow, 0, len(list.Items))
	for i := range list.Items {
		cj := &list.Items[i]

		lastSchedule := ""
		if cj.Status.LastScheduleTime != nil {
			lastSchedule = cj.Status.LastScheduleTime.Time.Format("2006-01-02T15:04:05Z")
		}

		out = append(out, ResourceRow{
			Name:              cj.Name,
			Namespace:         cj.Namespace,
			CreationTimestamp: cj.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
			Labels:            cj.Labels,
			Status: map[string]any{
				"schedule":       cj.Spec.Schedule,
				"suspend":        fmt.Sprintf("%v", cj.Spec.Suspend != nil && *cj.Spec.Suspend),
				"activeJobs":     len(cj.Status.Active),
				"lastScheduleAt": lastSchedule,
			},
		})
	}
	return out, nil
}
