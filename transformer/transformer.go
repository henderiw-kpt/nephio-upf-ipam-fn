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
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/ipam/v1alpha1"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// SetIP contains the information to perform the mutator function on a package
type SetIP struct {
	//targetResId resid.ResId
	targetAPIVErsion string
	targetKind       string

	data map[string]*transformData
}

type transformData struct {
	targetSelectorPathPrefix  string
	targetSelectorPathGateway string
	prefix                    string
	gateway                   string
}

func Run(rl *fn.ResourceList) (bool, error) {
	tc := &SetIP{
		targetAPIVErsion: "nf.nephio.org/v1alpha1",
		targetKind:       "UPFDeployment",
		//targetResId: resid.NewResIdWithNamespace(
		//	resid.Gvk{Group: "nf.nephio.org", Version: "v1alpha1", Kind: "UPFDeployment"}, "upf-us-central1", "default"),
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
			"n6pool": {targetSelectorPathPrefix: "spec.n6Interfaces.0.ueIPPool"},
		},
	}
	// gathers the ip info from the ip-allocations
	tc.GatherIPInfo(rl)
	/*
		for epName, transformData := range tc.data {
			fmt.Printf("transformData: %s, prefix: %s, gateway: %s\n",
				epName,
				transformData.prefix,
				transformData.gateway,
			)
		}
	*/
	// transforms the upf with the ip info collected/gathered
	tc.Transform(rl)
	return true, nil
}

func (t *SetIP) GatherIPInfo(rl *fn.ResourceList) {
	for _, o := range rl.Items {
		// parse the node using kyaml
		rn, err := yaml.Parse(o.String())
		if err != nil {
			rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
		}
		if rn.GetApiVersion() == "ipam.nephio.org/v1alpha1" && rn.GetKind() == "IPAllocation" {
			prefix, err := GetPrefixFromIPAlloc(rn)
			if err != nil {
				rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
			}
			gateway, err := GetGatewayFromIPAlloc(rn)
			if err != nil {
				rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
			}
			name := rn.GetLabels()[ipamv1alpha1.NephioInterfaceKey]
			t.data[name].prefix = strings.TrimSuffix(prefix.MustString(), "\n")
			t.data[name].gateway = strings.TrimSuffix(gateway.MustString(), "\n")
		}
	}
}

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
