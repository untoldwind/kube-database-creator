package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type Controller struct {
	creators        map[string]*Creator
	clientset       kubernetes.Interface
	configMapLister corelisters.ConfigMapLister
	configMapSynced cache.InformerSynced
	workqueue       workqueue.RateLimitingInterface
}

func NewController(creators map[string]*Creator, clientset kubernetes.Interface, configMapInformer coreinformer.ConfigMapInformer) *Controller {
	controller := &Controller{
		creators:        creators,
		clientset:       clientset,
		configMapLister: configMapInformer.Lister(),
		configMapSynced: configMapInformer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Database Create Requests"),
	}

	klog.Info("Setting up event handlers")

	configMapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFoo,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueFoo(new)
		},
	})

	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	utilruntime.HandleCrash()

	klog.Info("Starting kube-database-creator controller")

	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.configMapSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)

		if key, ok := obj.(string); ok {
			if err := c.syncHandler(key); err != nil {
				c.workqueue.AddRateLimited(key)
				return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
			}
		} else {
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		}

		c.workqueue.Forget(obj)

		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	configMap, err := c.configMapLister.ConfigMaps(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("configmap '%s' in work queue no longer exists", key))
			return nil
		}
	}

	serverName, ok := configMap.Data["server"]
	if !ok {
		utilruntime.HandleError(fmt.Errorf("configmap '%s' does not have a 'server' name", key))
		return nil
	}
	databaseName, ok := configMap.Data["database"]
	if !ok {
		utilruntime.HandleError(fmt.Errorf("configmap '%s' does not have a 'database' name", key))
		return nil
	}
	creator, ok := c.creators[serverName]
	if !ok {
		utilruntime.HandleError(fmt.Errorf("in configmap '%s' no configuration found for server=%s", key, serverName))
		return nil
	}

	return creator.HandleRequest(databaseName)
}

func (c *Controller) enqueueFoo(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
