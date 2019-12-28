// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2019 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cluster

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	vapi "github.com/hashicorp/vault/api"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type rolePolicy struct {
	roleID       string
	roleSecretID string
}

type policyResult struct {
	Policy rolePolicy
	Error  error
}

type kmsCnxResult struct {
	Cnx   *vapi.Client
	Error error
}

func connectToKms() chan kmsCnxResult {
	ch := make(chan kmsCnxResult)
	go func() {
		defer close(ch)
		kmsAddress := getKmsAddress()
		if kmsAddress == "" {
			ch <- kmsCnxResult{Error: errors.New("missing kms address")}
			return
		}
		kmsToken := getKmsToken()
		if kmsToken == "" {
			ch <- kmsCnxResult{Error: errors.New("missing kms token")}
			return
		}

		client, err := vapi.NewClient(&vapi.Config{Address: kmsAddress})
		if err != nil {
			ch <- kmsCnxResult{Error: err}
			return
		}

		client.SetToken(kmsToken)

		health, err := client.Sys().Health()

		if err != nil {
			ch <- kmsCnxResult{Error: err}
			return
		}

		if !health.Initialized {
			ch <- kmsCnxResult{Error: errors.New("kms is not initialized")}
			return
		}

		if health.Sealed {
			ch <- kmsCnxResult{Error: errors.New("kms is sealed")}
			return
		}

		ch <- kmsCnxResult{Cnx: client}
	}()
	return ch
}

func creatNewPolicyOnExternalKMS(KmsClient *vapi.Client, tenant string) <-chan policyResult {
	doneCh := make(chan policyResult)
	go func() {
		defer close(doneCh)
		kms := KmsClient
		policyName := fmt.Sprintf("%s-kes-policy", tenant)

		existingPolicy, _ := kms.Sys().GetPolicy(policyName)

		if existingPolicy != "" {
			log.Println("Error creating policy on external kms because the policy already exists ", policyName, existingPolicy)
			doneCh <- policyResult{Error: errors.New("a policy with that name already exists on the kms")}
			return
		}

		policyRules := fmt.Sprintf(`
		path "kv/%s/*" {
				capabilities = [ "create", "read", "delete" ]
		}
		`, tenant)

		err := kms.Sys().PutPolicy(policyName, policyRules)
		if err != nil {
			log.Println("Error creating policy on external kms: ")
			doneCh <- policyResult{Error: err}
			return
		}
		data := map[string]interface{}{
			"policy":             policyName,
			"token_num_uses":     0,
			"secret_id_num_uses": 0,
			"period":             "5m",
		}
		roleName := fmt.Sprintf("%s-kes-role", tenant)
		_, err = kms.Logical().Write(fmt.Sprintf("auth/approle/role/%s", roleName), data)
		if err != nil {
			log.Println("Error creating new role on external kms: ", tenant, data, err)
			doneCh <- policyResult{Error: err}
			return
		}
		role, err := kms.Logical().Read(fmt.Sprintf("auth/approle/role/%s/role-id", roleName))
		if err != nil {
			log.Println("Error reading role_id from external kms: ", tenant, err)
			doneCh <- policyResult{Error: err}
			return
		}
		roleSecret, err := kms.Logical().Write(fmt.Sprintf("auth/approle/role/%s/secret-id", roleName), map[string]interface{}{})
		if err != nil {
			log.Println("Error reading role_secret_id from external kms: ", tenant)
			doneCh <- policyResult{Error: err}
			return
		}

		roleID := role.Data["role_id"].(string)
		roleSecretID := roleSecret.Data["secret_id"].(string)
		doneCh <- policyResult{Policy: rolePolicy{roleID: roleID, roleSecretID: roleSecretID}}
	}()
	return doneCh
}

