package generated

import (
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
	"testing"
)

func TestParse(t *testing.T) {
	var deployment = `
config:
  igor:
    service:
      kind: Service
      apiVersion: v1
      metadata:
        name: spin-igor
        namespace: nc5
        labels:
          app: spin
          cluster: spin-igor
      spec:
        selector:
          app: spin
          cluster: spin-igor
        type: ClusterIP
        ports:
        - protocol: TCP
          port: 8088
          targetPort: 8088
    resources:
    - apiVersion: v1
      kind: Secret
      metadata:
        name: spin-igor-files-817732999
        namespace: nc5
        labels:
          app: spin
          cluster: spin-igor
      type: Opaque
      data:
        spinnaker.yml: IyMgV0FSTklORwojIyBUaGlzIGZpbGUgd2FzIGF1dG9nZW5lcmF0ZWQsIGFuZCBfd2lsbF8gYmUgb3ZlcndyaXR0ZW4gYnkgSGFseWFyZC4KIyMgQW55IGVkaXRzIHlvdSBtYWtlIGhlcmUgX3dpbGxfIGJlIGxvc3QuCgpzZXJ2aWNlczoKICBjbG91ZGRyaXZlcjoKICAgIGhvc3Q6IDAuMC4wLjAKICAgIHBvcnQ6IDcwMDIKICAgIGJhc2VVcmw6IGh0dHA6Ly9zcGluLWNsb3VkZHJpdmVyLm5jNTo3MDAyCiAgICBlbmFibGVkOiB0cnVlCiAgY2xvdWRkcml2ZXJDYWNoaW5nOgogICAgaG9zdDogMC4wLjAuMAogICAgcG9ydDogNzAwMgogICAgYmFzZVVybDogaHR0cDovL3NwaW4tY2xvdWRkcml2ZXItY2FjaGluZy5uYzU6NzAwMgogICAgZW5hYmxlZDogZmFsc2UKICBjbG91ZGRyaXZlclJvOgogICAgaG9zdDogMC4wLjAuMAogICAgcG9ydDogNzAwMgogICAgYmFzZVVybDogaHR0cDovL3NwaW4tY2xvdWRkcml2ZXItcm8ubmM1OjcwMDIKICAgIGVuYWJsZWQ6IGZhbHNlCiAgY2xvdWRkcml2ZXJSb0RlY2s6CiAgICBob3N0OiAwLjAuMC4wCiAgICBwb3J0OiA3MDAyCiAgICBiYXNlVXJsOiBodHRwOi8vc3Bpbi1jbG91ZGRyaXZlci1yby1kZWNrLm5jNTo3MDAyCiAgICBlbmFibGVkOiBmYWxzZQogIGNsb3VkZHJpdmVyUnc6CiAgICBob3N0OiAwLjAuMC4wCiAgICBwb3J0OiA3MDAyCiAgICBiYXNlVXJsOiBodHRwOi8vc3Bpbi1jbG91ZGRyaXZlci1ydy5uYzU6NzAwMgogICAgZW5hYmxlZDogZmFsc2UKICBkZWNrOgogICAgaG9zdDogMC4wLjAuMAogICAgcG9ydDogOTAwMAogICAgYmFzZVVybDogaHR0cDovL25jNS5kZXYuYXJtb3J5LmlvCiAgICBlbmFibGVkOiB0cnVlCiAgZWNobzoKICAgIGhvc3Q6IDAuMC4wLjAKICAgIHBvcnQ6IDgwODkKICAgIGJhc2VVcmw6IGh0dHA6Ly9zcGluLWVjaG8ubmM1OjgwODkKICAgIGVuYWJsZWQ6IHRydWUKICBlY2hvU2NoZWR1bGVyOgogICAgaG9zdDogMC4wLjAuMAogICAgcG9ydDogODA4OQogICAgYmFzZVVybDogaHR0cDovL3NwaW4tZWNoby1zY2hlZHVsZXIubmM1OjgwODkKICAgIGVuYWJsZWQ6IGZhbHNlCiAgZWNob1dvcmtlcjoKICAgIGhvc3Q6IDAuMC4wLjAKICAgIHBvcnQ6IDgwODkKICAgIGJhc2VVcmw6IGh0dHA6Ly9zcGluLWVjaG8td29ya2VyLm5jNTo4MDg5CiAgICBlbmFibGVkOiBmYWxzZQogIGZpYXQ6CiAgICBob3N0OiAwLjAuMC4wCiAgICBwb3J0OiA3MDAzCiAgICBiYXNlVXJsOiBodHRwOi8vc3Bpbi1maWF0Lm5jNTo3MDAzCiAgICBlbmFibGVkOiBmYWxzZQogIGZyb250NTA6CiAgICBob3N0OiAwLjAuMC4wCiAgICBwb3J0OiA4MDgwCiAgICBiYXNlVXJsOiBodHRwOi8vc3Bpbi1mcm9udDUwLm5jNTo4MDgwCiAgICBlbmFibGVkOiB0cnVlCiAgZ2F0ZToKICAgIGhvc3Q6IDAuMC4wLjAKICAgIHBvcnQ6IDgwODQKICAgIGJhc2VVcmw6IGh0dHA6Ly9uYzUtYXBpLmRldi5hcm1vcnkuaW8KICAgIGVuYWJsZWQ6IHRydWUKICBpZ29yOgogICAgaG9zdDogMC4wLjAuMAogICAgcG9ydDogODA4OAogICAgYmFzZVVybDogaHR0cDovL3NwaW4taWdvci5uYzU6ODA4OAogICAgZW5hYmxlZDogdHJ1ZQogIGtheWVudGE6CiAgICBob3N0OiAwLjAuMC4wCiAgICBwb3J0OiA4MDkwCiAgICBiYXNlVXJsOiBodHRwOi8vc3Bpbi1rYXllbnRhLm5jNTo4MDkwCiAgICBlbmFibGVkOiBmYWxzZQogIG9yY2E6CiAgICBob3N0OiAwLjAuMC4wCiAgICBwb3J0OiA4MDgzCiAgICBiYXNlVXJsOiBodHRwOi8vc3Bpbi1vcmNhLm5jNTo4MDgzCiAgICBlbmFibGVkOiB0cnVlCiAgcmVkaXM6CiAgICBob3N0OiAwLjAuMC4wCiAgICBwb3J0OiA2Mzc5CiAgICBiYXNlVXJsOiByZWRpczovL3NwaW4tcmVkaXMubmM1OjYzNzkKICAgIGVuYWJsZWQ6IHRydWUKICByb3NjbzoKICAgIGhvc3Q6IDAuMC4wLjAKICAgIHBvcnQ6IDgwODcKICAgIGJhc2VVcmw6IGh0dHA6Ly9zcGluLXJvc2NvLm5jNTo4MDg3CiAgICBlbmFibGVkOiB0cnVlCiAgbW9uaXRvcmluZ0RhZW1vbjoKICAgIGhvc3Q6IDAuMC4wLjAKICAgIHBvcnQ6IDgwMDgKICAgIGJhc2VVcmw6IGh0dHA6Ly9zcGluLW1vbml0b3JpbmctZGFlbW9uLm5jNTo4MDA4CiAgICBlbmFibGVkOiB0cnVlCgpnbG9iYWwuc3Bpbm5ha2VyLnRpbWV6b25lOiBBbWVyaWNhL0xvc19BbmdlbGVz
        igor.yml: IyMgV0FSTklORwojIyBUaGlzIGZpbGUgd2FzIGF1dG9nZW5lcmF0ZWQsIGFuZCBfd2lsbF8gYmUgb3ZlcndyaXR0ZW4gYnkgSGFseWFyZC4KIyMgQW55IGVkaXRzIHlvdSBtYWtlIGhlcmUgX3dpbGxfIGJlIGxvc3QuCgpzcGVjdGF0b3I6CiAgYXBwbGljYXRpb25OYW1lOiAke3NwcmluZy5hcHBsaWNhdGlvbi5uYW1lfQogIHdlYkVuZHBvaW50OgogICAgZW5hYmxlZDogdHJ1ZQoKZG9ja2VyUmVnaXN0cnkuZW5hYmxlZDogdHJ1ZQphcnRpZmFjdHM6CiAgYml0YnVja2V0OgogICAgZW5hYmxlZDogdHJ1ZQogICAgYWNjb3VudHM6CiAgICAtIG5hbWU6IHRlc3QKICAgICAgdXNlcm5hbWU6IG5pY2tvbmV0K3Rlc3RAZ21haWwuY29tCiAgICAgIHBhc3N3b3JkOiB3ZzNFbkwyS3pUWEFMdmRNeW56NzM2d2hwQWRncDN3WQogIGdjczoKICAgIGVuYWJsZWQ6IGZhbHNlCiAgICBhY2NvdW50czogW10KICBvcmFjbGU6CiAgICBlbmFibGVkOiBmYWxzZQogICAgYWNjb3VudHM6IFtdCiAgZ2l0aHViOgogICAgZW5hYmxlZDogZmFsc2UKICAgIGFjY291bnRzOiBbXQogIGdpdGxhYjoKICAgIGVuYWJsZWQ6IGZhbHNlCiAgICBhY2NvdW50czogW10KICBodHRwOgogICAgZW5hYmxlZDogZmFsc2UKICAgIGFjY291bnRzOiBbXQogIGhlbG06CiAgICBlbmFibGVkOiBmYWxzZQogICAgYWNjb3VudHM6IFtdCiAgczM6CiAgICBlbmFibGVkOiBmYWxzZQogICAgYWNjb3VudHM6IFtdCiAgbWF2ZW46CiAgICBlbmFibGVkOiBmYWxzZQogICAgYWNjb3VudHM6IFtdCiAgdGVtcGxhdGVzOiBbXQoKYXJ0aWZhY3Rvcnk6CiAgZW5hYmxlZDogZmFsc2UKICBzZWFyY2hlczogW10KCmplbmtpbnM6CiAgZW5hYmxlZDogZmFsc2UKICBtYXN0ZXJzOiBbXQp0cmF2aXM6CiAgZW5hYmxlZDogZmFsc2UKICBtYXN0ZXJzOiBbXQp3ZXJja2VyOgogIGVuYWJsZWQ6IGZhbHNlCiAgbWFzdGVyczogW10KY29uY291cnNlOgogIGVuYWJsZWQ6IGZhbHNlCiAgbWFzdGVyczogW10KZ2NiOgogIGVuYWJsZWQ6IGZhbHNlCiAgYWNjb3VudHM6IFtdCgpzZXJ2ZXI6CiAgcG9ydDogJHtzZXJ2aWNlcy5pZ29yLnBvcnQ6ODA4OH0KICBhZGRyZXNzOiAke3NlcnZpY2VzLmlnb3IuaG9zdDpsb2NhbGhvc3R9CgpyZWRpczoKICBjb25uZWN0aW9uOiAke3NlcnZpY2VzLnJlZGlzLmJhc2VVcmw6cmVkaXM6Ly9sb2NhbGhvc3Q6NjM3OX0K
    - apiVersion: v1
      kind: Secret
      metadata:
        name: spin-igor-files-1144163964
        namespace: nc5
        labels:
          app: spin
          cluster: spin-igor
      type: Opaque
      data:
        spinnaker-monitoring.yml: IyMgV0FSTklORwojIyBUaGlzIGZpbGUgd2FzIGF1dG9nZW5lcmF0ZWQsIGFuZCBfd2lsbF8gYmUgb3ZlcndyaXR0ZW4gYnkgSGFseWFyZC4KIyMgQW55IGVkaXRzIHlvdSBtYWtlIGhlcmUgX3dpbGxfIGJlIGxvc3QuCgpkYXRhZG9nOgogIGVuYWJsZWQ6IGZhbHNlCiAgdGFnczogW10KcHJvbWV0aGV1czoKICBlbmFibGVkOiB0cnVlCiAgYWRkX3NvdXJjZV9tZXRhbGFiZWxzOiB0cnVlCnN0YWNrZHJpdmVyOgogIGVuYWJsZWQ6IGZhbHNlCnBlcmlvZDogMzAKZW5hYmxlZDogdHJ1ZQoKc2VydmVyOgogIGhvc3Q6IDAuMC4wLjAKICBwb3J0OiA4MDA4Cgptb25pdG9yOgogIHBlcmlvZDogMzAKICBtZXRyaWNfc3RvcmU6CiAgLSBwcm9tZXRoZXVzCgojIGhhbGNvbmZpZyAK
    deployment:
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: spin-igor
        namespace: nc5
        annotations:
          moniker.spinnaker.io/application: '"spin"'
          moniker.spinnaker.io/cluster: '"igor"'
        labels:
          app: spin
          cluster: spin-igor
          app.kubernetes.io/name: igor
          app.kubernetes.io/managed-by: halyard
          app.kubernetes.io/part-of: spinnaker
          app.kubernetes.io/version: 1.14.2
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: spin
            cluster: spin-igor
        template:
          metadata:
            annotations: {}
            labels:
              app: spin
              cluster: spin-igor
              app.kubernetes.io/name: igor
              app.kubernetes.io/managed-by: halyard
              app.kubernetes.io/part-of: spinnaker
              app.kubernetes.io/version: 1.14.2
          spec:
            containers:
            - name: igor
              image: gcr.io/spinnaker-marketplace/igor:1.3.0-20190515102735
              ports:
              - containerPort: 8088
              readinessProbe:
                exec:
                  command:
                  - wget
                  - --no-check-certificate
                  - --spider
                  - -q
                  - http://localhost:8088/health
                initialDelaySeconds: null
              livenessProbe: null
              securityContext: null
              volumeMounts:
              - name: spin-igor-files-817732999
                mountPath: /opt/spinnaker/config
              - name: spin-igor-files-1144163964
                mountPath: /opt/spinnaker-monitoring/config
              - name: spin-igor-files-1766711109
                mountPath: /opt/spinnaker-monitoring/registry
              lifecycle: {}
              env:
              - name: JAVA_OPTS
                value: -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap
                  -XX:MaxRAMFraction=2
              - name: SPRING_PROFILES_ACTIVE
                value: local
              resources:
                requests: {}
                limits: {}
            - name: monitoring-daemon
              image: gcr.io/spinnaker-marketplace/monitoring-daemon:0.13.0-20190430163248
              ports:
              - containerPort: 8008
              readinessProbe:
                tcpSocket:
                  port: 8008
                initialDelaySeconds: null
              livenessProbe: null
  `
	g := &SpinnakerGeneratedConfig{}
	err := yaml.Unmarshal([]byte(deployment), g)
	if assert.Nil(t, err) {
		if assert.Equal(t, 1, len(g.Config)) {
			i := g.Config["igor"]
			assert.NotNil(t, i)
			if assert.NotNil(t, i.Service) {
				assert.Equal(t, "nc5", i.Service.ObjectMeta.Namespace)
			}
			assert.Equal(t, 2, len(i.Resources))
		}
	}
}
