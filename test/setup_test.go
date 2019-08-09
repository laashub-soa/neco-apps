package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	argocd "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	appSyncOrderFile   = "../app-sync-order.txt"
	alertmanagerSecret = `
route:
  receiver: slack
  group_wait: 5s # Send a notification after 5 seconds
  routes:
  - receiver: slack
    continue: true # Continue notification to next receiver

# Receiver configurations
receivers:
- name: slack
  slack_configs:
  - channel: '#test'
    api_url: https://hooks.slack.com/services/XXX/XXX
    icon_url: https://avatars3.githubusercontent.com/u/3380462 # Prometheus icon
    http_config:
      proxy_url: http://squid.internet-egress.svc.cluster.local:3128
`

	grafanaSecret = `apiVersion: v1
kind: Secret
metadata:
  labels:
    app.kubernetes.io/name: grafana
  name: grafana
  namespace: monitoring
type: Opaque
data:
  admin-password: QVVKVWwxSzJ4Z2Vxd01kWjNYbEVGYzFRaGdFUUl0T0RNTnpKd1FtZQ==
  admin-user: YWRtaW4=
  ldap-toml: ""
`

	teleportSecret = `
apiVersion: v1
kind: Secret
metadata:
  name: teleport-auth-secret
  namespace: teleport
  labels:
    app.kubernetes.io/name: teleport
stringData:
  teleport.yaml: |
    auth_service:
      authentication:
        second_factor: "off"
        type: local
      cluster_name: gcp0
      public_addr: teleport-auth:3025
      tokens:
        - "proxy,node:{{ .Token }}"
    teleport:
      data_dir: /var/lib/teleport
      auth_token: {{ .Token }}
      log:
        output: stderr
        severity: DEBUG
      storage:
        type: etcd
        peers: ["https://cke-etcd.kube-system.svc:2379"]
        tls_cert_file: /var/lib/etcd-certs/tls.crt
        tls_key_file: /var/lib/etcd-certs/tls.key
        tls_ca_file: /var/lib/etcd-certs/ca.crt
        prefix: /teleport
        insecure: false
---
apiVersion: v1
kind: Secret
metadata:
  name: teleport-proxy-secret
  namespace: teleport
  labels:
    app.kubernetes.io/name: teleport
stringData:
  teleport.yaml: |
    proxy_service:
      https_cert_file: /var/lib/certs/tls.crt
      https_key_file: /var/lib/certs/tls.key
      kubernetes:
        enabled: true
        listen_addr: 0.0.0.0:3026
      listen_addr: 0.0.0.0:3023
      public_addr: [ "teleport.gcp0.dev-ne.co:443" ]
      web_listen_addr: 0.0.0.0:3080
    teleport:
      data_dir: /var/lib/teleport
      auth_token: {{ .Token }}
      auth_servers:
        - teleport-auth:3025
      log:
        output: stderr
        severity: DEBUG
`
	elasticSecret = `
apiVersion: v1
kind: Secret
metadata:
  name: webhook-server-secret
  namespace: elastic-system
`
)

