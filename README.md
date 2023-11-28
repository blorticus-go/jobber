# jobber

Run Kubernetes Job pipeline

## Sketch

### Templates

```yaml
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx-no-test
  namespace: "{{ .Config.Namespaces.PerfTest }}"
spec:
  containers:
  - name: nginx
    image: f5vwells/cgam-perf-test-nginx:{{ .Values.Producer.ImageVersion }}
    imagePullPolicy: IfNotPresent
    securityContext:
        seccompProfile:
            type: RuntimeDefault
```

### Configuration file

```yaml
# run-config.yaml
---
RunConfiguration:
    Namespaces:
        - ReferenceName: PerfTest
          BaseName: perf-test
    DefaultNamespace: PerfTest
    Pipeline:
        - resources/nginx-producer.yaml
        - resources/telemetry.yaml
        - values-transforms/post-asm.sh
        - resources/shared-pvc.yaml
        - resources/jmeter-job.yaml
        - resources/jtl-processor-job.yaml
        - resources/container-resources-job.yaml
        - resources/retrieval-pod.yaml
        - runner/extract-data.sh
```

### Values files

```yaml
# producer-values.yaml
---
Producer:
    ImageVersion: 0.8.0
```

### Transforms

```bash
#!/bin/bash
#  version-transform.sh
sed 's/^( +)ImageVersion: .+/   \1ImageVersion: 0.9.0'
```

