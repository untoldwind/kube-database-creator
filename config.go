package main

import (
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
)

type Config struct {
	Servers []ServerConfig `json:"servers"`
}

type ServerConfig struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	SecretsStore string `json:"secrets_store,omitempty"`
}

func (c *Config) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (c *Config) DeepCopyObject() runtime.Object {
	return c
}

func LoadConfig(filename string) (*Config, error) {
	configbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config := &Config{}

	_, _, err = clientcmdlatest.Codec.Decode(configbytes, nil, config)

	if err != nil {
		return nil, err
	}

	return config, nil

}