// testSetup tests setup of Argo CD
func testSetup() {
	It("should disable CKE-sabakan integration feature", func() {
		ExecSafeAt(boot0, "ckecli", "sabakan", "disable")
	})

	It("should list all apps in app-sync-order.txt", func() {
		appList := loadSyncOrder()
		kustomFile, err := filepath.Abs("../argocd-config/base/kustomization.yaml")
		Expect(err).ShouldNot(HaveOccurred())
		stdout, err := ioutil.ReadFile(kustomFile)
		Expect(err).ShouldNot(HaveOccurred())
		k := struct {
			Resources []string `yaml:"resources"`
		}{}
		Expect(yaml.Unmarshal(stdout, &k)).ShouldNot(HaveOccurred())
		var resources []string
		for _, r := range k.Resources {
			r = r[:len(r)-len(filepath.Ext(r))]
			resources = append(resources, r)
		}
		sort.Strings(appList)
		sort.Strings(resources)
		Expect(appList).Should(Equal(resources))
	})

	It("should re-issue kubeconfig", func() {
		ExecSafeAt(boot0, "ckecli", "kubernetes", "issue", ">", ".kube/config")
	})

	if !doUpgrade {
		It("should prepare secrets", func() {
			By("loading account.json")
			data, err := ioutil.ReadFile("account.json")
			Expect(err).ShouldNot(HaveOccurred())

			By("creating namespace and secrets for external-dns")
			ExecSafeAt(boot0, "kubectl", "create", "namespace", "external-dns")
			_, stderr, err := ExecAtWithInput(boot0, data, "kubectl", "--namespace=external-dns",
				"create", "secret", "generic", "external-dns", "--from-file=account.json=/dev/stdin")
			Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)

			By("creating namespace and secrets for alertmanager")
			stdout, stderr, err := ExecAtWithInput(boot0, []byte(alertmanagerSecret), "dd", "of=alertmanager.yaml")
			Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
			ExecSafeAt(boot0, "kubectl", "create", "namespace", "monitoring")
			ExecSafeAt(boot0, "kubectl", "--namespace=monitoring", "create", "secret",
				"generic", "alertmanager", "--from-file", "alertmanager.yaml")

			By("creating namespace and secrets for grafana")
			stdout, stderr, err = ExecAtWithInput(boot0, []byte(grafanaSecret), "dd", "of=grafana.yaml")
			Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
			ExecSafeAt(boot0, "kubectl", "apply", "-f", "grafana.yaml")

			By("creating namespace and secrets for teleport")
			stdout, stderr, err = ExecAt(boot0, "env", "ETCDCTL_API=3", "etcdctl", "--cert=/etc/etcd/backup.crt", "--key=/etc/etcd/backup.key",
				"get", "--print-value-only", "/neco/teleport/auth-token")
			Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
			teleportToken := strings.TrimSpace(string(stdout))
			teleportTmpl := template.Must(template.New("").Parse(teleportSecret))
			buf := bytes.NewBuffer(nil)
			err = teleportTmpl.Execute(buf, struct {
				Token string
			}{
				Token: teleportToken,
			})
			Expect(err).NotTo(HaveOccurred())
			ExecSafeAt(boot0, "kubectl", "create", "namespace", "teleport")
			stdout, stderr, err = ExecAtWithInput(boot0, buf.Bytes(), "kubectl", "apply", "-n", "teleport", "-f", "-")
			Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

			ExecSafeAt(boot0, "ckecli", "etcd", "user-add", "teleport", "/teleport")
			ExecSafeAt(boot0, "ckecli", "etcd", "issue", "teleport", "--output", "file")
			ExecSafeAt(boot0, "kubectl", "-n", "teleport", "create", "secret", "generic",
				"teleport-etcd-certs", "--from-file=ca.crt=etcd-ca.crt",
				"--from-file=tls.crt=etcd-teleport.crt", "--from-file=tls.key=etcd-teleport.key")

			By("creating namespace and secrets for elastic")
			ExecSafeAt(boot0, "kubectl", "create", "namespace", "elastic-system")
			stdout, stderr, err = ExecAtWithInput(boot0, []byte(elasticSecret), "kubectl", "--namespace=elastic-system", "create", "-f", "-")
			Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
		})
	} else {
		// Workaround for teleport initial release
		// TODO: Remove this else block when teleport is merged into the release branch.
		It("should prepare secrets for teleport", func() {
			By("checking teleport namespace")
			stdout := ExecSafeAt(boot0, "kubectl", "get", "namespaces", "-o", "json")
			var nsList corev1.NamespaceList
			err := json.Unmarshal(stdout, &nsList)
			Expect(err).ShouldNot(HaveOccurred(), "stdout=%s", string(stdout))
			setupTeleport := true
			for _, v := range nsList.Items {
				if v.Name == "teleport" {
					setupTeleport = false
					break
				}
			}

			if setupTeleport {
				By("creating namespace and secrets for teleport")
				stdout, stderr, err := ExecAt(boot0, "env", "ETCDCTL_API=3", "etcdctl", "--cert=/etc/etcd/backup.crt", "--key=/etc/etcd/backup.key",
					"get", "--print-value-only", "/neco/teleport/auth-token")
				Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
				teleportToken := strings.TrimSpace(string(stdout))
				teleportTmpl := template.Must(template.New("").Parse(teleportSecret))
				buf := bytes.NewBuffer(nil)
				err = teleportTmpl.Execute(buf, struct {
					Token string
				}{
					Token: teleportToken,
				})
				Expect(err).NotTo(HaveOccurred())
				ExecSafeAt(boot0, "kubectl", "create", "namespace", "teleport")
				stdout, stderr, err = ExecAtWithInput(boot0, buf.Bytes(), "kubectl", "apply", "-n", "teleport", "-f", "-")
				Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)

				ExecSafeAt(boot0, "ckecli", "etcd", "user-add", "teleport", "/teleport")
				ExecSafeAt(boot0, "ckecli", "etcd", "issue", "teleport", "--output", "file")
				ExecSafeAt(boot0, "kubectl", "-n", "teleport", "create", "secret", "generic",
					"teleport-etcd-certs", "--from-file=ca.crt=etcd-ca.crt",
					"--from-file=tls.crt=etcd-teleport.crt", "--from-file=tls.key=etcd-teleport.key")
			}
		})

		// Workaround for teleport-auth upgrade
		// TODO: Remove this section when teleport-auth using pv is merged into the release branch.
		It("should upgrade statefulset (teleport-auth)", func() {
			By("checking teleport-auth spec")
			stdout := ExecSafeAt(boot0, "kubectl", "-n", "teleport", "get", "statefulset", "-o", "json")
			var stsList appsv1.StatefulSetList
			err := json.Unmarshal(stdout, &stsList)
			Expect(err).ShouldNot(HaveOccurred(), "stdout=%s", string(stdout))
			for _, sts := range stsList.Items {
				if sts.Name != "teleport-auth" {
					continue
				}
				if len(sts.Spec.VolumeClaimTemplates) == 0 {
					By("deleting previous teleport-auth")
					ExecSafeAt(boot0, "kubectl", "-n", "teleport", "delete", "statefulset", "teleport-auth")
				}
			}
		})

		// TODO: Remove this else block when elastic is merged into the release branch.
		It("should prepare secrets for elastic", func() {
			By("checking elastic namespace")
			stdout := ExecSafeAt(boot0, "kubectl", "get", "namespaces", "-o", "json")
			var nsList corev1.NamespaceList
			err := json.Unmarshal(stdout, &nsList)
			Expect(err).ShouldNot(HaveOccurred(), "stdout=%s", string(stdout))
			setupElastic := true
			for _, v := range nsList.Items {
				if v.Name == "elastic-system" {
					setupElastic = false
					break
				}
			}

			if setupElastic {
				By("creating namespace and secrets for elastic")
				ExecSafeAt(boot0, "kubectl", "create", "namespace", "elastic-system")
				stdout, stderr, err := ExecAtWithInput(boot0, []byte(elasticSecret), "kubectl", "--namespace=elastic-system", "create", "-f", "-")
				Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
			}
		})

		// TODO: Remove this else block when grafana is merged into the release branch.
		It("should prepare secrets for grafana", func() {
			By("checking elastic namespace")
			stdout := ExecSafeAt(boot0, "kubectl", "get", "deployment", "-n", "monitoring", "-o", "json")
			var deployList appsv1.DeploymentList
			err := json.Unmarshal(stdout, &deployList)
			Expect(err).ShouldNot(HaveOccurred(), "stdout=%s", string(stdout))
			setupGrafana := true
			for _, d := range deployList.Items {
				if d.Name == "grafana" {
					setupGrafana = false
					break
				}
			}

			if setupGrafana {
				By("creating secrets for grafana")
				stdout, stderr, err := ExecAtWithInput(boot0, []byte(grafanaSecret), "kubectl", "--namespace=monitoring", "create", "-f", "-")
				Expect(err).NotTo(HaveOccurred(), "stdout: %s, stderr: %s", stdout, stderr)
			}
		})
	}

	It("should checkout neco-apps repository@"+commitID, func() {
		ExecSafeAt(boot0, "rm", "-rf", "neco-apps")
		ExecSafeAt(boot0, "env", "https_proxy=http://10.0.49.3:3128",
			"git", "clone", "https://github.com/cybozu-go/neco-apps")
		ExecSafeAt(boot0, "cd neco-apps; git checkout "+commitID)
	})

	It("should setup applications", func() {
		if !doUpgrade {
			setupArgoCD()
		}
		ExecSafeAt(boot0, "sed", "-i", "s/release/"+commitID+"/", "./neco-apps/argocd-config/base/*.yaml")
		applyAndWaitForApplications()
	})
}

