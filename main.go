package main

import (
	"flag"
	"time"

	"github.com/untoldwind/kube-database-creator/secrets"
	"github.com/untoldwind/kube-database-creator/signals"
	apiV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	configFile string
	masterURL  string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	config, err := LoadConfig(configFile)
	if err != nil {
		klog.Fatalf("Error reading config: %s", err.Error())
	}

	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	kubeInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, time.Second*30, informers.WithTweakListOptions(tweakListOptions))

	secretStores, err := secrets.SecretsStores(cfg)
	if err != nil {
		klog.Fatalf("Error building secrets stores backend clientset: %s", err.Error())
	}

	creators := map[string]*Creator{}

	for _, serverConfig := range config.Servers {
		creator, err := NewCreator(serverConfig, secretStores)
		if err != nil {
			klog.Fatalf("Error initializing creator: %s", err.Error())
		}
		creators[creator.Name] = creator
	}

	controller := NewController(creators, kubeClient, kubeInformerFactory.Core().V1().ConfigMaps())

	kubeInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&configFile, "config", "/run/database-creator/config.json", "Path to the database creator config")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func tweakListOptions(listOptions *apiV1.ListOptions) {
	listOptions.LabelSelector = "kube-database-creator = request"
}
