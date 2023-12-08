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
  namespace: "{{ .Config.Namespaces.PerfTest.GeneratedName }}"
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
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    Pipeline:
    - resources/nginx-producer.yaml
    - resources/telemetry.yaml
    - values-transforms/post-asm.sh
    - resources/shared-pvc.yaml
    - resources/jmeter-job.yaml
    - resources/jtl-processor-job.yaml
    - resources/container-resources-job.yaml
    - resources/retrieval-pod.yaml
    - executables/extract-data.sh
  Cases:
  - Name: 100TPS
    Values:
      TPS: 100
  - Name: 500TPS
    Values:
      TPS: 500
  Units:
  - Name: NoSidcar
    Values:
      TestDurationInSeconds: 600
      InjectASidecar: no
  - Name: mTLSOnly
    Values:
      TestDurationInSeconds: 600
      InjectASidecar: yes
      UseMtls: yes
      UseTelemetry: no
      UsePcapper: no
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

