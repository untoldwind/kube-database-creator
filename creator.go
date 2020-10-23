package main

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/untoldwind/kube-database-creator/secrets"

	"github.com/untoldwind/kube-database-creator/platforms"
	"k8s.io/klog/v2"
)

type Creator struct {
	Name         string
	plaform      platforms.Platform
	secretsStore secrets.SecretsStore
}

func NewCreator(config ServerConfig, secretsStores map[string]secrets.SecretsStore) (*Creator, error) {
	plaform, err := platforms.NewPlatform(config.URL)
	if err != nil {
		return nil, err
	}
	secretsStoreName := config.SecretsStore
	if secretsStoreName == "" {
		secretsStoreName = "kubernetes"
	}
	secretsStore, ok := secretsStores[secretsStoreName]
	if !ok {
		return nil, fmt.Errorf("No secrets store backend for %s", secretsStoreName)
	}

	return &Creator{
		Name:         config.Name,
		plaform:      plaform,
		secretsStore: secretsStore,
	}, nil
}

func (c *Creator) HandleRequest(requestKey string, databaseName string) error {
	klog.Infof("Creator %s: Handle request for database %s", c.Name, databaseName)

	exists, err := c.plaform.CheckExists(databaseName)
	if err != nil {
		return err
	}
	if exists {
		klog.Infof("Creator %s: Database %s already exists (ignoring request)", c.Name, databaseName)

		return nil
	}

	adminUsername := fmt.Sprintf("%s_admin", databaseName)
	adminPassword, err := generatePassword()
	if err != nil {
		return err
	}

	if err := c.plaform.Create(databaseName, adminUsername, adminPassword); err != nil {
		return err
	}

	klog.Infof("Creator %s: Successfully created database %s", c.Name, databaseName)

	if err := c.secretsStore.Store(requestKey, databaseName, adminUsername, adminPassword); err != nil {
		return err
	}

	return nil
}

func generatePassword() (string, error) {
	length := 20
	random := make([]byte, length)

	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[int(random[i])%len(chars)])
	}
	return b.String(), nil
}
