/*
Copyright 2022 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package transformer

import (
	"github/henderiw-nephio/nephio-upf-ipam-fn/pkg/ipam"
	"github/henderiw-nephio/nephio-upf-ipam-fn/pkg/upf"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/ipam/v1alpha1"
)

// SetIP contains the information to perform the mutator function on a package
type SetIP struct {
	//targetResId resid.ResId
	//targetAPIVErsion string
	//targetKind       string

	//data                   map[string]*transformData
	//upfDeployment          *upf.UpfDeploymentSpec
	//upfDeploymentIndex     int
	//upfDeploymentName      string
	//upfDeploymentNamespace string
	ipamAllocations map[string]*ipamv1alpha1.IPAllocationStatus
}

/*
type transformData struct {
	targetSelectorPathPrefix  string
	targetSelectorPathGateway string
	prefix                    string
	gateway                   string
}
*/

func Run(rl *fn.ResourceList) (bool, error) {
	t := &SetIP{
		ipamAllocations: map[string]*ipamv1alpha1.IPAllocationStatus{},
	}
	// gathers the ip info from the ip-allocations
	t.GatherIPInfo(rl)

	t.Transform(rl)

	return true, nil
}

func (t *SetIP) GatherIPInfo(rl *fn.ResourceList) {
	for _, o := range rl.Items {

		if o.GetAPIVersion() == "ipam.nephio.org/v1alpha1" && o.GetKind() == "IPAllocation" {
			name := o.GetLabels()[ipamv1alpha1.NephioInterfaceKey]

			ipamAlloc := ipam.IpamAllocation{
				Obj: *o,
			}
			t.ipamAllocations[name] = ipamAlloc.GetStatus()
		}
	}
}

func (t *SetIP) Transform(rl *fn.ResourceList) {
	for _, o := range rl.Items {
		if o.GetAPIVersion() == "nf.nephio.org/v1alpha1" && o.GetKind() == "UPFDeployment" {
			spec := o.GetMap("spec")
			for _, upfInterfaceName := range upf.UpfEndpointInterfaces {
				if upfInterfaceName == "n6Interfaces" {
					n6itfces, ok, err := spec.NestedSlice(upfInterfaceName)
					if err != nil {
						rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
					}
					if ok {
						for _, n6itfce := range n6itfces {
							if ipamAllocStatus, ok := t.ipamAllocations[n6itfce.GetString("dnn")]; ok {
								n6itfce.SetNestedString(ipamAllocStatus.AllocatedPrefix, "ueIPPool")
							}

							if ipamAllocStatus, ok := t.ipamAllocations[n6itfce.GetMap("interface").GetString("name")]; ok {
								n6itfce.GetMap("interface").SetNestedStringSlice([]string{ipamAllocStatus.Gateway}, "gatewayIPs")
								n6itfce.GetMap("interface").SetNestedStringSlice([]string{ipamAllocStatus.AllocatedPrefix}, "ips")
							}
						}
					}
				} else {
					itfces, ok, err := spec.NestedSlice(upfInterfaceName)
					if err != nil {
						rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
					}
					if ok {
						for _, itfce := range itfces {
							if ipamAllocStatus, ok := t.ipamAllocations[itfce.GetString("name")]; ok {
								itfce.SetNestedStringSlice([]string{ipamAllocStatus.Gateway}, "gatewayIPs")
								itfce.SetNestedStringSlice([]string{ipamAllocStatus.AllocatedPrefix}, "ips")
							}
						}
					}
				}
			}
		}
	}
}
