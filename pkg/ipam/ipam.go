package ipam

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/ipam/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IpamAllocation struct {
	Obj fn.KubeObject
}

func (r *IpamAllocation) GetSpec() (*ipamv1alpha1.IPAllocationSpec, error) {
	spec := r.Obj.GetMap("spec")
	selectorLabels, _, err := spec.NestedStringMap("selector", "matchLabels")
	if err != nil {
		return nil, err
	}

	ipAllocSpec := &ipamv1alpha1.IPAllocationSpec{
		PrefixKind:    spec.GetString("kind"),
		AddressFamily: spec.GetString("addressFamily"),
		Prefix:        spec.GetString("prefix"),
		PrefixLength:  uint8(spec.GetInt("prefixLength")),
		Selector: &metav1.LabelSelector{
			MatchLabels: selectorLabels,
		},
	}

	return ipAllocSpec, nil
}

func (r *IpamAllocation) GetStatus() *ipamv1alpha1.IPAllocationStatus {
	status := r.Obj.GetMap("status")

	return &ipamv1alpha1.IPAllocationStatus{
		AllocatedPrefix: status.GetString("allocatedprefix"),
		Gateway:         status.GetString("gateway"),
	}
}
