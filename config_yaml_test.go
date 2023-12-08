package jobber_test

import (
	"strings"
	"testing"

	"github.com/blorticus-go/jobber"
)

var goodConfigs = map[string]string{
	"goodConfig01": `---
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
`,
}

var badConfigs = map[string]string{
	"config cannot be empty": `---
`,
	"invalid yaml will fail": `---
  florb
`,
	".Test must be defined": `---
Foo:
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
  `,
	".Test.Definition must exist": `---
Test:
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
`,
	".Test.Definition.Namespaces must exist": `---
Test:
  Definition:
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
`,
	".Test.Definition.Namespaces cannot be empty": `---
Test:
  Definition:
    Namespaces:
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
`,
	".Test.Definition.Namespaces.Default must exist": `---
Test:
  Definition:
    Namespaces:
      Florp:
        Basename: florp
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
`,
	".Test.Definition.Pipeline must exist": `---
Test:
    Definition:
      Namespaces:
        Default:
          Basename: perftest
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
  `,
	".Test.Definition.Pipeline cannot be an empty list": `---
Test:
    Definition:
      Namespaces:
        Default:
          Basename: perftest
      Pipeline: []
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
  `,
	".Test.Cases must exist": `---
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
  `,
	".Test.Cases cannot be an empty list": `---
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
    Cases: []
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
  `,
	".Test.Units must be defined": `---
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
  `,
	".Test.Units cannot be empty": `---
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
    Units: []
  `,
	".Test.Definition.Pipeline action type (florp) is not valid": `---
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - florp/shared-pvc.yaml
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
`,
	".Test.Definition.Pipeline entries must be <type>/<target>": `---
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - post-asm.sh
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
`,
	".Test.Definition.Pipeline entries must be <type>/<target> without additional slashes": `---
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - resources/shared-pvc.yaml/baz
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
`,
}

func TestGoodConfigs(t *testing.T) {
	for configName, configYamlString := range goodConfigs {
		if _, err := jobber.ReadConfigurationYamlFromReader(strings.NewReader(configYamlString)); err != nil {
			if err != nil {
				t.Errorf("on good config named [%s] received an error while processing: %s", configName, err.Error())
			}
		}
	}
}

func TestBadConfigs(t *testing.T) {
	for configName, configYamlString := range badConfigs {
		if _, err := jobber.ReadConfigurationYamlFromReader(strings.NewReader(configYamlString)); err != nil {
			if err == nil {
				t.Errorf("on bad config named [%s] expected an error while processing, but received none", configName)
			}
		}
	}
}
