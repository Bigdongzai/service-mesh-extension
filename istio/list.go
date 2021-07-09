package istio

import (
	"fmt"
	"log"
	"strconv"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	resourceService "github.com/kubernetes/dashboard/src/app/backend/resource/service"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	client "k8s.io/client-go/kubernetes"
)

type IstioList struct {
	// 总数用于分页
	ListMeta api.ListMeta `json:"listMeta"`

	// 服务列表信息
	Istios []Istio `json:"istios"`

	// 在资源检索期间发生的非关键错误列表
	Errors []error `json:"errors"`
}

type Istio struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`

	// InternalEndpoint of all Kubernetes services that have the same label selector as connected Replication
	// Controller. Endpoint is DNS name merged with ports.
	InternalEndpoint common.Endpoint `json:"internalEndpoint"`

	// ExternalEndpoints of all Kubernetes services that have the same label selector as connected Replication
	// Controller. Endpoint is external IP address name merged with ports.
	ExternalEndpoints []common.Endpoint `json:"externalEndpoints"`

	// Label selector of the service.
	Selector map[string]string `json:"selector"`

	// Type determines how the service will be exposed.  Valid options: ClusterIP, NodePort, LoadBalancer, ExternalName
	Type v1.ServiceType `json:"type"`

	// ClusterIP is usually assigned by the master. Valid values are None, empty string (""), or
	// a valid IP address. None can be specified for headless services when proxying is not required
	ClusterIP string `json:"clusterIP"`
	// 服务对应的服务版本
	Details []Detail `json:"details"`
}

type Detail struct {
	// 版本号
	Version string `json:"version"`
	// 工作负载
	DeploymentName string `json:"deploymentName"`
	// 状态
	Status string `json:"status"`
	// 实例个数
	Instance string `json:"instance"`
	// 边车注入状态
	SiderCarStatus bool `json:"siderCarStatus"`
	// 运行时长
	CreationTimestamp metaV1.Time `json:"creationTimestamp"`
}

func GetIstioList(client client.Interface, nsQuery *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*IstioList, error) {
	log.Print("在集群中获取服务列表")
	serviceList, _ := resourceService.GetServiceList(client, nsQuery, dsQuery)
	istioList := &IstioList{
		ListMeta: serviceList.ListMeta,
		Errors:   serviceList.Errors,
		Istios:   make([]Istio, 0),
	}
	for _, service := range serviceList.Services {
		istioList.Istios = append(istioList.Istios, toIstio(client, service))
	}
	return istioList, nil
}

func toIstio(client client.Interface, service resourceService.Service) Istio {

	istio := &Istio{
		ObjectMeta:        service.ObjectMeta,
		InternalEndpoint:  service.InternalEndpoint,
		ExternalEndpoints: service.ExternalEndpoints,
		Selector:          service.Selector,
		Type:              service.Type,
		ClusterIP:         service.ClusterIP,
		Details:           make([]Detail, 0),
	}
	if service.Selector != nil {
		labelSelector := labels.SelectorFromSet(service.Selector)
		channels := &common.ResourceChannels{
			PodList: common.GetPodListChannelWithOptions(client, common.NewSameNamespaceQuery(service.ObjectMeta.Namespace),
				metaV1.ListOptions{
					LabelSelector: labelSelector.String(),
					FieldSelector: fields.Everything().String(),
				}, 1),
		}
		apiPodList := <-channels.PodList.List
		for _, apiPod := range apiPodList.Items {
			// labels := apiPod.ObjectMeta.Labels
			// // 版本
			// version := labels["version"]
			// // 负载名称
			// deploymentName := apiPod.ObjectMeta.Name
			// // 运行时长
			// creationTimestamp := apiPod.ObjectMeta.CreationTimestamp
			// // 状态
			// status := getPodStatus(apiPod)
			// // 边车状态
			// siderCarStatuss := true
			// // 运行个数
			// instance := strconv.Itoa(len(apiPod.Spec.Containers)) + " | " + strconv.Itoa(len(apiPod.Spec.Containers))
			detail := &Detail{
				Version:           apiPod.ObjectMeta.Labels["version"],
				DeploymentName:    apiPod.ObjectMeta.Name,
				Status:            getPodStatus(apiPod),
				SiderCarStatus:    true,
				Instance:          strconv.Itoa(len(apiPod.Spec.Containers)) + " | " + strconv.Itoa(len(apiPod.Spec.Containers)),
				CreationTimestamp: apiPod.ObjectMeta.CreationTimestamp,
			}
			istio.Details = append(istio.Details, *detail)
		}
	}
	return *istio
}

// getPodStatus returns status string calculated based on the same logic as kubectl
// Base code: https://github.com/kubernetes/kubernetes/blob/master/pkg/printers/internalversion/printers.go#L734
func getPodStatus(pod v1.Pod) string {
	restarts := 0
	readyContainers := 0

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init: Signal %d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init: ExitCode %d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = fmt.Sprintf("Init: %s", container.State.Waiting.Reason)
			initializing = true
		default:
			reason = fmt.Sprintf("Init: %d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal: %d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode: %d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			if hasPodReadyCondition(pod.Status.Conditions) {
				reason = string(v1.PodRunning)
			} else {
				reason = "NotReady"
			}
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = string(v1.PodUnknown)
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	if len(reason) == 0 {
		reason = string(v1.PodUnknown)
	}

	return reason
}

func hasPodReadyCondition(conditions []v1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}