func getNewKesDeployment(deploymentName string, kesSecretsNames map[string]string) appsv1.Deployment {
	kesReplicas := int32(1)
	kesPodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            deploymentName,
				Image:           getKesContainerImage(),
				ImagePullPolicy: "IfNotPresent",
				Command:         []string{"kes"},
				Args:            []string{"server", "--config=kes-config/server-config.toml", "--mtls-auth=ignore"},
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 7373,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "configuration",
						MountPath: "/kes-config",
						ReadOnly:  true,
					},
					{
						Name:      "server-keypair-key",
						MountPath: "/kes-config/server/key",
						ReadOnly:  true,
					},
					{
						Name:      "server-keypair-cert",
						MountPath: "/kes-config/server/cert",
						ReadOnly:  true,
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "configuration",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: kesSecretsNames["kesServerConfigSecretName"],
					},
				},
			},
			{
				Name: "server-keypair-key",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: kesSecretsNames["kesServerKeyPairKeySecretName"],
					},
				},
			},
			{
				Name: "server-keypair-cert",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: kesSecretsNames["kesServerKeyPairCertSecretName"],
					},
				},
			},
		},
	}
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Labels: map[string]string{
				"app":  deploymentName,
				"type": "kes",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &kesReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  deploymentName,
					"type": "kes",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  deploymentName,
						"type": "kes",
					},
				},
				Spec: kesPodSpec,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}
}

func createNewKesDeployment(clientset *kubernetes.Clientset, deploymentName string, kesSecretsNames map[string]string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		factory := informers.NewSharedInformerFactory(clientset, 0)
		deploymentInformer := factory.Apps().V1().Deployments().Informer()
		deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				deployment := newObj.(*appsv1.Deployment)
				if deployment.Name == deploymentName && len(deployment.Status.Conditions) > 0 && deployment.Status.Conditions[0].Status == "True" {
					log.Println("kes deployment created correctly")
					return
				}
			},
		})
		go deploymentInformer.Run(doneCh)

		//Creating nginx-resolver deployment with new rules
		kesDeployment := getNewKesDeployment(deploymentName, kesSecretsNames)
		_, err := appsV1API(clientset).Deployments("default").Create(&kesDeployment)
		if err != nil {
			log.Println(err)
			return
		}
	}()
	return doneCh
}

func createNewKesService(clientset *kubernetes.Clientset, serviceName string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		factory := informers.NewSharedInformerFactory(clientset, 0)
		serviceInformer := factory.Core().V1().Services().Informer()
		serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				service := obj.(*v1.Service)
				if service.Name == serviceName {
					log.Println("kes service created correctly")
					return
				}
			},
		})

		go serviceInformer.Run(doneCh)

		pgSvc := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName,
				Labels: map[string]string{
					"name": serviceName,
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Port: 7373,
					},
				},
				Selector: map[string]string{
					"app": serviceName,
				},
			},
		}
		_, err := clientset.CoreV1().Services("default").Create(&pgSvc)
		if err != nil {
			log.Println(err)
			return
		}
	}()
	return doneCh
}

type KeyPair struct {
	cert         string
	certIdentity string
	key          string
}

func generateKeyPair(name string) chan *KeyPair {
	doneCh := make(chan *KeyPair)
	go func() {
		defer close(doneCh)
		public, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			log.Println(err)
			return
		}
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			log.Println(err)
			return
		}
		now := time.Now()
		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				CommonName: name,
			},
			NotBefore:             now,
			NotAfter:              now.Add(87660 * time.Hour),
			KeyUsage:              x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			BasicConstraintsValid: true,
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, public, private)
		if err != nil {
			log.Println(err)
			return
		}
		privBytes, err := x509.MarshalPKCS8PrivateKey(private)
		if err != nil {
			log.Println(err)
			return
		}
		key := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}))
		cert := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))

		h := crypto.SHA256.New()
		publicCertificate, _ := x509.ParseCertificate(derBytes)
		h.Write(publicCertificate.RawSubjectPublicKeyInfo)

		doneCh <- &KeyPair{
			key:          key,
			cert:         cert,
			certIdentity: hex.EncodeToString(h.Sum(nil)),
		}
	}()
	return doneCh
}

