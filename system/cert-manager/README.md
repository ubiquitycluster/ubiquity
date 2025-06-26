# cert-manager

cert-manager is a Kubernetes addon to automate the management and issuance of TLS certificates from various issuing sources.

It will ensure certificates are valid and up to date periodically, and attempt to renew certificates at an appropriate time before expiry.

## Prerequisites
- Kubernetes 1.20+

## Installing the Chart
Full installation instructions, including details on how to configure extra functionality in cert-manager can be found in the .

Before installing the chart, you must first install the cert-manager CustomResourceDefinition resources. This is performed in a separate step to allow you to easily uninstall and reinstall cert-manager without deleting your installed custom resources.
```
$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.crds.yaml
To install the chart with the release name my-release:

## Add the Jetstack Helm repository
$ helm repo add jetstack https://charts.jetstack.io

## Install the cert-manager helm chart
$ helm install my-release --namespace cert-manager --version v1.13.2 jetstack/cert-manager
```
In order to begin issuing certificates, you will need to set up a ClusterIssuer or Issuer resource (for example, by creating a 'letsencrypt-staging' issuer).

More information on the different types of issuers and how to configure them can be found in .

For information on how to configure cert-manager to automatically provision Certificates for Ingress resources, take a look at the .

Tip: List all releases using helm list

## Upgrading the Chart
Special considerations may be required when upgrading the Helm chart, and these are documented in our full .

Please check here before performing upgrades!

## Uninstalling the Chart
To uninstall/delete the my-release deployment:

$ helm delete my-release
The command removes all the Kubernetes components associated with the chart and deletes the release.

If you want to completely uninstall cert-manager from your cluster, you will also need to delete the previously installed CustomResourceDefinition resources:

$ kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.crds.yaml
Configuration
The following table lists the configurable parameters of the cert-manager chart and their default values.

