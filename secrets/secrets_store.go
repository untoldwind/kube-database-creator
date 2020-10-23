package secrets

import "k8s.io/client-go/rest"

type SecretsStore interface {
	Store(requestKey string, databaseName string, adminUsername string, adminPassword string) error
}

func SecretsStores(restConfig *rest.Config) (map[string]SecretsStore, error) {
	stores := map[string]SecretsStore{}

	kubernetes, err := newKubernetesSecretsStore(restConfig)
	if err != nil {
		return nil, err
	}
	stores["kubernetes"] = kubernetes

	return stores, nil
}
