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
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// SetIP contains the information to perform the mutator function on a package
type SetIP struct {
	targetResId resid.ResId

	data map[string]*transformData
}

type transformData struct {
	targetSelectorPath string
	prefix             string
	gateway            string
}

func Run(rl *fn.ResourceList) (bool, error) {
	tc := &SetIP{
		targetResId: resid.NewResIdWithNamespace(
			resid.Gvk{Group: "networkfunction.nephio.io", Version: "v1alpha1", Kind: "Upf"}, "free5gc-upf-1", "default"),
		data: map[string]*transformData{
			"n3":     {targetSelectorPath: "spec.n3.endpoints.0"},
			"n4":     {targetSelectorPath: "spec.n4.endpoints.0"},
			"n6":     {targetSelectorPath: "spec.n6.endpoints.internet.ipendpoints"},
			"n6pool": {targetSelectorPath: "spec.n6.endpoints.internet.ipaddrpool"},
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
			t.data[rn.GetName()].prefix = strings.TrimSuffix(prefix.MustString(), "\n")
			t.data[rn.GetName()].gateway = strings.TrimSuffix(gateway.MustString(), "\n")
		}
	}
}

func (t *SetIP) Transform(rl *fn.ResourceList) {
	// run over the IP addresses and get the resources
	// apply them to the upf
	for epName, transformData := range t.data {
		selector := &types.TargetSelector{
			Select: &types.Selector{
				ResId: t.targetResId,
			},
			FieldPaths: []string{
				transformData.targetSelectorPath,
			},
			Options: &types.FieldOptions{
				Create: true,
			},
		}
		// input validation
		if selector.Select == nil {
			rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(fmt.Errorf("target must specify a resource to select"), rl.FunctionConfig))
		}
		if len(selector.FieldPaths) == 0 {
			rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(fmt.Errorf("no fieldPaths selected"), rl.FunctionConfig))
		}
		for i, o := range rl.Items {
			// parse the node using kyaml
			node, err := yaml.Parse(o.String())
			if err != nil {
				rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
			}
			// provide a resource id based on GVKNNS
			ids, err := MakeResIds(node)
			if err != nil {
				rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
			}

			// filter targets by matching resource IDs
			//fmt.Printf("resid %v, selectorResId: %v\n", ids, selector.Select.ResId)
			for _, id := range ids {
				if id.IsSelectedBy(selector.Select.ResId) {
					//fmt.Printf("selected by resid, selector: %v\n", selector)

					switch epName {
					case "n6pool":
						data, err := getIP(transformData)
						if err != nil {
							rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
						}
						//fmt.Printf("transform input data: %v\n", data.MustString())
						err = CopyValueToTarget(node, data, selector)
						if err != nil {
							rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
						}
					default:
						data, err := getIPEndpoint(transformData)
						if err != nil {
							rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, o))
						}
						//fmt.Printf("transform input data: %v\n", data.MustString())
						err = CopyValueToTarget(node, data, selector)
						if err != nil {
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
}

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
