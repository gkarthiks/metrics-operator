package metrics

import (
	"context"
	"reflect"

	gkarthiksv1alpha1 "github.com/gkarthiks/metrics-operator/pkg/apis/gkarthiks/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_metrics")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Metrics Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMetrics{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("metrics-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Metrics
	err = c.Watch(&source.Kind{Type: &gkarthiksv1alpha1.Metrics{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Metrics
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gkarthiksv1alpha1.Metrics{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMetrics{}

// ReconcileMetrics reconciles a Metrics object
type ReconcileMetrics struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Metrics object and makes changes based on the state read
// and what is in the Metrics.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMetrics) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Metrics")

	// Fetch the Metrics instance
	instance := &gkarthiksv1alpha1.Metrics{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Metrics resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Metrics")
		return reconcile.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Defining a new deployment for creating one
		deploy := r.newDeploymentForMetrics(instance)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deploy.Namespace, "Deployment.Name", deploy.Name)
		err = r.client.Create(context.TODO(), deploy)
		if err != nil {
			reqLogger.Error(err, "Failed to create the new Deployment", "Deployment.Namespace", deploy.Namespace, "Deployment.Name", deploy.Name)
			return reconcile.Result{}, err
		}
		// If no error, the new deployment is created
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get the Deployment")
		return reconcile.Result{}, err
	}

	// Validating the size is same as the spec
	size := instance.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	// Update the Metrics status with the pod names
	// List the pods for this metrics instance's deployment
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForMetrics(instance.Name))
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err = r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods", "Pigops.Namespace", instance.Namespace, "Pigops.Name", instance.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Metrics status")
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// newDeploymentForMetrics returns a metrics Deployment object
func (r *ReconcileMetrics) newDeploymentForMetrics(p *gkarthiksv1alpha1.Metrics) *appsv1.Deployment {
	ls := labelsForMetrics(p.Name)
	replicas := p.Spec.Size
	imagePath := p.Prometheus.Image

	// imagePullPolicy := p.Prometheus.ImagePullPolicy

	// if imagePullPolicy != "" {
	// 	imagePullPolicy = "Always"
	// }
	strReqCPU := p.Prometheus.Resources.Requests.CPU
	strReqMemory := p.Prometheus.Resources.Requests.Memory

	strLimitCPU := p.Prometheus.Resources.Limits.CPU
	strLimitMemory := p.Prometheus.Resources.Limits.Memory

	reqCPU, _ := resource.ParseQuantity(strReqCPU)
	reqMemory, _ := resource.ParseQuantity(strReqMemory)

	limitCPU, _ := resource.ParseQuantity(strLimitCPU)
	limitMemory, _ := resource.ParseQuantity(strLimitMemory)

	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: imagePath,
						// ImagePullPolicy: imagePullPolicy,
						Name: "prometheus",
						Args: []string{"--storage.tsdb.retention=5d", "--config.file=/etc/config/prometheus.yml", "--storage.tsdb.path=/data", "--web.console.libraries=/etc/prometheus/console_libraries", "--web.console.templates=/etc/prometheus/consoles", "--web.enable-lifecycle"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9090,
							Name:          "prometheus",
						}},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    reqCPU,
								corev1.ResourceMemory: reqMemory,
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    limitCPU,
								corev1.ResourceMemory: limitMemory,
							},
						},
					}},
				},
			},
		},
	}
	// Set Pigops instance as the owner and controller
	controllerutil.SetControllerReference(p, deploy, r.scheme)
	return deploy
}

// labelsForMetrics returns the labels for selecting the resources belonging to the given instance CR name.
func labelsForMetrics(name string) map[string]string {
	return map[string]string{"app": "pigops", "pigops_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
