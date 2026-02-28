package postgresdb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	v1alpha1 "github.com/apranto/pgoperator/pkg/apis/postgresdb/v1alpha1"
)

var postgresDBResource = schema.GroupVersionResource{
	Group:    v1alpha1.GroupName,
	Version:  v1alpha1.GroupVersion,
	Resource: "postgresdbs",
}

type Controller struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
	informer      cache.SharedIndexInformer
	queue         workqueue.TypedRateLimitingInterface[string]
	reconciler    *Reconciler
}

func NewController(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *Controller {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 30*time.Second)
	informer := factory.ForResource(postgresDBResource).Informer()

	queue := workqueue.NewTypedRateLimitingQueue(
		workqueue.DefaultTypedControllerRateLimiter[string](),
	)

	c := &Controller{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		informer:      informer,
		queue:         queue,
		reconciler:    NewReconciler(kubeClient),
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				klog.Errorf("Postgres DB update failed: %s", err)
			}
			c.queue.Add(key)
			klog.Infof("Postgres DB added: %s", key)
		},

		UpdateFunc: func(obj interface{}, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				klog.Errorf("Postgres DB update failed: %s", err)
			}
			klog.Infof("Postgres DB updated: %s", key)
			c.queue.Add(key)
		},

		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				klog.Errorf("Postgres DB update failed: %s", err)
			}
			klog.Infof("Postgres DB deleted: %s", key)
			c.queue.Add(key)
		},
	})

	return c
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("Starting PostgresDB controller")

	go c.informer.Run(stopCh)

	klog.Info("Waiting for informer cache to sync...")
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return fmt.Errorf("timed out waiting for cache sync")
	}
	klog.Info("Cache synced, starting workers")

	for i := 0; i < 2; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Controller is running")
	<-stopCh
	klog.Info("Shutting down controller")
	return nil
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// do nothing for now
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	defer c.queue.Done(key)
	if quit {
		return false
	}
	err := c.reconcile(key)
	if err != nil {
		klog.Errorf("Error reconciling %s: %v", key, err)
		c.queue.AddRateLimited(key)
		return true
	} else {
		c.queue.Forget(key)
		return true
	}
}

func (c *Controller) reconcile(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("Error occured while splitting metadata namespace key: %v", err)
	}
	obj, exists, err := c.informer.GetStore().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error occured while geting key from informer: %v", err)
	}
	if !exists {
		klog.Infof("PostgresDB %s/%s was deleted, skipping", namespace, name)
		return nil
	}
	unstructuredObject := obj.(*unstructured.Unstructured)
	pgdb := &v1alpha1.PostgresDB{}
	err = fromUnstructured(unstructuredObject, pgdb)
	if err != nil {
		return fmt.Errorf("Could not parse pgdb object from unstructured object: %v", err)
	}
	klog.Infof("Reconciling: %s", key)

	if err := c.reconciler.Reconcile(context.TODO(), pgdb); err != nil {
		return fmt.Errorf("reconciliation failed for %s: %w", key, err)
	}
	return nil
}

func fromUnstructured(obj *unstructured.Unstructured, target interface{}) error {
	data, err := json.Marshal(obj.Object)
	if err != nil {
		return fmt.Errorf("failed to marshal unstructured: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target: %w", err)
	}
	return nil
}
