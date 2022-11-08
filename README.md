# set-value
 
## dev test

arguments

```
kpt fn source data | go run main.go
```

## run

kpt fn eval -s --type mutator ./blueprint/admin  -i docker.io/henderiw/set-value:latest --fn-config ./blueprint/admin/env-fn-config.yaml

kpt fn eval --type mutator ./data  -i docker.io/henderiw/nephio-upf-ipam-fn:latest --fn-config ./data/package-context.yaml

