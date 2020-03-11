package bansslice

import (
	"context"
	"reflect"

	bansv1alpha1 "github.com/stevenchiu30801/bans5gc-operator/pkg/apis/bans/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var reqLogger = logf.Log.WithName("controller_bansslice")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new BansSlice Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBansSlice{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("bansslice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BansSlice
	err = c.Watch(&source.Kind{Type: &bansv1alpha1.BansSlice{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner BansSlice
	err = c.Watch(&source.Kind{Type: &bansv1alpha1.Free5GCSlice{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &bansv1alpha1.BansSlice{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &bansv1alpha1.BandwidthSlice{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &bansv1alpha1.BansSlice{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBansSlice implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBansSlice{}

// ReconcileBansSlice reconciles a BansSlice object
type ReconcileBansSlice struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BansSlice object and makes changes based on the state read
// and what is in the BansSlice.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBansSlice) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger.Info("Reconciling BansSlice", "Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the BansSlice instance
	instance := &bansv1alpha1.BansSlice{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Return all bansslice in the request namespace
	banssliceList := &bansv1alpha1.BansSliceList{}
	opts := []client.ListOption{
		client.InNamespace(request.NamespacedName.Namespace),
	}
	err = r.client.List(context.TODO(), banssliceList, opts...)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Validate reconcile request
	for _, item := range banssliceList.Items {
		if reflect.DeepEqual(*instance, item) {
			// Request itself
			free5gcslice := &bansv1alpha1.Free5GCSlice{}
			err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-free5gcslice", Namespace: instance.Namespace}, free5gcslice)
			if err != nil && errors.IsNotFound(err) {
				// No Free5GCSlice found
				// Break the loop and create new Free5GCSlice and BandwidthSlice objects
				break
			} else if err != nil {
				reqLogger.Error(err, "Failed to get Free5GCSlice")
				return reconcile.Result{}, err
			}
			// Free5GCSlice exists
			return reconcile.Result{}, nil
		} else if reflect.DeepEqual(instance.Spec, item.Spec) {
			// BansSlice instance with same BansSliceSpec
			// Return and don't requeue
			reqLogger.Info("BansSlice instance with same BansSliceSpec exists", "BansSlice.Name", item.Name)
			return reconcile.Result{}, nil
		} else if reflect.DeepEqual(instance.Spec.SnssaiList, item.Spec.SnssaiList) {
			// BansSlice instance with same BansSliceSpec.SnssaiList
			reqLogger.Info("BansSlice instance with same BansSliceSpec.SnssaiList exists", "BansSlice.Name", item.Name)
			reqLogger.Info("Reconfiguring BandwidthSliceSpec", "MinRate", instance.Spec.MinRate, "MaxRate", instance.Spec.MaxRate)
			// TODO(dev): Reconfigure BandwidthSliceSpec
			return reconcile.Result{}, nil
		}
	}

	// Create new Free5GCSlice
	free5gcslice := newFree5GCSlice(instance)

	// Set BansSlice instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, free5gcslice, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Creating new free5GCSlice", "Namespace", instance.Namespace, "Name", instance.Name+"-free5gcslice")
	err = r.client.Create(context.TODO(), free5gcslice)
	if err != nil {
		reqLogger.Error(err, "Failed to create new Free5GCSlice", "Namespace", free5gcslice.Namespace, "Name", free5gcslice.Name)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// newFree5GCSlice returns a new Free5GCSlice object with BansSliceSpec
func newFree5GCSlice(b *bansv1alpha1.BansSlice) *bansv1alpha1.Free5GCSlice {
	labels := map[string]string{
		"app": b.Name,
	}
	return &bansv1alpha1.Free5GCSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name + "-free5gcslice",
			Namespace: b.Namespace,
			Labels:    labels,
		},
		Spec: bansv1alpha1.Free5GCSliceSpec{
			SnssaiList: b.Spec.SnssaiList,
			GNBAddr:    b.Spec.GNBAddr,
		},
	}
}
