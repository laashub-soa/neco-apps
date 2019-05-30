package test

import (
	"encoding/json"
	"errors"
	"fmt"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

func testCertManager() {
	It("should be deployed successfully", func() {
		Eventually(func() error {
			stdout, _, err := ExecAt(boot0, "kubectl", "--namespace=external-dns",
				"get", "deployment", "--selector=component=cert-manager", "-o=json")
			if err != nil {
				return err
			}

			deploymentList := new(appsv1.DeploymentList)
			err = json.Unmarshal(stdout, deploymentList)
			if err != nil {
				return err
			}

			for _, deploy := range deploymentList.Items {
				if deploy.Status.AvailableReplicas != 1 {
					return fmt.Errorf("%s deployment's AvailableReplica is not 1: %d", deploy.Name, int(deploy.Status.AvailableReplicas))
				}
			}

			return nil
		}).Should(Succeed())
	})

	It("should issue certificate", func() {
		domainName := testID + "-cert-manager.gcp0.dev-ne.co"
		By("deploying Certificate")
		certificate := fmt.Sprintf(`
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: test-certificate
  namespace: external-dns
spec:
  secretName: example-com-tls
  issuerRef:
    name: clouddns
  commonName: gcp0.dev-ne.co
  dnsNames:
  - %s
`, domainName)
		_, stderr, err := ExecAtWithInput(boot0, []byte(certificate), "kubectl", "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred(), "stderr: %s", stderr)

		By("checking certificate is issued for xxx.gcp0.dev-ne.co")
		Eventually(func() error {
			stdout, stderr, err := ExecAt(boot0, "kubectl", "get", "-n=external-dns", "certificate", "-o", "json")
			if err != nil {
				return fmt.Errorf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
			}

			var cert certmanagerv1alpha1.Certificate
			err = json.Unmarshal(stdout, &cert)
			if err != nil {
				return err
			}

			if len(cert.Status.Conditions) == 0 {
				return errors.New("status not found")
			}

			status := cert.Status.Conditions[0]
			if status.Status != certmanagerv1alpha1.ConditionTrue {
				return fmt.Errorf("Certificate status is not True: %s", status.Status)
			}
			if status.Reason != "CertIssued" {
				return fmt.Errorf("Certificate reason not CertIssued: %s", status.Reason)
            }

            return nil
		}).Should(Succeed())
	})
}
