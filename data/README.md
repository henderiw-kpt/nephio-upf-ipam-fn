# data

## Description
sample description

## Usage

### Fetch the package
`kpt pkg get REPO_URI[.git]/PKG_PATH[@VERSION] data`
Details: https://kpt.dev/reference/cli/pkg/get/

### View package content
`kpt pkg tree data`
Details: https://kpt.dev/reference/cli/pkg/tree/

### Apply the package
```
kpt live init data
kpt live apply data --reconcile-timeout=2m --output=table
```
Details: https://kpt.dev/reference/cli/live/
