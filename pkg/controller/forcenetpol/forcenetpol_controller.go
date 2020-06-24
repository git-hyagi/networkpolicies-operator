package forcenetpol

import (
	"context"
	"reflect"

	labv1 "github.com/lab/networkpolicies-operator/pkg/apis/lab/v1"
	netv1 "k8s.io/api/networking/v1"
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

var log = logf.Log.WithName("controller_forcenetpol")

// Add creates a new ForceNetPol Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileForceNetPol{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("forcenetpol-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ForceNetPol
	err = c.Watch(&source.Kind{Type: &labv1.ForceNetPol{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner ForceNetPol
	err = c.Watch(&source.Kind{Type: &netv1.NetworkPolicy{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &labv1.ForceNetPol{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileForceNetPol implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileForceNetPol{}

// ReconcileForceNetPol reconciles a ForceNetPol object
type ReconcileForceNetPol struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ForceNetPol object and makes changes based on the state read
// and what is in the ForceNetPol.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileForceNetPol) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ForceNetPol")

	// Fetch the ForceNetPol instance
	instance := &labv1.ForceNetPol{}
	err := r.client.Get(context.TODO(), client.ObjectKey{Name: "forcenetpol", Namespace: ""}, instance)
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

	for _, namespace := range instance.Spec.Projects {
		netwpol := networkPolicyAllowFromSameNamespace(namespace)

		// Set this instance as the owner and controller of the new network policy object
		if err := controllerutil.SetControllerReference(instance, netwpol, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this netwpol already exists
		obj_type := &netv1.NetworkPolicy{}

		err = r.client.Get(context.TODO(), types.NamespacedName{Name: netwpol.Name, Namespace: netwpol.Namespace}, obj_type)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new NetworkPolicy (allow-from-same-namespace)", "NetwPol.Namespace", netwpol.Namespace, "NetwPol.Name", netwpol.Name)

			err = r.client.Create(context.TODO(), netwpol)
			if err != nil {
				return reconcile.Result{}, err
			}

			// NetworkPolicy created successfully - don't requeue
			return reconcile.Result{}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Check if it wasn't modified
		expected_spec := networkPolicyAllowFromSameNamespaceSpec()

		if !reflect.DeepEqual(expected_spec, obj_type.Spec) {
			reqLogger.Info("The NetworkPolicy allow-from-same-namespace is not matching the one from operator! Reconciling ...")
			err = r.client.Update(context.TODO(), netwpol)
			if err != nil {
				reqLogger.Error(err, "Error trying to update the NetworkPolicy allow-from-same-namespace object ... ")
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}

		/* BEGIN DENY-BY-DEFAULT */
		netwpol = networkPolicyDenyByDefault(namespace)

		// Set this instance as the owner and controller of the new network policy object
		if err = controllerutil.SetControllerReference(instance, netwpol, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this netwpol already exists
		obj_type = &netv1.NetworkPolicy{}

		err = r.client.Get(context.TODO(), types.NamespacedName{Name: netwpol.Name, Namespace: netwpol.Namespace}, obj_type)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new NetworkPolicy (deny-by-default)", "NetwPol.Namespace", netwpol.Namespace, "NetwPol.Name", netwpol.Name)

			err = r.client.Create(context.TODO(), netwpol)
			if err != nil {
				return reconcile.Result{}, err
			}

			// NetworkPolicy created successfully - don't requeue
			return reconcile.Result{}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Check if it wasn't modified
		expected_spec = networkPolicyDenyByDefaultSpec()

		if !reflect.DeepEqual(expected_spec, obj_type.Spec) {
			reqLogger.Info("The NetworkPolicy deny-by-default is not matching the one from operator! Reconciling ...")
			err = r.client.Update(context.TODO(), netwpol)
			if err != nil {
				reqLogger.Error(err, "Error trying to update the NetworkPolicy object deny-by-default ... ")
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
		/* END DENY-BY-DEFAULT */

	}
	return reconcile.Result{}, nil
}

// return a network policy object of "type" deny-by-default
func networkPolicyDenyByDefault(project string) *netv1.NetworkPolicy {
	return &netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{Kind: "NetworkPolicy", APIVersion: "extensions/v1beta"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deny-by-default",
			Namespace: project,
			Labels: map[string]string{
				"operator": "force-netpol-operator",
			},
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress:     []netv1.NetworkPolicyIngressRule{{}},
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{"Ingress"},
		},
	}
}

// return a network policy object of "type" allow-from-same-namespace
func networkPolicyAllowFromSameNamespace(project string) *netv1.NetworkPolicy {
	return &netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{Kind: "NetworkPolicy", APIVersion: "extensions/v1beta"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-from-same-namespace",
			Namespace: project,
			Labels: map[string]string{
				"operator": "force-netpol-operator",
			},
		},
		Spec: netv1.NetworkPolicySpec{
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
						},
					},
				},
			},
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{"Ingress"},
		},
	}
}

// expected spec for the allow-from-same-namespace network policy
func networkPolicyAllowFromSameNamespaceSpec() netv1.NetworkPolicySpec {
	return netv1.NetworkPolicySpec{
		Ingress: []netv1.NetworkPolicyIngressRule{
			{
				From: []netv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{},
					},
				},
			},
		},
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []netv1.PolicyType{"Ingress"},
	}
}

// expected spec for the deny-by-default network policy
func networkPolicyDenyByDefaultSpec() netv1.NetworkPolicySpec {
	return netv1.NetworkPolicySpec{
		Ingress:     []netv1.NetworkPolicyIngressRule{{}},
		PodSelector: metav1.LabelSelector{},
		PolicyTypes: []netv1.PolicyType{"Ingress"},
	}
}
