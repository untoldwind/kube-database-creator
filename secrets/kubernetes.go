package secrets

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	v1 "k8s.io/client-go/deprecated/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type kubernetesSecretsStore struct {
	client *v1.CoreV1Client
}

func newKubernetesSecretsStore(config *rest.Config) (SecretsStore, error) {
	client, err := v1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kubernetesSecretsStore{
		client: client,
	}, nil
}

func (s *kubernetesSecretsStore) Store(requestKey string, databaseName string, adminUsername string, adminPassword string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(requestKey)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", requestKey))
		return nil
	}

	secrets := s.client.Secrets(namespace)
	_, err = secrets.Get(name, metav1.GetOptions{})

	if err == nil {
		klog.Infof("Secret %s alread exists", requestKey)
		return nil
	} else if !errors.IsNotFound(err) {
		return err
	}

	klog.Infof("Creating secret %s", requestKey)

	if _, err := secrets.Create(&apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"kube-database-creator": "admin-user",
			},
		},
		Type: apiv1.SecretTypeOpaque,
		StringData: map[string]string{
			"database":                 databaseName,
			apiv1.BasicAuthUsernameKey: adminUsername,
			apiv1.BasicAuthPasswordKey: adminPassword,
		},
	}); err != nil {
		return err
	}

	return nil
}
