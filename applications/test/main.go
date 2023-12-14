package main

import (
	"os"
	"text/template"
)

var template01 = `---
apiVersion: v1
kind: Pod
metadata:
  name: nginx-no-test
spec:
  containers:
  - name: nginx
    image: f5vwells/cgam-perf-test-nginx:{{ .Values.nested.n1 }}
    imagePullPolicy: IfNotPresent
    securityContext:
        seccompProfile:
            type: RuntimeDefault
`

// var input = `---
// Test:
//   Definition:
//     DefaultValues:
//       ImageVersions:
//         Producer: 0.9.0
//     Namespaces:
//       Default:
//         Basename: perftest
//     PipelineRootDirectory: /home/vwells/perftest-pipeline-root
//     Pipeline:
//     - resources/nginx-producer.yaml
//   Cases:
//   - Name: 100TPS
//     Values:
//       TPS: 100
//   Units:
//   - Name: NoSidcar
//     Values:
//       TestDurationInSeconds: 600
//       InjectASidecar: no
// `

type T struct {
	Values map[string]any
}

func main() {
	t := &T{
		Values: map[string]any{
			"foo": "bar",
			"baz": 10,
			"nested": map[string]any{
				"n1": "0.9.0",
			},
		},
	}

	tmpl, err := template.New("T").Parse(template01)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stdout, t)
	if err != nil {
		panic(err)
	}
}
