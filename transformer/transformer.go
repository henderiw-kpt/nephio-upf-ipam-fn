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
	"k8s.io/apimachinery/pkg/types"
)

// SetIP contains the information to perform the mutator function on a package
type SetIP struct {
	//targetResId resid.ResId
	//targetAPIVErsion string
	//targetKind       string

	//data                   map[string]*transformData
	upfDeployment          *upf.UpfDeploymentSpec
	upfDeploymentIndex     int
	upfDeploymentName      string
	upfDeploymentNamespace string
	ipamAllocations        map[string]*ipamv1alpha1.IPAllocationStatus
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
		//targetAPIVErsion: "nf.nephio.org/v1alpha1",
		//targetKind:       "UPFDeployment",
		//targetResId: resid.NewResIdWithNamespace(
		//	resid.Gvk{Group: "nf.nephio.org", Version: "v1alpha1", Kind: "UPFDeployment"}, "upf-us-central1", "default"),

		ipamAllocations: map[string]*ipamv1alpha1.IPAllocationStatus{},
		/*
			data: map[string]*transformData{
				"n3": {
					targetSelectorPathPrefix:  "spec.n3Interfaces.0.ips.0",
					targetSelectorPathGateway: "spec.n3Interfaces.0.gatewayIPs.0",
				},
				"n4": {
					targetSelectorPathPrefix:  "spec.n4Interfaces.0.ips.0",
					targetSelectorPathGateway: "spec.n4Interfaces.0.gatewayIPs.0",
				},
				"n6": {
					targetSelectorPathPrefix:  "spec.n6Interfaces.0.interface.ips.0",
					targetSelectorPathGateway: "spec.n6Interfaces.0.interface.gatewayIPs.0",
				},
				"internet": {targetSelectorPathPrefix: "spec.n6Interfaces.0.ueIPPool"},
			},
		*/
	}
	// gathers the ip info from the ip-allocations
	t.GatherIPInfo(rl)

	/*
		for epName, ipamStatus := range t.ipamAllocations {
			fmt.Printf("ipam network: %s, prefix: %s, gateway: %s\n",
				epName,
				ipamStatus.AllocatedPrefix,
				ipamStatus.Gateway,
			)
		}
		if t.upfDeployment == nil {
			return false, nil
		}
	*/

	// transforms the upf with the ip info collected/gathered
	/*
		if t.upfDeployment != nil {
			t.Transform2(rl)
		}
	*/
	t.Transform3(rl)

	/*
		b, _ := json.MarshalIndent(t.upfDeployment, "", "  ")
		fmt.Printf("upfdeployment:\n%s\n", string(b))
	*/
	return true, nil
}

func (t *SetIP) GatherIPInfo(rl *fn.ResourceList) {
	for i, o := range rl.Items {
		// parse the node using kyaml
		/*
			rn, err := yaml.Parse(o.String())
			if err != nil {
				rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
			}
		*/
		if o.GetAPIVersion() == "nf.nephio.org/v1alpha1" && o.GetKind() == "UPFDeployment" {
			upfDeployment := upf.UpfDeployment{
				Obj: *o,
			}
			var err error
			t.upfDeployment, err = upfDeployment.GetSpec()
			if err != nil {
				rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
			}
			t.upfDeploymentIndex = i
			t.upfDeploymentName = o.GetName()
			t.upfDeploymentNamespace = o.GetNamespace()
		}
		if o.GetAPIVersion() == "ipam.nephio.org/v1alpha1" && o.GetKind() == "IPAllocation" {
			name := o.GetLabels()[ipamv1alpha1.NephioInterfaceKey]

			//fmt.Printf("name: %s\n", name)
			ipamAlloc := ipam.IpamAllocation{
				Obj: *o,
			}

			t.ipamAllocations[name] = ipamAlloc.GetStatus()

			/*
				prefix, err := GetPrefixFromIPAlloc(rn)
				if err != nil {
					rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
				} else {
					t.data[name].prefix = strings.TrimSuffix(prefix.MustString(), "\n")
				}
				gateway, err := GetGatewayFromIPAlloc(rn)
				if err != nil {
					rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
				} else {
					t.data[name].gateway = strings.TrimSuffix(gateway.MustString(), "\n")
				}
			*/
		}
	}
}

