package bansslice

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"

	bansv1alpha1 "github.com/stevenchiu30801/bans5gc-operator/pkg/apis/bans/v1alpha1"
	"golang.org/x/net/http2"
	corev1 "k8s.io/api/core/v1"
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

// State of Free5GCSlice
const (
	StateNull     string = ""
	StateCreating string = "Creating"
	StateRunning  string = "Running"
)

// IP protocol number
const (
	UDPProtocol  uint8 = 17
	SCTPProtocol uint8 = 132
)

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

	// Check if Free5GCSlice of corresponding BansSlice already exists
	free5gcslice := &bansv1alpha1.Free5GCSlice{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-free5gcslice", Namespace: instance.Namespace}, free5gcslice)
	if err != nil && errors.IsNotFound(err) {
		// No Free5GCSlice found
		// Continue reconciling
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Free5GCSlice")
		return reconcile.Result{}, err
	} else {
		// Free5GCSlice exists
		// Check if BansSliceSpec.SnssaiList or BansSliceSpec.GNBAddr is reconfigured
		if !reflect.DeepEqual(instance.Spec.SnssaiList, free5gcslice.Spec.SnssaiList) || instance.Spec.GNBAddr != free5gcslice.Spec.GNBAddr {
			// Remove the original Free5GCSlice and BandwidthSlice for reconfiguration
			err := r.client.Delete(context.TODO(), free5gcslice, client.GracePeriodSeconds(5))
			if err != nil {
				return reconcile.Result{}, err
			}

			// Wait for Free5GCSlice being removed
			for {
				free5gcslice := &bansv1alpha1.Free5GCSlice{}
				err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-free5gcslice", Namespace: instance.Namespace}, free5gcslice)
				if err != nil && errors.IsNotFound(err) {
					break
				} else if err != nil {
					return reconcile.Result{}, err
				}
				time.Sleep(1 * time.Second)
			}

			bandwidthslice := &bansv1alpha1.BandwidthSlice{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-bandwidthslice", Namespace: instance.Namespace}, bandwidthslice)
			if err != nil && errors.IsNotFound(err) {
				// No BandwidthSlice exists
				reqLogger.Info("No BandwidthSlice exists for BansSlice", "BansSlice.Namespace", instance.Namespace, "BansSlice.Name", instance.Name)
			} else if err != nil {
				return reconcile.Result{}, err
			} else {
				err := r.client.Delete(context.TODO(), bandwidthslice, client.GracePeriodSeconds(5))
				if err != nil {
					return reconcile.Result{}, err
				}
			}
		} else {
			// BansSliceSpec.SnssaiList and BansSliceSpec.GNBAddr are the same
			// Reconfigure BandwidthSlice if needed
			bandwidthslice := &bansv1alpha1.BandwidthSlice{}
			err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-bandwidthslice", Namespace: instance.Namespace}, bandwidthslice)
			if err != nil && errors.IsNotFound(err) {
				// No BandwidthSlice exists
				reqLogger.Info("No BandwidthSlice exists for BansSlice", "BansSlice.Namespace", instance.Namespace, "BansSlice.Name", instance.Name)
				return reconcile.Result{Requeue: true}, nil
			} else if err != nil {
				return reconcile.Result{}, err
			}

			targetBandwidthSlice := newBandwidthSlice(instance, free5gcslice)
			if !reflect.DeepEqual(targetBandwidthSlice.Spec, bandwidthslice.Spec) {
				// Update BandwidthSlice
				reqLogger.Info("Reconfiguring BandwidthSliceSpec", "MinRate", instance.Spec.MinRate, "MaxRate", instance.Spec.MaxRate)
				bandwidthslice.Spec = targetBandwidthSlice.Spec
				if err := r.client.Update(context.Background(), bandwidthslice); err != nil {
					return reconcile.Result{}, err
				}
				reqLogger.Info("Successfully reconfigure BandwidthSlice", "Namespace", bandwidthslice.Namespace, "Name", bandwidthslice.Name)
			}

			return reconcile.Result{}, nil
		}
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
			continue
		} else if reflect.DeepEqual(instance.Spec, item.Spec) {
			// BansSlice instance with same BansSliceSpec
			// Return and don't requeue
			reqLogger.Info("BansSlice instance with same BansSliceSpec exists", "BansSlice.Name", item.Name)
			return reconcile.Result{}, nil
		} else if reflect.DeepEqual(instance.Spec.SnssaiList, item.Spec.SnssaiList) && instance.Spec.GNBAddr == item.Spec.GNBAddr {
			// BansSlice instance with same BansSliceSpec.SnssaiList
			reqLogger.Info("BansSlice instance with same BansSliceSpec.SnssaiList and BansSliceSpec.GNBAddr exists", "BansSlice.Name", item.Name)

			// Check if BandwidthSlice already exists
			bandwidthslice := &bansv1alpha1.BandwidthSlice{}
			err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-bandwidthslice", Namespace: instance.Namespace}, bandwidthslice)
			if err != nil && errors.IsNotFound(err) {
				// No BandwidthSlice exists
				reqLogger.Info("No BandwidthSlice exists for BansSlice", "BansSlice.Namespace", instance.Namespace, "BansSlice.Name", instance.Name)
				// Create new BandwidthSlice only for free5GC slices
				free5gcslice := newFree5GCSlice(instance)
				bandwidthslice := newBandwidthSlice(instance, free5gcslice)
				// Set BansSlice instance as the owner and controller
				if err := controllerutil.SetControllerReference(instance, bandwidthslice, r.scheme); err != nil {
					return reconcile.Result{}, err
				}
				reqLogger.Info("Creating new BandwidthSlice", "Namespace", instance.Namespace, "Name", instance.Name+"-bandwidthslice")
				err = r.client.Create(context.TODO(), bandwidthslice)
				if err != nil {
					reqLogger.Error(err, "Failed to create new BandwidthSlice", "Namespace", bandwidthslice.Namespace, "Name", bandwidthslice.Name)
					return reconcile.Result{}, err
				}

				reqLogger.Info("Successfully create new BandwidthSlice", "Namespace", bandwidthslice.Namespace, "Name", bandwidthslice.Name)
				return reconcile.Result{}, nil
			} else if err != nil {
				return reconcile.Result{}, err
			}
			// BandwidthSlice exists
			return reconcile.Result{}, nil
		}
	}

	// Create new Free5GCSlice
	free5gcslice = newFree5GCSlice(instance)
	// Set BansSlice instance as the owner and controller
	if err = controllerutil.SetControllerReference(instance, free5gcslice, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	reqLogger.Info("Creating new free5GCSlice", "Namespace", instance.Namespace, "Name", instance.Name+"-free5gcslice")
	err = r.client.Create(context.TODO(), free5gcslice)
	if err != nil {
		reqLogger.Error(err, "Failed to create new Free5GCSlice", "Namespace", free5gcslice.Namespace, "Name", free5gcslice.Name)
		return reconcile.Result{}, err
	}

	// Wait for Free5GCSlice object being running
	startTime := time.Now()
	for {
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-free5gcslice", Namespace: instance.Namespace}, free5gcslice)
		if err != nil && errors.IsNotFound(err) {
			// No Free5GCSlice found
			reqLogger.Info("Free5GCSlice not found after created", "Namespace", instance.Namespace, "Name", instance.Name+"-free5gcslice")
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			reqLogger.Error(err, "Failed to get Free5GCSlice")
			return reconcile.Result{}, err
		}
		// Free5GCSlice exists
		if free5gcslice.Status.State == StateCreating {
			reqLogger.Info("Waiting 3 seconds for Free5GCSlice to be created", "Namespace", free5gcslice.Namespace, "Name", free5gcslice.Name)
		} else if free5gcslice.Status.State == StateRunning {
			endTime := time.Now()
			elapsed := endTime.Sub(startTime)
			msg := "Successfully create Free5GCSlice in " + strconv.FormatFloat(float64(elapsed)/float64(time.Second), 'f', 2, 64) + " seconds"
			reqLogger.Info(msg, "Namespace", free5gcslice.Namespace, "Name", free5gcslice.Name)
			break
		} else {
			// StateNull
		}
		time.Sleep(3 * time.Second)
	}

	// Create new BandwidthSlice for free5GC slices
	bandwidthslice := newBandwidthSlice(instance, free5gcslice)
	// Set BansSlice instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, bandwidthslice, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	reqLogger.Info("Creating new BandwidthSlice", "Namespace", instance.Namespace, "Name", instance.Name+"-bandwidthslice")
	err = r.client.Create(context.TODO(), bandwidthslice)
	if err != nil {
		reqLogger.Error(err, "Failed to create new BandwidthSlice", "Namespace", bandwidthslice.Namespace, "Name", bandwidthslice.Name)
		return reconcile.Result{}, err
	}
	reqLogger.Info("Successfully create new BandwidthSlice", "Namespace", bandwidthslice.Namespace, "Name", bandwidthslice.Name)

	// TODO(dev): Wait for ONOS Bandwidth Management flows being added

	// Return all NSSFs
	nssfList := &corev1.PodList{}
	opts = []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(map[string]string{"app.kubernetes.io/instance": "free5gc", "app.kubernetes.io/name": "nssf"}),
	}
	err = r.client.List(context.TODO(), nssfList, opts...)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Configure all NSSFs through management REST API
	client := &http.Client{}
	client.Transport = &http2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	nssfManagementDocument, err := r.newNssfManagementDocument(instance)
	if err != nil {
		reqLogger.Error(err, "Cannot generate NSSF management object")
		return reconcile.Result{}, nil
	}
	buf, err := json.Marshal(nssfManagementDocument)
	if err != nil {
		reqLogger.Error(err, "Cannot marshal NssfManagementDocument to JSON format")
		return reconcile.Result{}, nil
	}
	for _, nssf := range nssfList.Items {
		nssfIp := nssf.Status.PodIP
		reqLogger.Info("Configuring NSSF with new S-NSSAI list through management REST API", "PodIP", nssfIp, "S-NSSAIList", instance.Spec.SnssaiList)
		req, err := http.NewRequest("POST",
			"https://"+nssfIp+":29531/nnssf-management/v1/network-slice-information",
			bytes.NewReader(buf))
		if err != nil {
			return reconcile.Result{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Read response from NSSF
		if resp.StatusCode == http.StatusCreated {
			reqLogger.Info("Successfully configure NSSF", "PodIP", nssfIp)
		} else {
			defer resp.Body.Close()
			buf, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("Fail to configure NSSF", "ResponseBody", buf)
		}
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

// newBandwidthSlice returns a new BandwidthSlice object with BansSliceSpec and Free5GCSliceStatus
func newBandwidthSlice(b *bansv1alpha1.BansSlice, f *bansv1alpha1.Free5GCSlice) *bansv1alpha1.BandwidthSlice {
	labels := map[string]string{
		"app": b.Name,
	}
	var (
		priority uint = 2
		minRate  uint = b.Spec.MinRate
		maxRate  uint = b.Spec.MaxRate
	)
	return &bansv1alpha1.BandwidthSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name + "-bandwidthslice",
			Namespace: b.Namespace,
			Labels:    labels,
		},
		Spec: bansv1alpha1.BandwidthSliceSpec{
			Slices: []bansv1alpha1.Slice{
				// Control traffic slice
				{
					Priority: &priority,
					Flows: []bansv1alpha1.Flow{
						// Uplink
						{
							SrcAddr:  b.Spec.GNBAddr,
							DstAddr:  f.Status.AmfAddr,
							Protocol: SCTPProtocol,
						},
						// Downlink
						{
							SrcAddr:  f.Status.AmfAddr,
							DstAddr:  b.Spec.GNBAddr,
							Protocol: SCTPProtocol,
						},
					},
				},
				// Downlink data traffic slice
				{
					MinRate: &minRate,
					MaxRate: &maxRate,
					Flows: []bansv1alpha1.Flow{
						{
							SrcAddr:  f.Status.UpfAddr,
							DstAddr:  b.Spec.GNBAddr,
							Protocol: UDPProtocol,
						},
					},
				},
			},
		},
	}
}
