package virtualservice

import (
	"context"
	"log"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/rest"
)

// type VirtualServiceList struct {
// 	VirtualServices []VirtualService `json:"virtualServices"`
// }
// type VirtualService struct {
// 	VSName string `json:"vsName"`
// }

func GetVirtualServiceList(restConfig *rest.Config, nsQuery *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*v1alpha3.VirtualServiceList, error) {
	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("创建istio客户端失败: %s", err)
	}
	vsList, err := ic.NetworkingV1alpha3().VirtualServices("default").List(context.TODO(), metav1.ListOptions{})
	return vsList, err
}

// func main() {
// 	kubeconfig := "/root/.kube/config"
// 	namespace := "default"

// 	if len(kubeconfig) == 0 || len(namespace) == 0 {
// 		log.Fatalf("Environment variables KUBECONFIG and NAMESPACE need to be set")
// 	}

// 	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
// 	if err != nil {
// 		log.Fatalf("Failed to create k8s rest client: %s", err)
// 	}

// 	ic, err := versionedclient.NewForConfig(restConfig)
// 	if err != nil {
// 		log.Fatalf("Failed to create istio client: %s", err)
// 	}

// 	// Test VirtualServices
// 	vsList, err := ic.NetworkingV1alpha3().VirtualServices(namespace).List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		log.Fatalf("Failed to get VirtualService in %s namespace: %s", namespace, err)
// 	}

// 	for i := range vsList.Items {
// 		vs := vsList.Items[i]
// 		log.Printf("Index: %d VirtualService Hosts: %+v\n", i, vs.Spec.GetHosts())
// 	}

// 	// Test DestinationRules
// 	drList, err := ic.NetworkingV1alpha3().DestinationRules(namespace).List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		log.Fatalf("Failed to get DestinationRule in %s namespace: %s", namespace, err)
// 	}

// 	for i := range drList.Items {
// 		dr := drList.Items[i]
// 		log.Printf("Index: %d DestinationRule Host: %+v\n", i, dr.Spec.GetHost())
// 	}

// 	// Test Gateway
// 	gwList, err := ic.NetworkingV1alpha3().Gateways(namespace).List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		log.Fatalf("Failed to get Gateway in %s namespace: %s", namespace, err)
// 	}

// 	for i := range gwList.Items {
// 		gw := gwList.Items[i]
// 		for _, s := range gw.Spec.GetServers() {
// 			log.Printf("Index: %d Gateway servers: %+v\n", i, s)
// 		}
// 	}

// 	// Test ServiceEntry
// 	seList, err := ic.NetworkingV1alpha3().ServiceEntries(namespace).List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		log.Fatalf("Failed to get ServiceEntry in %s namespace: %s", namespace, err)
// 	}

// 	for i := range seList.Items {
// 		se := seList.Items[i]
// 		for _, h := range se.Spec.GetHosts() {
// 			log.Printf("Index: %d ServiceEntry hosts: %+v\n", i, h)
// 		}
// 	}
// }