|Parameter	|Description	|Default
|---|---|---
global.imagePullSecrets	|Reference to one or more secrets to be used when pulling images	|[]
global.commonLabels	|Labels to apply to all resources	|{}
global.rbac.create	|If true, create and use RBAC resources (includes sub-charts)	|true
global.priorityClassName	|Priority class name for cert-manager and webhook pods	|""
global.podSecurityPolicy.enabled	|If true, create and use PodSecurityPolicy (includes sub-charts)	|false
global.podSecurityPolicy.useAppArmor	|If true, use Apparmor seccomp profile in PSP	|true
global.leaderElection.namespace	|Override the namespace used to store the ConfigMap for leader election	|kube-system
global.leaderElection.leaseDuration	|The duration that non-leader candidates will wait after observing a leadership renewal until attempting to acquire leadership of a led but unrenewed leader slot. This is effectively the maximum duration that a leader can be stopped before it is replaced by another candidate	|
global.leaderElection.renewDeadline	|The interval between attempts by the acting master to renew a leadership slot before it stops leading. This must be less than or equal to the lease duration	|
global.leaderElection.retryPeriod	|The duration the clients should wait between attempting acquisition and renewal of a leadership	|
installCRDs	|If true, CRD resources will be installed as part of the Helm chart. If enabled, when uninstalling CRD resources will be deleted causing all installed custom resources to be DELETED	|false
image.repository	|Image repository	|quay.io/jetstack/cert-manager-controller
image.tag	|Image tag	|v1.13.2
image.pullPolicy	|Image pull policy	|IfNotPresent
replicaCount	|Number of cert-manager replicas	|1
clusterResourceNamespace	|Override the namespace used to store DNS provider credentials etc. for ClusterIssuer resources	|Same namespace as cert-manager pod
featureGates	|Set of comma-separated key=value pairs that describe feature gates on the controller. Some feature gates may also have to be enabled on other components, and can be set supplying the feature-gate flag to <component>.extraArgs	|``
extraArgs	|Optional flags for cert-manager	|[]
extraEnv	|Optional environment variables for cert-manager	|[]
serviceAccount.create	|If true, create a new service account	|true
serviceAccount.name	|Service account to be used. If not set and serviceAccount.create is true, a name is generated using the fullname template	|
serviceAccount.annotations	|Annotations to add to the service account	|
serviceAccount.automountServiceAccountToken	|Automount API credentials for the Service Account	|true
volumes	|Optional volumes for cert-manager	|[]
volumeMounts	|Optional volume mounts for cert-manager	|[]
resources	|CPU/memory resource requests/limits	|{}
securityContext	|Security context for the controller pod assignment	|refer to
containerSecurityContext	|Security context to be set on the controller component container	|refer to
nodeSelector	|Node labels for pod assignment	|{}
affinity	|Node affinity for pod assignment	|{}
tolerations	|Node tolerations for pod assignment	|[]
topologySpreadConstraints	|Topology spread constraints for pod assignment	|[]
livenessProbe.enabled	|Enable or disable the liveness probe for the controller container in the controller Pod. See  to learn about when you might want to enable this livenss probe.	|false
livenessProbe.initialDelaySeconds	|The liveness probe initial delay (in seconds)	|10
livenessProbe.periodSeconds	|The liveness probe period (in seconds)	|10
livenessProbe.timeoutSeconds	|The liveness probe timeout (in seconds)	|10
livenessProbe.periodSeconds	|The liveness probe period (in seconds)	|10
livenessProbe.successThreshold	|The liveness probe success threshold	|1
livenessProbe.failureThreshold	|The liveness probe failure threshold	|8
ingressShim.defaultIssuerName	|Optional default issuer to use for ingress resources	|
ingressShim.defaultIssuerKind	|Optional default issuer kind to use for ingress resources	|
ingressShim.defaultIssuerGroup	|Optional default issuer group to use for ingress resources	|
prometheus.enabled	|Enable Prometheus monitoring	|true
prometheus.servicemonitor.enabled	|Enable Prometheus Operator ServiceMonitor monitoring	|false
prometheus.servicemonitor.namespace	|Define namespace where to deploy the ServiceMonitor resource	|(namespace where you are deploying)
prometheus.servicemonitor.prometheusInstance	|Prometheus Instance definition	|default
prometheus.servicemonitor.targetPort	|Prometheus scrape port	|9402
prometheus.servicemonitor.path	|Prometheus scrape path	|/metrics
prometheus.servicemonitor.interval	|Prometheus scrape interval	|60s
prometheus.servicemonitor.labels	|Add custom labels to ServiceMonitor	|
prometheus.servicemonitor.scrapeTimeout	|Prometheus scrape timeout	|30s
prometheus.servicemonitor.honorLabels	|Enable label honoring for metrics scraped by Prometheus (see  for details). By setting honorLabels to true, Prometheus will prefer label contents given by cert-manager on conflicts. Can be used to remove the "exported_namespace" label for example.	|false
podAnnotations	|Annotations to add to the cert-manager pod	|{}
deploymentAnnotations	|Annotations to add to the cert-manager deployment	|{}
podDisruptionBudget.enabled	|Adds a PodDisruptionBudget for the cert-manager deployment	|false
podDisruptionBudget.minAvailable	|Configures the minimum available pods for voluntary disruptions. Cannot used if maxUnavailable is set.	|1
podDisruptionBudget.maxUnavailable	|Configures the maximum unavailable pods for voluntary disruptions. Cannot used if minAvailable is set.	|
podDnsPolicy	|Optional cert-manager pod 	|
podDnsConfig	|Optional cert-manager pod 	|
podLabels	|Labels to add to the cert-manager pod	|{}
serviceLabels	|Labels to add to the cert-manager controller service	|{}
serviceAnnotations	|Annotations to add to the cert-manager service	|{}
http_proxy	|Value of the HTTP_PROXY environment variable in the cert-manager pod	|
https_proxy	|Value of the HTTPS_PROXY environment variable in the cert-manager pod	|
no_proxy	|Value of the NO_PROXY environment variable in the cert-manager pod	|
dns01RecursiveNameservers	|Comma separated string with host and port of the recursive nameservers cert-manager should query	|``
dns01RecursiveNameserversOnly	|Forces cert-manager to only use the recursive nameservers for verification.	|false
enableCertificateOwnerRef	|When this flag is enabled, secrets will be automatically removed when the certificate resource is deleted	|false
config	|ControllerConfiguration YAML used to configure flags for the controller. Generates a ConfigMap containing contents of the field. See values.yaml for example.	|{}
enableServiceLinks	|Indicates whether information about services should be injected into pod's environment variables, matching the syntax of Docker links.	|false
webhook.replicaCount	|Number of cert-manager webhook replicas	|1
webhook.timeoutSeconds	|Seconds the API server should wait the webhook to respond before treating the call as a failure.	|10
webhook.podAnnotations	|Annotations to add to the webhook pods	|{}
webhook.podLabels	|Labels to add to the cert-manager webhook pod	|{}
webhook.serviceLabels	|Labels to add to the cert-manager webhook service	|{}
webhook.deploymentAnnotations	|Annotations to add to the webhook deployment	|{}
webhook.podDisruptionBudget.enabled	|Adds a PodDisruptionBudget for the cert-manager deployment	|false
webhook.podDisruptionBudget.minAvailable	|Configures the minimum available pods for voluntary disruptions. Cannot used if maxUnavailable is set.	|1
webhook.podDisruptionBudget.maxUnavailable	|Configures the maximum unavailable pods for voluntary disruptions. Cannot used if minAvailable is set.	|
webhook.mutatingWebhookConfigurationAnnotations	|Annotations to add to the mutating webhook configuration	|{}
webhook.validatingWebhookConfigurationAnnotations	|Annotations to add to the validating webhook configuration	|{}
webhook.serviceAnnotations	|Annotations to add to the webhook service	|{}
webhook.config	|WebhookConfiguration YAML used to configure flags for the webhook. Generates a ConfigMap containing contents of the field. See values.yaml for example.	|{}
webhook.extraArgs	|Optional flags for cert-manager webhook component	|[]
webhook.serviceAccount.create	|If true, create a new service account for the webhook component	|true
webhook.serviceAccount.name	|Service account for the webhook component to be used. If not set and webhook.serviceAccount.create is true, a name is generated using the fullname template	|
webhook.serviceAccount.annotations	|Annotations to add to the service account for the webhook component	|
webhook.serviceAccount.automountServiceAccountToken	|Automount API credentials for the webhook Service Account	|
webhook.resources	|CPU/memory resource requests/limits for the webhook pods	|{}
webhook.nodeSelector	|Node labels for webhook pod assignment	|{}
webhook.networkPolicy.enabled	|Enable default network policies for webhooks egress and ingress traffic	|false
webhook.networkPolicy.ingress	|Sets ingress policy block. See NetworkPolicy documentation. See values.yaml for example.	|{}
webhook.networkPolicy.egress	|Sets ingress policy block. See NetworkPolicy documentation. See values.yaml for example.	|{}
webhook.affinity	|Node affinity for webhook pod assignment	|{}
webhook.tolerations	|Node tolerations for webhook pod assignment	|[]
webhook.topologySpreadConstraints	|Topology spread constraints for webhook pod assignment	|[]
webhook.image.repository	|Webhook image repository	|quay.io/jetstack/cert-manager-webhook
webhook.image.tag	|Webhook image tag	|v1.13.2
webhook.image.pullPolicy	|Webhook image pull policy	|IfNotPresent
webhook.image.pullSecrets	|Webhook image pull secrets	|[]
webhook.service.type	|Webhook service type	|ClusterIP
webhook.service.port	|Webhook service port	|443
webhook.service.targetPort	|Webhook service target port	|9876
webhook.service.nodePort	|Webhook service node port	|
webhook.service.loadBalancerIP	|Webhook service load balancer IP	|
webhook.service.loadBalancerSourceRanges	|Webhook service load balancer source ranges	|[]
webhook.service.externalTrafficPolicy	|Webhook service external traffic policy	|Cluster
webhook.service.annotations	|Annotations to add to the webhook service	|{}
webhook.service.labels	|Labels to add to the webhook service	|{}
webhook.service.extraPorts	|Extra ports to expose on the webhook service	|[]
webhook.service.extraIPs	|Extra IP addresses to expose on the webhook service	|[]
webhook.service.extraHosts	|Extra hostnames to expose on the webhook service	|[]
webhook.service.extraTLS	|Extra TLS configuration for the webhook service	|[]
webhook.serviceAccount.extraAnnotations	|Extra annotations to add to the webhook service account	|{}
webhook.podDisruptionBudget.enabled	|Adds a PodDisruptionBudget for the webhook deployment	|false
webhook.podDisruptionBudget.minAvailable	|Configures the minimum available pods for voluntary disruptions. Cannot used if maxUnavailable is set.	|1
webhook.podDisruptionBudget.maxUnavailable	|Configures the maximum unavailable pods for voluntary disruptions. Cannot used if minAvailable is set.	|
webhook.livenessProbe.enabled	|Enable or disable the liveness probe for the webhook container in the webhook Pod. See  to learn about when you might want to enable this livenss probe.	|false
webhook.livenessProbe.initialDelaySeconds	|The liveness probe initial delay (in seconds)	|10
webhook.livenessProbe.periodSeconds	|The liveness probe period (in seconds)	|10
webhook.livenessProbe.timeoutSeconds	|The liveness probe timeout (in seconds)	|10
webhook.livenessProbe.periodSeconds	|The liveness probe period (in seconds)	|10
webhook.livenessProbe.successThreshold	|The liveness probe success threshold	|1
webhook.livenessProbe.failureThreshold	|The liveness probe failure threshold	|8
webhook.readinessProbe.enabled	|Enable or disable the readiness probe for the webhook container in the webhook Pod. See  to learn about when you might want to enable this readiness probe.	|false
webhook.readinessProbe.initialDelaySeconds	|The readiness probe initial delay (in seconds)	|10
webhook.readinessProbe.periodSeconds	|The readiness probe period (in seconds)	|10
webhook.readinessProbe.timeoutSeconds	|The readiness probe timeout (in seconds)	|10
webhook.readinessProbe.periodSeconds	|The readiness probe period (in seconds)	|10
webhook.readinessProbe.successThreshold	|The readiness probe success threshold	|1
webhook.readinessProbe.failureThreshold	|The readiness probe failure threshold	|8
webhook.enableServiceLinks	|Indicates whether information about services should be injected into pod's environment variables, matching the syntax of Docker links.	|false
caInjector.enabled	|Enable the ca-injector component	|true
caInjector.replicaCount	|Number of ca-injector replicas	|1
caInjector.podAnnotations	|Annotations to add to the ca-injector pods	|{}
caInjector.podLabels	|Labels to add to the ca-injector pods	|{}
caInjector.deploymentAnnotations	|Annotations to add to the ca-injector deployment	|{}
caInjector.podDisruptionBudget.enabled	|Adds a PodDisruptionBudget for the ca-injector deployment	|false
caInjector.podDisruptionBudget.minAvailable	|Configures the minimum available pods for voluntary disruptions. Cannot used if maxUnavailable is set.	|1
caInjector.podDisruptionBudget.maxUnavailable	|Configures the maximum unavailable pods for voluntary disruptions. Cannot used if minAvailable is set.	|
caInjector.extraArgs	|Optional flags for cert-manager ca-injector component	|[]
caInjector.serviceAccount.create	|If true, create a new service account for the ca-injector component	|true
caInjector.serviceAccount.name	|Service account for the ca-injector component to be used. If not set and caInjector.serviceAccount.create is true, a name is generated using the fullname template	|
caInjector.serviceAccount.annotations	|Annotations to add to the service account for the ca-injector component	|
caInjector.serviceAccount.automountServiceAccountToken	|Automount API credentials for the ca-injector Service Account	|   true
caInjector.resources	|CPU/memory resource requests/limits for the ca-injector pods	|{}
caInjector.nodeSelector	|Node labels for ca-injector pod assignment	|{}
caInjector.affinity	|Node affinity for ca-injector pod assignment	|{}
caInjector.tolerations	|Node tolerations for ca-injector pod assignment	|[]
caInjector.topologySpreadConstraints	|Topology spread constraints for ca-injector pod assignment	|[]
caInjector.image.repository	|ca-injector image repository	|quay.io/jetstack/cert-manager-cainjector
caInjector.image.tag	|ca-injector image tag	|v1.13.2
caInjector.image.pullPolicy	|ca-injector image pull policy	|IfNotPresent
caInjector.image.securityContext	|ca-injector image security context	|refer to Default Security Contexts
caInjector.image.containerSecurityContext	|ca-injector image security context	|refer to Default Security Contexts
caInjector.enableServiceLinks	|Indicates whether information about services should be injected into pod's environment variables, matching the syntax of Docker links.	|false
acmesolver.image.repository	|acmesolver image repository	|quay.io/jetstack/cert-manager-acmesolver
acmesolver.image.tag	|acmesolver image tag	|v1.13.2
acmesolver.image.pullPolicy	|acmesolver image pull policy	|IfNotPresent
startupapicheck.enabled	|Enable the startupapicheck component	|true
startupapicheck.securityContext	|startupapicheck security context	|refer to Default Security Contexts
startupapicheck.containerSecurityContext	|startupapicheck security context	|refer to Default Security Contexts
startupapicheck.timeout | Timeout for the startupapicheck command 'kubectl check api' | 1m
startupapicheck.backoffLimit | Number of retries for the startupapicheck command 'kubectl check api' | 4
startupapicheck.jobannotations | Annotations to add to the startupapicheck job | {}
startupapicheck.podAnnotations	|Annotations to add to the startupapicheck pods	|{}
startupapicheck.extraArgs    |Optional flags for cert-manager startupapicheck component	|[]
startupapicheck.resources	|CPU/memory resource requests/limits for the startupapicheck pods	|{}
startupapicheck.nodeSelector	|Node labels for startupapicheck pod assignment	|{}
startupapicheck.affinity	|Node affinity for startupapicheck pod assignment	|{}
startupapicheck.tolerations	|Node tolerations for startupapicheck pod assignment	|[]
startupapicheck.podLabels	|Labels to add to the startupapicheck pods	|{}
startupapicheck.image.repository	|startupapicheck image repository	|quay.io/jetstack/cert-manager-ctl
startupapicheck.image.tag	|startupapicheck image tag	|v1.13.2
startupapicheck.image.pullPolicy	|startupapicheck image pull policy	|IfNotPresent
startupapicheck.serviceAccount.create	|If true, create a new service account for the startupapicheck component	|true
startupapicheck.serviceAccount.name	|Service account for the startupapicheck component to be used. If not set and startupapicheck.serviceAccount.create is true, a name is generated using the fullname template	|
startupapicheck.serviceAccount.annotations	|Annotations to add to the service account for the startupapicheck component	|
startupapicheck.serviceAccount.automountServiceAccountToken	|Automount API credentials for the startupapicheck Service Account	|true
startupapicheck.enableServiceLinks	|Indicates whether information about services should be injected into pod's environment variables, matching the syntax of Docker links.	|false
maxConcurrentChallenges	|Maximum number of concurrent challenges per Issuer	|60


## Default Security Contexts
The default pod-level and container-level security contexts, below, adhere to the  Pod Security Standards policies.

Default pod-level securityContext:
```
runAsNonRoot: true
seccompProfile:
  type: RuntimeDefault
```

Default containerSecurityContext:
```
allowPrivilegeEscalation: false
capabilities:
  drop:
  - ALL
```
## Assigning Values
Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,
```
$ helm install my-release -f values.yaml .
```
Tip: You can use the default `values.yaml`

## Contributing
This chart is maintained at [github.com/jetstack/cert-manager](github.com/jetstack/cert-manager). Please direct all issues and PRs there.