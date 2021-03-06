package cachers

import (
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/chongyangshi/Order/logging"
)

// deploymentsCacheController holds an eventually consistent cache of deployments
// to allow Order to determine what DaemonSet pods need to be rolling restarted
// quickly.
type deploymentsCacheController struct {
	factory informers.SharedInformerFactory
	lister  appslisters.DeploymentLister
	synced  cache.InformerSynced
}

// newDeploymentsController initialises a Deployments controller
func newDeploymentsController(clientSet kubernetes.Interface, resyncInterval time.Duration) *deploymentsCacheController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Apps().V1().Deployments()

	controller := &deploymentsCacheController{
		factory: informerFactory,
	}

	controller.lister = informer.Lister()
	controller.synced = informer.Informer().HasSynced

	return controller
}

// run initialises and starts the controller
func (c *deploymentsCacheController) run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	logging.Log("Starting deployment cache controller.")
	defer logging.Log("Shutting down deployment cache controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.synced); !ok {
		logging.Fatal("Failed to wait for cache synchronization")
	}

	<-stopChan
}
