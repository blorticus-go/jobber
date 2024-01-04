# jobber

Run Kubernetes Job pipeline

## Rationale

For performance testing Kubernetes based components, a useful pattern is to decompose the testing and results processing into a set of [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/) in a pipeline.  For example, imagine that you wish to test HTTP/1.1 performance through an Envoy sidecar injected into two Pods, one acting as a client, the other as a server.  As a test client, one might use [Jmeter](https://jmeter.apache.org/) and as a test server, one might use [NGINX](https://nginx.com).  JMeter can produce a log file with one entry for each request it sends.  Each line of the file records the request URL, the response code, the time-to-first-byte and any error that was generated.  This file could be processed to generate a summary, particularly the min/mean/median/max/stdev for TTFB.

Let us say that, as part of the performance test, one wishes to know the CPU and memory min/mean/median/max/stdev of each of the Envoy sidecar containers during a test that runs for 10 minutes, as well as the min/mean/media/max/stdev for TTFB for retrieved objects of various sizes.  One could write a Job that interacts with a cluster-wide [Prometheus](https://prometheus.io/) instance to get the CPU and memory usage of the containers.  If the JTL file produced by JMeter were available to a Job, it could process the JTL file and produce the desired summary stats.  A [Persistent Volume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) shared between Jobs could do this.  The results of the JTL and cpu/memory summarization could also be written to this PV.  When the test and processing are complete, one might wish to connect a Pod to the shared PV and extract the raw data files and summarizations, writing them to some local data store.

The reason for using Jobs to accomplish each step is that the Job functions (e.g., creating a statistical summary from a JTL file) are generic and can be used for any number of test scenarios.  Decoupling this from the test scenario itself (in this case, the JMeter scenario that executes) allows it to be easily slotted into other testing scenarios.

When a test completes, if everything is successful, it makes sense to delete the created Kubernetes resources automatically.

One's first impulse may be to use [Helm](https://helm.sh).  Helm, however, is patterned around deploying configurable packages of resources that are related and non-transient.  Jobs, by their nature, are transient.  They start, run (via one or more Pods), then exit when the associated unit fo work is done.  Helm can be coerced into handling Jobs, but doing so is unwieldy and difficult to both maintain and debug.

Naturally, there are many powerful pipeline tools available.  At the most basic, this sort of pipeline could be run using [Ansible](https://www.ansible.com) or even Bash.  The problem with a system like Ansible is that the output cannot be completely controlled.  If a performance test requires multiple discreet, repeated sub-tests, with a set of variables changing between each sub-test (see the section on Test Cases and Test Units below), there may be many steps in the test.  If one step fails, it is desirable to keep the Kubernetes resources created for the sub-test case to remain.  This makes troubleshooting much easier.  But it must also be manually cleaned up.  Ansible output makes it difficult to understand what has been created and when a failure happens, what went wrong.  There are ways to coerce it into writing this information to a human-readable file, for example, but this is unnatural for Ansible and would be unwieldy to implement, to test, and to debug.

Moreover, this type of Job-driven pipeline performance testing can be patterned for many different testing scenarios.  Using Ansible generally means each scenario would be a playbook, likely with lots of repeated code across different testing scenarios.  This could be modularized, but at that point, Ansible isn't providing much help.

Finally, neither Ansible nor shell scripting lend themselves well to readability, extensibility and framework testing.

The panoply of sophisticated pipeline tools, on the other hand, are extensible specifically for this type of work.  Unfortunately, their flexibility -- and concomitant complexity -- is the problem.  Performance testing for this Job-driven pattern really only needs to support a few types of actions.  These can all be modeled in those pipeline frameworks, but so much scaffolding is necessary that the shape of this pattern gets easily lost.  And when one wishes to extend the resulting system for a specific feature, a developer must spend considerable effort to understand the complex framework to make what might otherwise be simple feature additions.

This leads to an application written specifically to support this Job-driven pattern.  It is opinionated, and thus considerably less flexible that other generic pipeline frameworks, but the tradeoff is, it is much simpler to understand and extend.

## The Testing Pattern

`jobber` divides performance testing into Test Units and Test Cases.  A Test Case is single, self-contained performance test with a fixed set of parameters.  For example, a Test Case might use JMeter in a Pod with an Istio-injected Envoy sidecar.  The Test Case may run a specific JMeter scenario for 10 minutes using two worker threads in Envoy, eight concurrent client connections, at a fixed rate of 100 HTTP/1.1 transactions per second (TPS).  A different Test Case for a Test might use Jmeter (with an Envoy sidecar) to run the same specific JMeter scenario for 10 minutes using four worker threads, sixteen client connections and a fixed rate of 1000 TPS.  Notice these two Test Cases use a common scenario, varying the TPS rate.  The worker threads and client connections are also varied, as perhaps is necessary to achieve the higher TPS rate.

For a Test Case, one might create a Kubernetes Namespace, start an NGINX Pod in that Namespace, created a shared PersistentVolumeClaim, run the JMeter Job for 10 minutes, run the JTL processor Job, run the prometheus collection Job, then extract all of this data into a local tar archive.  Each of these ("create Namespace", "create NGINX Pod", etc.) are steps of a common Pipeline.  All Test Cases share a common Pipeline (i.e., a common set of steps to complete the Test Case).  Pipeline steps are called Actions.

A Test Unit is a set of Test Cases for which a different set of variables is altered.  For example, one might want to run the two Test Cases previously described, but in one case, Envoy telemetry is employed, while in the other case, it is not.  This allows one to test the effect of enabling telemetry on CPU, memory and latency at various TPS levels.  To do this, one would create the two Test Cases above (which in turn, each use a common Pipeline with a common set of Pipeline Actions).  Then, one would create two Test Units, altering one variable ("use telemetry" or "do not use telemetry").  Notice that each Test Unit that is part of a Test execute the same set of Test Cases in the same order, but the variables set for the Test Unit may affect the nature of the Pipeline steps.  For example setting "do not use telemetry" may cause the creation of Telemetry resource, whereas, "use telemetry" does not (assuming here that the use of telemetry is the Istio default).

## Pipeline Actions

For this test pattern, one generally composes a Test Case from the following Actions:
1. Create a Kubernetes resource that is automatically deleted when the Test Case is done;
2. Execute an arbitrary script to do some action, where the script has knowledge about the current Test Case context (e.g., resources that have already been created for Test Case);
3. Transform supplied variables based on some condition during the Test Case.

When creating Kubernetes resources, there may be a dependency on other resources or contextual information.  For example, when standing up the NGINX Pod, the Pod itself is assigned an IP.  Imagine that the JMeter Pod must use an IP address of the server to target it.  The NGINX Pod IP is not known until the Pod is vivified, and this must be fed into the Job definition for the JMeter client.

A common way to provide this flexibility is to use a templating system like [golang templates](https://pkg.go.dev/text/template).  Indeed, go-templates is the basic templating system used by Helm.  Golang templates can be supplied variables used during the template expansion phase, as well as custom functions that can be used within the template.  `jobber` uses golang templates, and includes the [sprig](https://masterminds.github.io/sprig/) function set, as well as a few additional application-specific functions.  Helm also incorporates `sprig`, so the templating should be reasonably familiar to anybody that uses Helm.

In order to communicate context to a executable/script that is part of the Pipeline Actions, the contextual variables (the same set visible to the resource templates) is delivered as json text to the executable's stdin.  This makes it easy to process them using libraries (for executables built from languages like Golang or Python) or (reasonably standard) system tools like `jq`, when the executable is a shell script.

For the executable variant that is designed to transform the contextual variables for use by subsequent Pipeline Actions, the complete set of contextual variables is sent by the executable to its own stdout.  This is read by `jobber` and the current contextual variables are replaced by this set.

## Configuration File

`jobber` consumes a (YAML) configuration file.  This file defines the Pipeline Actions, each Test Case (including its Case-specific variables), each Test Unit (including its Unit-specific variables) and some additional Test scaffolding.

This is an example config file:

```yaml
Test:
  AssetArchive:
    FilePath: /opt/performance-testing/results/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      cgam_perf_test_nginx: 0.9.0
      file_extractor: 0.1.0
      jmeter_http2: 0.8.0
      jtl_processor: 0.5.0
      prometheus_collector: 0.5.2
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /opt/performance-testing/pipeline-actions-root
    ActionsInOrder:
      - resources/istio-cni.yaml
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - resources/shared-pvc.yaml
      - resources/jmeter-job.yaml
      - values-transforms/jmeter-post-job.sh
      - resources/jtl-processor-job.yaml
      - resources/prom-summary-job.yaml
      - resources/extractor.yaml
      - executables/extract-test-results.sh
  Cases:
  - Name: 100TPS
    Values:
      TPS: 100
      ConcurrentClientConnections: 8
      Sidecar:
        WorkerThreads: 2
  - Name: 1000TPS
    Values:
      TPS: 1000
      ConcurrentClientConnections: 16
      Sidecar:
        WorkerThreads: 4
  Units:
  - Name: NoTelemetry
    Values:
      Sidecar:
        Use:
          Telemetry: false
  - Name: WithTelemetry
    Values:
      Sidecar:
        Use:
          Telemetry: true

```

None of the `Values` maps have any special meaning to `jobber`.  It simply provides these to the Pipeline Actions.  The `Values` for `Global`, `Unit` and `Case` are orthogonal and when they are presented to Actions they are presented according to their context.  For example, in a resource template, this shows how `Values` from each context are accessed during template expansion:

- `jtl_processor` in Global: `{{ .Values.Global.jtl_processor }}`
- `Telemetry` in Unit: `{{ .Values.Unit.Sidecar.Use.Telemetry }}`
- `WorkerThreads` in Case: `{{ .Values.Case.Sidecar.WorkerThreads }}`

In this example, the `Sidecar` map in the Test Cases and the Test Units aren't merged together.  They are completely separate in the values hierarchy.  Among other things, this means that Test Case values cannot be overridden or defaulted by setting them in the Unit, even though each Unit executes all Test Cases.  `Values` are designed in this fashion to avoid accidental collisions that would be difficult to debug.

The Pipeline Actions are simple paths.  The first element of the path has a special meaning, and determines the type of action:

- `resources/`: templates that are expanded.  The expanded value must be well-formed YAML, and must be the YAML version of a Kubernetes resource (meaning, for example, that the YAML could be passed to `kubectl create`);
- `executables/`: arbitrary executables (which must be executable by the same user as the `jobber` process effective UID).  These are fed a text json blob containing `Values` and additional context;
- `values-transforms/`: also arbitrary executables that are fed a text blob containing `Values` and additional context.  The executables are expected to print to stdout `Values` and additional context, presumably changed.

The rest of the Action path provides an Action Target.  The directory specified in `.Test.Pipeline.ActionDefinitionsRootDirectory` should contain the Targets using the same layout as the Target descriptors.  That is, given the example configuration above, `jobber` expects the following files to exist:

```text
/
  opt/
    performance-testing/
      pipeline-actions-root/
        resources/
          istio-cni.yaml
          nginx-producer.yaml
          telemetry.yaml
          shared-pvc.yaml
          jmeter-job.yaml
          jtl-processor-job.yaml
          prom-summary-job.yaml
          extractor.yaml
        values-transforms/
          jmeter-post-job.sh
        executables/
          extract-test-results.sh
```

## Resolution of Action Targets

The `resources` Targets may be Jobber templates.  As noted above, when the (possible) template is expanded (that is, the double-curly substitutions are resolved), the result must be well-formed YAML.  The YAML must also be a well-formed Kubernetes resource definition.  If template expansion fails or the creation of the resource fails, the Test stops.  If template expansion is successful, `jobber` captures and records the expanded template string.  If a resource is created `jobber` keeps track of it.  When a Test Case completes, any Kubernetes resource that `jobber` created is deleted, one-by-one, in reverse order of creation.  For example, if a Test Case creates a Namespace, a Pod, a Job (called Job1) and another Job (called Job2) in that order, upon successful Pipeline completion for a Test Case, `jobber` will delete Job2 then Job1 then Pod then Namespace.  Failure to delete a resource will also terminate the Test.

The `executables` Targets are arbitrary executables.  As discussed variously above, the executable is fed values and context as a json blob to stdin.  If the executable exits with any non-zero value, the Test stops.  `jobber` records anything output to the executables stdout and stderr.

The `values-transforms` Targets are also arbitrary executables and also receive values and context as a json blob to stdin.  The executable is expected to emit the complete values and context set to stdout (with any intended modifications) as a json text blob.  This will completely replace the values and context for all remaining Actions in the current Test Case Pipeline.  If the executable exits with any non-zero value, the Test stops.  `jobber` records anything output to stdout (which, again, should be the modified values/context) and stderr.

During the execution of a Test, `jobber` creates a temporary directory (in the system temporary directory, usually `/tmp`).  Under this directory, it creates a directory with the same name as each Test Unit.  Under each of these Test Unit directories, it creates a directory with the same name as each Test Case.  Under each of these Test Case directories, it creates a directory for each Action Target type (i.e., `resources/`, `executables/` and `values-transforms/`).  Under these directories, it places the assets that are recorded from each Action taken.  Finally, each Test Case directory contains a directory called `retrieved-assets`.  `jobber` places nothing there, but provides the path to it as part of the context for each Pipeline Action.  As we will see, this temp directory is converted to a tarball, so this `retrieved-assets` directory is a sensible place for `executables` Targets to place any assets retrieved for a Test Case.

## Implied Actions

At the start of a Pipeline, a default Namespace is created.  Actions can use this Namespace or not (along with other Namespaces created as a `resources` Target), but this is done as a convenience.  The Namespace name is generated the prefix identified in the configuration as `.Test.DefaultNamespace.Basename`.  As with all other created resources, the default Namespace is deleted when a Test Case Pipeline successfully completes.

If there is a file called `default-namespace.yaml` under `resources` in the action definition root directory, this template is used to create the Namespace.  Normally, the `.metadata` section should contain neither a `name` nor a `generatedName`.  A `.metadata.GeneratedName` is automatically inserted after template expansion.

## Values and Context

`executables` and `values-transforms` Targets are provided the following json:

```json
{
  "Values": {
    "Global": {
      // Global Values map
    },
    "Unit": {
      // current Unit Values map
    },
    "Case": {
      // current Case Values map
    }
  },
  "Context": {
    "TestUnitName": "<test-unit-name>",
    "TestCaseName": "<test-case-name>",
    "TestCaseRetrievedAssetsDirectoryPath": "<path/to/tmproot/current-unit-name/current-case-name/retrieved-assets>"
  },
  "Runtime": {
    "DefaultNamespace": {
      "Name": "<default-namespace-name>"
    }
  }
}
```

For example, given the example configuration above, when the first Test Case of the first Test Unit runs, assume that the temp directory is `/tmp/jobber.55555`.  The json blob would look like this:

```json
{
  "Values": {
    "Global": {
      "ImageVersions": {
        "cgam_perf_test_nginx": "0.9.0",
        "file_extractor": "0.1.0",
        "jmeter_http2": "0.8.0",
        "jtl_processor": "0.5.0",
        "prometheus_collector": "0.5.2"
      },
      "TestCaseDurationInSeconds": 600,
      "PipelinePvc": {
        "StorageRequest": "3Gi"
      }
    },
    "Unit": {
      "Sidecar": {
        "Use": {
            "Telemetry": false
        }
      }
    },
    "Case": {
      "TPS": 100,
      "ConcurrentClientConnections": 8,
      "Sidecar": {
        "WorkerThreads": 2
      }
    }
  },
  "Context": {
    "TestUnitName": "NoTelemetry",
    "TestCaseName": "100TPS",
    "TestCaseRetrievedAssetsDirectoryPath": "/tmp/jobber.55555/NoTelemetry/100TPS/retrieved-assets"
  },
  "Runtime": {
    "DefaultNamespace": {
      "Name": "asm-perftest-3f5xd"
    }
  }
}
```

This same set is visible during template expansion, using the golang-template dot notation.  For example:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: jmeter-consumer-job
  labels:
    testRole: consumer
spec:
  backoffLimit: 0
  template:
    spec:
      activeDeadlineSeconds: {{ add .Values.Unit.TestDurationInSeconds 60 }}
      restartPolicy: Never
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - topologyKey: "kubernetes.io/hostname"
              labelSelector:
                matchExpressions:
                  - key: testRole
                    operator: In
                    values:
                      - producer
      volumes:
        - name: shared-pvc
          persistentVolumeClaim:
            claimName: shared-pipeline-pvc
      containers:
        - name: consumer
          image: f5vwells/jmeter-http2:{{ .Values.Global.ImageVersions.jmeter_http2 }}
          env:
            - name: USING_SIDECAR
              value: {{ .Values.Unit.Sidecar.Inject | quote }}
            - name: USE_BUILTIN_SCENARIO
              value: "SingleServerTarget-PreciseTPS-StaticResponses"
            - name: __SCENARIO_VAR__producerIPorHostname
              value: {{ .Runtime.CreatedPod "nginx-producer" | pod_ip_string | quote }}
            - name: __SCENARIO_VAR__producerPort
              value: "8080"
            - name: __SCENARIO_VAR__httpTransactionsPerSecond
              value: {{ .Values.Case.TPS | quote }}
            - name: __SCENARIO_VAR__numberOfConcurrentClientConnections
              value: {{ .Values.Case.ConcurrentClientConnections | quote }}
            - name: __SCENARIO_VAR__testDurationInSeconds
              value: {{ .Values.Global.TestCaseDurationInSeconds | quote }}
          args: ["-l", "/opt/test_results/jmeter.jtl.log", "-j", "/opt/test_results/jmeter.log"] 
          volumeMounts:
            - name: shared-pvc
              mountPath: /opt/test_results
```

## Expansion Template Custom Functions

In addition to the variable values supplied to templates during expansion, there are some additional custom functions that are available.  Specifically:

- `.Runtime.CreatedPod "<podname>" ["<namespace>"]`: returns an object representing the Pod with the name `<podname>`.  If `<namespace>` is provided, the Pod is looked up in that Namespace.  Otherwise, the default Namespace is used.  The returned object is intended to be passed to golang-template pipes.
- `.Runtime.ServiceAccount "<sa-name>" "<namespace>"`: returns an object representing the named ServiceAccount in the named Namespace.  The returned object is intended to be passed to golang-template pipes.
- `pod_ip_string`: accepts a `CreatedPod` object, and returns the current `.Status.PodIP` value.
- `bound_bearer_token`: accepts a `ServiceAccount` object and using that ServiceAccount, generates an API bearer token, returning it as a string.

## Test Archive File

Given the Test definition above, the following directories and files would be created as the Test proceeds:

```text
<system-tmp-dir>/
  jobber.<unique-extension>/
    NoTelemetry/
      100TPS/
        resources/
          istio-cni.yaml
          nginx-producer.yaml
          telemetry.yaml
          shared-pvc.yaml
          jmeter-job.yaml
          jtl-processor-job.yaml
          prom-summary-job.yaml
          extractor.yaml
        executables/
          extract-test-results.sh.stdout
          extract-test-results.sh.stderr
        values-transforms/
          jmeter-post-job.sh.stdout
          jmeter-post-job.sh.stderr
        retrieved-assets/
      500TPS/
        resources/
          istio-cni.yaml
          nginx-producer.yaml
          telemetry.yaml
          shared-pvc.yaml
          jmeter-job.yaml
          jtl-processor-job.yaml
          prom-summary-job.yaml
          extractor.yaml
        executables/
          extract-test-results.sh.stdout
          extract-test-results.sh.stderr
        values-transforms/
          jmeter-post-job.sh.stdout
          jmeter-post-job.sh.stderr
        retrieved-assets/
    WithTelemetry/
      100TPS/
        resources/
          istio-cni.yaml
          nginx-producer.yaml
          telemetry.yaml
          shared-pvc.yaml
          jmeter-job.yaml
          jtl-processor-job.yaml
          prom-summary-job.yaml
          extractor.yaml
        executables/
          extract-test-results.sh.stdout
          extract-test-results.sh.stderr
        values-transforms/
          jmeter-post-job.sh.stdout
          jmeter-post-job.sh.stderr
        retrieved-assets/
      500TPS/
        resources/
          istio-cni.yaml
          nginx-producer.yaml
          telemetry.yaml
          shared-pvc.yaml
          jmeter-job.yaml
          jtl-processor-job.yaml
          prom-summary-job.yaml
          extractor.yaml
        executables/
          extract-test-results.sh.stdout
          extract-test-results.sh.stderr
        values-transforms/
          jmeter-post-job.sh.stdout
          jmeter-post-job.sh.stderr
        retrieved-assets/
```

If the Test completes successfully (that is, if each Pipeline Action of each Test Case of each Test Unit completes without error), an `tar` and `gzip`ped file is created from this temp directory.  The root of the archive starts just above the Test Unit directories.  The file is written to a file with a name specified in `.Test.AssetArchive.FilePath`.  The name must be a file that does not match an existing directory path, and the directory containing it must exist and be writable by the effective UID of the running `jobber` instance.

Once the archive is created, the temp directory is deleted.

## Troubleshooting a Pipeline

The reason `jobber` records expanded templates, and stdout/stderr from executables is to facilitate Pipeline Action debugging.  Usually, a failure of Pipeline Action occurs because of a bug in the Action definition (e.g., a resource template that contains a non-existant `Values` reference or which yields YAML that is not correct for a resource type).  When an Action fails, the Test stops.  At this point, the creator of the Pipeline can look at the still-existing temp directory contents to help determine what happened.  It is a good idea to remove the temp directory manually when troubleshooting is done.  If a Test terminates on an error, any resources already created for the last running Test Case will still exist.  These, too, should be manually deleted.

## Logging

`jobber` prints a stream of events to stdout in human-readable format.  Among other things, every directory, file and resource that are created is logged, including paths and names.  Errors that terminate a Test are also logged.  This logging allows the user to locate the still-existing temp directory, any still-existing resources, and the error that caused termination.

## Building Jobber

To build the `jobber` application, you must be on a system with [golang](https://go.dev/doc/install) 1.21 or higher.  Do the following:

### Clone the repository

```bash
git clone https://github.com/blorticus-go/jobber.git
```

### Build the application

```bash
cd jobber/applications/jobber
go build -o /tmp/jobber .
```

The executable is now `/tmp/jobber`.  Naturally, you may deposit anywhere you choose.

To run `jobber`, there must a [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file with appropriate kube-api access defined for the cluster you will target.  If the environmental variable `KUBECONFIG` is defined and points to a valid kubeconfig file, you don't need to do anything else.  Alternatively, you can point directly to a kubeconfig file using the `-kubeconfig` flag (followed by the path to a kubeconfig file).

`jobber` requires a properly formatted test configuration file as described above, and this file must reference a properly arranged action definition root directory.  The default location for the config file is `./config.yaml`.  To specific a different config file (and you really should), pass the `-config` flag (followed by the path to the configuration file).

`jobber` will log to stdout as described above.