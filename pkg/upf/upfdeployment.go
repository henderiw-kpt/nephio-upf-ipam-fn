package upf

import (
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	nfv1alpha1 "github.com/nephio-project/nephio-pocs/nephio-5gc-controller/apis/nf/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/printers"
)

var UpfEndpointInterfaces = []string{"n3Interfaces", "n4Interfaces", "n6Interfaces", "n9Interfaces"}

type UpfDeployment struct {
	Obj fn.KubeObject
}

type UpfDeploymentSpec struct {
	Spec *nfv1alpha1.UPFDeploymentSpec
}

func getN6InterfaceConfig(obj *fn.SubObject) nfv1alpha1.N6InterfaceConfig {
	return nfv1alpha1.N6InterfaceConfig{
		DNN:       obj.GetString("dnn"),
		UEIPPool:  obj.GetString("ueIPPool"),
		Interface: getInterfaceConfig(obj.GetMap("interface")),
	}
}

func getN6EndPointInterfaces(objs fn.SliceSubObjects) []nfv1alpha1.N6InterfaceConfig {
	n6ItfceConfigs := []nfv1alpha1.N6InterfaceConfig{}
	for _, obj := range objs {
		n6ItfceConfigs = append(n6ItfceConfigs, getN6InterfaceConfig(obj))
	}
	return n6ItfceConfigs
}

func getIPs(obj *fn.SubObject) []string {
	ips := []string{}
	objs := obj.GetSlice("ips")
	for _, o := range objs {
		ips = append(ips, o.String())
	}
	return ips
}

func getGatewayIPs(obj *fn.SubObject) []string {
	ips := []string{}
	objs := obj.GetSlice("gatewayIPs")
	for _, o := range objs {
		ips = append(ips, o.String())
	}
	return ips
}

func getInterfaceConfig(obj *fn.SubObject) nfv1alpha1.InterfaceConfig {
	return nfv1alpha1.InterfaceConfig{
		Name:       obj.GetString("name"),
		IPs:        getIPs(obj),
		GatewayIPs: getGatewayIPs(obj),
	}
}

func getEndPointInterfaces(objs fn.SliceSubObjects) []nfv1alpha1.InterfaceConfig {
	itfceConfigs := []nfv1alpha1.InterfaceConfig{}
	for _, obj := range objs {
		itfceConfigs = append(itfceConfigs, getInterfaceConfig(obj))
	}
	return itfceConfigs
}

func (r *UpfDeployment) GetSpec() (*UpfDeploymentSpec, error) {
	spec := r.Obj.GetMap("spec")
	upfDeployment := &UpfDeploymentSpec{
		Spec: &nfv1alpha1.UPFDeploymentSpec{
			Capacity: nfv1alpha1.UPFCapacity{
				UplinkThroughput:   resource.MustParse(spec.GetMap("capacity").GetString("uplinkThroughput")),
				DownlinkThroughput: resource.MustParse(spec.GetMap("capacity").GetString("uplinkThroughput")),
			},
		},
	}

	for _, upfEndpointItfces := range UpfEndpointInterfaces {
		//fmt.Printf("finding ep: %s\nupf:\n%s", epName, upf.String())
		epInterfaces, ok, err := spec.NestedSlice(upfEndpointItfces)
		if err != nil {
			return nil, err
		}
		if ok {
			//fmt.Printf("ep found: %s\ndata: %s\n", epName, eps)
			switch upfEndpointItfces {
			case "n3Interfaces":
				upfDeployment.Spec.N3Interfaces = getEndPointInterfaces(epInterfaces)
			case "n4Interfaces":
				upfDeployment.Spec.N4Interfaces = getEndPointInterfaces(epInterfaces)
			case "n9Interfaces":
				upfDeployment.Spec.N9Interfaces = getEndPointInterfaces(epInterfaces)
			case "n6Interfaces":
				upfDeployment.Spec.N6Interfaces = getN6EndPointInterfaces(epInterfaces)
			}
		}
	}
	return upfDeployment, nil
}

func getUPFDeployment(nsName types.NamespacedName, spec nfv1alpha1.UPFDeploymentSpec) string {
	x := &nfv1alpha1.UPFDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UPFDeployment",
			APIVersion: "nf.nephio.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Spec: spec,
	}

	b := new(strings.Builder)
	p := printers.YAMLPrinter{}
	p.PrintObj(x, b)
	return b.String()
}

func BuildUPFDeploymentFn(nsName types.NamespacedName, spec nfv1alpha1.UPFDeploymentSpec) (*fn.KubeObject, error) {
	x := getUPFDeployment(nsName, spec)
	return fn.ParseKubeObject([]byte(x))
}
