# jobber

Run Kubernetes Job pipeline

## Notes

### Variables

- The Values will be strictly hierarchical.  They will not merge down.  This means all values must be explicitly stated for each Unit and Case.

```go
type PipelineValues struct {
  Global map[string]any
  Unit map[string]any
  Case map[string]any
}

type PipelineConfigArchiveInformation struct {
  FilePath string
}

type PipelineConfig struct {
  Archive *PipelineConfigArchiveInformation
}

type PipelineRuntimeContextUnit struct {
  Name string
}

type PipelineRuntimeContextCase struct {
  Name string
  RetrievedAssetsDirectoryPath string
}

type PipelineRuntimeContext struct {
  CurrentUnit *PipelineRuntimeContextUnit
  CurrentCase *PipelineRuntimeContextCase
}

type PipelineDefaultNamespace {
  Name string
}

func (p *PipelineDefaultNamespace) Pod(name string) *wrapped.Pod {}

type PipelineRuntime struct {
  DefaultNamespace *PipelineDefaultNamespace
  Context *PipelineRuntimeContext
}

func (p *PipelineRuntime) Pod(name string, inNamespaceNamed string) *wrapped.Pod {}
func (p *PipelineRuntime) ServiceAccount(named string, inNamespaceNamed string) *wrapped.ServiceAccount {}

type PipelineVariables struct {
  Values *PipelineValues
  Config *PipelineConfig
  Runtime *PipelineRuntime
}
```

```go
package wrapped

type Resource interface {
  Name() string
  NamespaceName() string
  GroupVersionKind() schema.GroupVersionKind
  GroupVersionResource() schema.GroupVersionResource
  Create() error
  Delete() error
  UpdateStatus() error
  Unstructured() *unstructured.Unstructured
  UnstructuredMap() map[string]any
}

type Pod struct {}  // is a Resource

func (p *Pod) TypedApiObject() *corev1.Pod {}
func (p *Pod) PodIpAsAString() string {}

type ServiceAccount struct {}  // is a Resource

func (s *ServiceAccount) TypedApiObject() *corev1.ServiceAccount {}

func (n *Namespace) TypedApiObject() *corev1.Namespace {}

func NewNamespaceUsingGeneratedName(basename string) *Namespace {}
```

```go
package api

type Client struct {}

func (c *Client) Set() *ClientSet {}
func (c *Client) Dynamic() *DynamicClient {}
func (c *Client) Discovery() *DiscoveryClient {}
```

```go
package wrapped

type DeletionResult struct {
  SuccessfullyDeletedResources []Resource
  ResourceForWhichDeletionFailed Resource
  Error error
}

type ResourceTracker struct {}

func (t *ResourceTracker) AddCreatedResource(r Resource) *ResourceTracker {}
func (t *ResourceTracker) AttemptToDeleteAllAsYetUndeletedResources() *DeletionResult {}
```

### Configuration file

- expansion in config.yaml happens from command-line variables.  For example:

```bash
$ jobber -config /path/to/config -set target-version=18.3.1-am4 -set date=$(date +%F)
```

```yaml
# run-config.yaml
---
Test:
  Archive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
  - Name: NoSidecar
    Values:
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
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

