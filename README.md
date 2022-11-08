# set-value
 
## dev test

arguments

```
kpt fn source data --fn-config ./data/arg-fn-config.yaml | go run main.go
```


```
kpt fn source data --fn-config ./data/env-fn-config.yaml | go run main.go
```

```
kpt fn source data --fn-config ./data/kube-rbac-proxy-fn-config.yaml | go run main.go
```

## run

kpt fn eval -s --type mutator ./blueprint/admin  -i docker.io/henderiw/set-value:latest --fn-config ./blueprint/admin/env-fn-config.yaml

kpt fn eval -s --type mutator ./blueprint/admin  -i docker.io/henderiw/set-value:latest --fn-config ./blueprint/admin/arg-fn-config.yaml

kpt fn eval -s --type mutator ./blueprint/admin  -i docker.io/henderiw/set-value:latest --fn-config ./blueprint/admin/kube-rbac-proxy-fn-config.yaml