func (t *SetIP) Transform3(rl *fn.ResourceList) {
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

/*
func (t *SetIP) Transform3(rl *fn.ResourceList) {
	for i, o := range rl.Items {
		if o.GetAPIVersion() == "nf.nephio.org/v1alpha1" && o.GetKind() == "UPFDeployment" {
			spec := o.GetMap("spec")
			for epName, ipAllocStatus := range t.ipamAllocations {
				switch epName {
				case "n3":
					eps, ok, err := r.Obj.NestedSlice("n3Interfaces")
					if err != nil {
						rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
					}
					if ok {

					}
				case "n4":
				case "n9":
				case "n6":
				default:
					// pool
			}
		}
	}
}
*/

func (t *SetIP) Transform2(rl *fn.ResourceList) {
	for epName, ipAllocStatus := range t.ipamAllocations {
		switch epName {
		case "n3":
			for i, ifConfig := range t.upfDeployment.Spec.N3Interfaces {
				if ifConfig.Name == epName {
					t.upfDeployment.Spec.N3Interfaces[i].GatewayIPs = []string{ipAllocStatus.Gateway}
					t.upfDeployment.Spec.N3Interfaces[i].IPs = []string{ipAllocStatus.AllocatedPrefix}
				}
			}
		case "n4":
			for i, ifConfig := range t.upfDeployment.Spec.N4Interfaces {
				if ifConfig.Name == epName {
					t.upfDeployment.Spec.N4Interfaces[i].GatewayIPs = []string{ipAllocStatus.Gateway}
					t.upfDeployment.Spec.N4Interfaces[i].IPs = []string{ipAllocStatus.AllocatedPrefix}
				}
			}
		case "n9":
			for i, ifConfig := range t.upfDeployment.Spec.N9Interfaces {
				if ifConfig.Name == epName {
					t.upfDeployment.Spec.N9Interfaces[i].GatewayIPs = []string{ipAllocStatus.Gateway}
					t.upfDeployment.Spec.N9Interfaces[i].IPs = []string{ipAllocStatus.AllocatedPrefix}
				}
			}
		case "n6":
			for i, ifConfig := range t.upfDeployment.Spec.N6Interfaces {
				if ifConfig.Interface.Name == epName {
					t.upfDeployment.Spec.N6Interfaces[i].Interface.GatewayIPs = []string{ipAllocStatus.Gateway}
					t.upfDeployment.Spec.N6Interfaces[i].Interface.IPs = []string{ipAllocStatus.AllocatedPrefix}
				}
			}
		default:
			// pool
			for i, ifConfig := range t.upfDeployment.Spec.N6Interfaces {
				if ifConfig.DNN == epName {
					t.upfDeployment.Spec.N6Interfaces[i].UEIPPool = ipAllocStatus.AllocatedPrefix
				}
			}
		}
	}
	obj, err := upf.BuildUPFDeploymentFn(types.NamespacedName{
		Namespace: t.upfDeploymentNamespace,
		Name:      t.upfDeploymentName,
	}, *t.upfDeployment.Spec)
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, nil))
	}
	rl.Items[t.upfDeploymentIndex] = obj

}

/*
func (t *SetIP) Transform(rl *fn.ResourceList) {
	// run over the IP addresses and get the resources
	// apply them to the upf
	for epName, transformData := range t.data {
		for i, o := range rl.Items {
			if o.GetAPIVersion() == t.targetAPIVErsion && o.GetKind() == t.targetKind {
				// parse the node using kyaml
				node, err := yaml.Parse(o.String())
				if err != nil {
					rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
				}
				switch epName {
				case "n6pool":
					if err := transformObject(
						node,
						transformData.targetSelectorPathPrefix,
						transformData.prefix,
					); err != nil {
						rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
					}
				default:
					if err := transformObject(
						node,
						transformData.targetSelectorPathPrefix,
						transformData.prefix,
					); err != nil {
						rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
					}
					if err := transformObject(
						node,
						transformData.targetSelectorPathGateway,
						transformData.gateway,
					); err != nil {
						rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
					}
				}
				str, err := node.String()
				if err != nil {
					rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
				}
				newObj, err := fn.ParseKubeObject([]byte(str))
				if err != nil {
					rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
				}
				rl.Items[i] = newObj
				break
			}
		}
	}
}
*/
/*
func transformObject(target *yaml.RNode, fp, d string) error {
	data, err := yaml.Parse(d)
	if err != nil {
		return err
	}
	err = CopyValueToTarget(target, data, &types.TargetSelector{
		FieldPaths: []string{fp},
		Options:    &types.FieldOptions{Create: true},
	})
	if err != nil {
		return err
	}
	return nil
}
*/

/*
func getIPEndpoint(t *transformData) (*yaml.RNode, error) {
	var ipEndpointTemplate = `ipv4Addr:
- {{.Prefix}}
gwv4addr: {{.Gateway}}`

	tmpl, err := template.New("ep").Parse(ipEndpointTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Prefix":  t.prefix,
		"Gateway": t.gateway,
	})
	if err != nil {
		return nil, err
	}
	return yaml.Parse(buf.String())
}

func getIP(t *transformData) (*yaml.RNode, error) {
	var ipTemplate = `{{.Prefix}}`

	tmpl, err := template.New("ip").Parse(ipTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Prefix": t.prefix,
	})
	if err != nil {
		return nil, err
	}
	return yaml.Parse(buf.String())
}
*/