func storeKeyPairInSecret(secretName string, content map[string]string) <-chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		clientSet, err := k8sClient()
		if err != nil {
			log.Println(err)
			return
		}
		factory := informers.NewSharedInformerFactory(clientSet, 0)
		secretInformer := factory.Core().V1().Secrets().Informer()
		secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				secret := obj.(*corev1.Secret)
				if secret.Name == secretName {
					log.Println(secret)
					return
				}
			},
		})

		go secretInformer.Run(doneCh)

		_, err = clientSet.CoreV1().Secrets("default").Create(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
				Labels: map[string]string{
					"name": secretName,
				},
			},
			StringData: content,
		})
		if err != nil {
			log.Println(err)
			return
		}
	}()
	return doneCh
}

func generateKeyPairAndStoreInSecret(name string) *KeyPair {
	kesKeyPair := <-generateKeyPair(name)
	if kesKeyPair != nil && kesKeyPair.cert != "" && kesKeyPair.key != "" {
		<-storeKeyPairInSecret(fmt.Sprintf("%s-cert", name), map[string]string{
			"cert":         kesKeyPair.cert,
			"certIdentity": kesKeyPair.certIdentity,
		})
		<-storeKeyPairInSecret(fmt.Sprintf("%s-key", name), map[string]string{
			"key": kesKeyPair.key,
		})
	}
	return kesKeyPair
}

func createKesConfigurations(KmsClient *vapi.Client, tenant string, roleID string, roleSecretID string) map[string]string {
	kms := KmsClient
	kesAppKeyPairSecretName := fmt.Sprintf("%s-kes-app-keypair", tenant)
	kesServerKeyPairSecretName := fmt.Sprintf("%s-kes-server-keypair", tenant)
	appKeys := generateKeyPairAndStoreInSecret(kesAppKeyPairSecretName)
	generateKeyPairAndStoreInSecret(kesServerKeyPairSecretName)

	kesServerConfig := fmt.Sprintf(`
			address = "127.0.0.1:7373"
			root = "disabled"
			
			[tls]
			key  = "kes-config/server/key/key"
			cert = "kes-config/server/cert/cert"
			
			[policy.prod-app]
			paths      = [ "/v1/key/create/app-key", "/v1/key/generate/app-key" , "/v1/key/decrypt/app-key"]
			identities = [ "%s" ]
			
			[keystore.vault]
			address = "%s"
			[keystore.vault.approle]
			id     = "%s"
			secret = "%s"
			retry  = "15s"
			[keystore.vault.status]
			ping = "10s"
		`, appKeys.certIdentity, kms.Address(), roleID, roleSecretID)

	kesServerConfigSecretName := fmt.Sprintf("%s-kes-server-config", tenant)
	<-storeKeyPairInSecret(kesServerConfigSecretName, map[string]string{
		"server-config.toml": kesServerConfig,
	})
	return map[string]string{
		"kesServerConfigSecretName":      kesServerConfigSecretName,
		"kesServerKeyPairKeySecretName":  fmt.Sprintf("%s-key", kesServerKeyPairSecretName),
		"kesServerKeyPairCertSecretName": fmt.Sprintf("%s-cert", kesServerKeyPairSecretName),
		"kesAppKeyPairKeySecretName":     fmt.Sprintf("%s-key", kesAppKeyPairSecretName),
		"kesAppKeyPairCertSecretName":    fmt.Sprintf("%s-cert", kesAppKeyPairSecretName),
	}
}

func StartNewKes(shortName string) chan error {
	doneCh := make(chan error)
	go func() {
		defer close(doneCh)
		clientset, err := k8sClient()
		if err != nil {
			doneCh <- err
			return
		}

		kmsResult := <-connectToKms()
		if kmsResult.Error != nil {
			log.Println(kmsResult.Error)
			doneCh <- kmsResult.Error
			return
		}

		policy := <-creatNewPolicyOnExternalKMS(kmsResult.Cnx, shortName)
		if policy.Error != nil {
			log.Println(policy.Error)
			doneCh <- policy.Error
			return
		}
		kesSecretsNames := createKesConfigurations(kmsResult.Cnx, shortName, policy.Policy.roleID, policy.Policy.roleSecretID)

		tenantKesName := fmt.Sprintf("%s-kes", shortName)
		<-createNewKesDeployment(clientset, tenantKesName, kesSecretsNames)
		<-createNewKesService(clientset, tenantKesName)
	}()
	return doneCh
}