func loadSyncOrder() []string {
	orderFileAbs, err := filepath.Abs(appSyncOrderFile)
	Expect(err).ShouldNot(HaveOccurred())
	stdout, err := ioutil.ReadFile(orderFileAbs)
	Expect(err).ShouldNot(HaveOccurred())
	lines := strings.Split(string(stdout), "\n")
	var results []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		results = append(results, line)
	}

	return results
}

func applyAndWaitForApplications() {
	By("creating Argo CD app")
	ExecSafeAt(boot0, "kubectl", "apply", "-k", "./neco-apps/argocd-config/overlays/gcp")

	syncOrder := loadSyncOrder()

	By("waiting initialization")
	Eventually(func() error {
	OUTER:
		for _, appName := range syncOrder {
			stdout, stderr, err := ExecAt(boot0, "argocd", "app", "get", "-o", "json", appName)
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}
			var app argocd.Application
			err = json.Unmarshal(stdout, &app)
			if err != nil {
				return err
			}
			if app.Status.Sync.ComparedTo.Source.TargetRevision != commitID {
				return errors.New(appName + " does not have correct target yet")
			}
			st := app.Status
			if st.Sync.Status == argocd.SyncStatusCodeSynced &&
				(st.Health.Status == argocd.HealthStatusHealthy || st.Health.Status == argocd.HealthStatusProgressing) &&
				app.Operation == nil {
				continue
			}
			for _, cond := range st.Conditions {
				if cond.Type == argocd.ApplicationConditionSyncError {
					continue OUTER
				}
			}
			return errors.New(appName + " is not initialized")
		}
		return nil
	}).Should(Succeed())

	for _, appName := range syncOrder {
		By("syncing " + appName + " manually")
		ExecSafeAt(boot0, "argocd", "app", "sync", appName)
	}
}

