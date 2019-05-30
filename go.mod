module github.com/cybozu-go/neco-ops

go 1.12

replace launchpad.net/gocheck => github.com/go-check/check v0.0.0-20180628173108-788fd7840127

require (
	github.com/apache/thrift v0.12.0 // indirect
	github.com/argoproj/argo-cd v0.12.1
	github.com/codegangsta/cli v1.20.0 // indirect
	github.com/cybozu-go/log v1.5.0
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jetstack/cert-manager v0.8.0
	github.com/kubernetes-incubator/external-dns v0.5.14
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/olekukonko/tablewriter v0.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openzipkin/zipkin-go v0.1.6 // indirect
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/common v0.3.0
	golang.org/x/crypto v0.0.0-20190426145343-a29dc8fdc734
	gopkg.in/src-d/go-git.v4 v4.11.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190413052509-3cc1b3fb6d0f
	k8s.io/client-go v11.0.0+incompatible // indirect
	k8s.io/klog v0.3.0 // indirect
	k8s.io/utils v0.0.0-20190308190857-21c4ce38f2a7 // indirect
)
