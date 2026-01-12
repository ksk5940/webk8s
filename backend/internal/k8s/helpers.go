package k8s

import (
	v1 "k8s.io/api/core/v1"
)

func PodRestarts(pod *v1.Pod) int32 {
	var total int32 = 0
	for _, cs := range pod.Status.ContainerStatuses {
		total += cs.RestartCount
	}
	return total
}

func PodReadyCount(pod *v1.Pod) (ready int, total int) {
	total = len(pod.Spec.Containers)
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}
	return
}

func PodWaitingReason(pod *v1.Pod) string {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			return cs.State.Waiting.Reason
		}
	}
	return ""
}