func setupArgoCD() {
	By("installing Argo CD")
	data, err := ioutil.ReadFile("install.yaml")
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() error {
		stdout, stderr, err := ExecAtWithInput(boot0, data, "kubectl", "apply", "-n", "argocd", "-f", "-")
		if err != nil {
			return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}
		return nil
	}).Should(Succeed())

	By("logging in to Argo CD")
	// admin password is same as pod name
	var podList corev1.PodList
	Eventually(func() error {
		stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "pods", "-n", "argocd",
			"-l", "app.kubernetes.io/name=argocd-server", "-o", "json")
		if err != nil {
			return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}
		err = json.Unmarshal(stdout, &podList)
		if err != nil {
			return err
		}
		if podList.Items == nil {
			return errors.New("podList.Items is nil")
		}
		if len(podList.Items) != 1 {
			return fmt.Errorf("podList.Items is not 1: %d", len(podList.Items))
		}
		return nil
	}).Should(Succeed())

	password := podList.Items[0].Name

	By("getting node address")
	var nodeList corev1.NodeList
	data = ExecSafeAt(boot0, "kubectl", "get", "nodes", "-o", "json")
	err = json.Unmarshal(data, &nodeList)
	Expect(err).ShouldNot(HaveOccurred(), "data=%s", string(data))
	Expect(nodeList.Items).ShouldNot(BeEmpty())
	node := nodeList.Items[0]

	var nodeAddress string
	for _, addr := range node.Status.Addresses {
		if addr.Type != corev1.NodeInternalIP {
			continue
		}
		nodeAddress = addr.Address
	}
	Expect(nodeAddress).ShouldNot(BeNil())

	By("getting node port")
	var svc corev1.Service
	data = ExecSafeAt(boot0, "kubectl", "get", "svc/argocd-server", "-n", "argocd", "-o", "json")
	err = json.Unmarshal(data, &svc)
	Expect(err).ShouldNot(HaveOccurred(), "data=%s", string(data))
	Expect(svc.Spec.Ports).ShouldNot(BeEmpty())

	var nodePort string
	for _, port := range svc.Spec.Ports {
		if port.Name != "http" {
			continue
		}
		nodePort = strconv.Itoa(int(port.NodePort))
	}
	Expect(nodePort).ShouldNot(BeNil())

	By("logging in to Argo CD")
	Eventually(func() error {
		stdout, stderr, err := ExecAt(boot0, "argocd", "login", nodeAddress+":"+nodePort,
			"--insecure", "--username", "admin", "--password", password)
		if err != nil {
			return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}
		return nil
	}).Should(Succeed())
}
