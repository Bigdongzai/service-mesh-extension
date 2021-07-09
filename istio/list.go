package istio

import (
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	resourceService "github.com/kubernetes/dashboard/src/app/backend/resource/service"
	v1 "k8s.io/api/core/v1"
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
	CreationTimestamp string `json:"creationTimestamp"`
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
		istioList.Istios = append(istioList.Istios, toIstio(service))
	}
	return istioList, nil
}

func toIstio(service resourceService.Service) Istio {
	istio := &Istio{
		ObjectMeta:        service.ObjectMeta,
		InternalEndpoint:  service.InternalEndpoint,
		ExternalEndpoints: service.ExternalEndpoints,
		Selector:          service.Selector,
		Type:              service.Type,
		ClusterIP:         service.ClusterIP,
		Details:           make([]Detail, 0),
	}
	return *istio
}
