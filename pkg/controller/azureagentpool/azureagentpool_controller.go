package azureagentpool

import (
	"context"
	"strconv"
	"time"

	apov1alpha1 "github.com/azure-pipeline-operator/azure-pipeline-operator/pkg/apis/apo/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

	"github.com/benmatselby/go-azuredevops/azuredevops"
)

var log = logf.Log.WithName("controller_azureagentpool")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AzureAgentPool Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAzureAgentPool{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("azureagentpool-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AzureAgentPool
	err = c.Watch(&source.Kind{Type: &apov1alpha1.AzureAgentPool{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AzureAgentPool
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &apov1alpha1.AzureAgentPool{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAzureAgentPool implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAzureAgentPool{}

// ReconcileAzureAgentPool reconciles a AzureAgentPool object
type ReconcileAzureAgentPool struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

//Polling period
var period = 5 * time.Second

// Reconcile reads that state of the cluster for a AzureAgentPool object and makes changes based on the state read
// and what is in the AzureAgentPool.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAzureAgentPool) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AzureAgentPool")

	// Fetch the AzureAgentPool instance
	instance := &apov1alpha1.AzureAgentPool{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{RequeueAfter: period}, err
	}
	reqLogger.Info("AzureAgentPool", "Account", instance.Spec.Account, "Project", instance.Spec.Project, "AgentPool", instance.Spec.AgentPool)

	/************
	* Reconcile main agent pod for the pool
	************/
	// Define a new Pod object
	pod := newMainAgentForPool(instance)
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{RequeueAfter: period}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{RequeueAfter: period}, err
		}
	} else if err != nil {
		return reconcile.Result{RequeueAfter: period}, err
	}
	// Pod already exists
	reqLogger.Info("Skip reconcile: Kubernetes-main pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)

	/************
	* Reconcile agent pods for build
	************/
	//TODO: Query one project or all where the pool is assigned?
	//TODO: Query the builds or maybe the required agents?

	//Project specific client
	v := azuredevops.NewClient(instance.Spec.Account, instance.Spec.Project, instance.Spec.AccessToken)

	//TODO: TEMP - Log all builds;
	// allbuilds, err := v.Builds.List(&azuredevops.BuildsListOptions{})
	// if err != nil {
	// 	reqLogger.Info("Error", "err", err.Error())
	// 	return reconcile.Result{RequeueAfter: period}, err
	// }
	// reqLogger.Info("AllBuilds", "len", len(allbuilds))
	// for _, b := range allbuilds {
	// 	reqLogger.Info("AllBuild", "BuildNumber", b.BuildNumber, "Id", b.ID, "Status", b.Status, "Name", b.Definition.Name)
	// }

	// Query builds with status "inProgress" in the projects
	builds, err := v.Builds.List(&azuredevops.BuildsListOptions{
		Status: "inProgress",
	})
	if err != nil {
		reqLogger.Info("Error", "err", err.Error())
		return reconcile.Result{RequeueAfter: period}, err
	}

	// Check each build if they have a pod created
	reqLogger.Info("Pending builds", "len", len(builds))
	for _, b := range builds {
		reqLogger.Info("Build", "BuildNumber", b.BuildNumber, "Id", b.ID, "Status", b.Status, "Name", b.Definition.Name)

		// Define a new Pod object
		pod := podForBuild(&b, instance)
		if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
			return reconcile.Result{RequeueAfter: period}, err
		}

		// Check if this Pod already exists
		found := &corev1.Pod{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err != nil {
			if errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)

				// Create Pod, ignore is already exists
				err = r.client.Create(context.TODO(), pod)
				if err != nil {
					return reconcile.Result{RequeueAfter: period}, err
				}

			} else { // Other error than "NotFound"
				return reconcile.Result{RequeueAfter: period}, err
			}
		}
	}

	// Look for pods and if build is completed then delete the Pod and agent.
	// List all pods owned by this AzureAgentPool instance
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(
		map[string]string{
			"devops.example.com/azure-agent-pool": instance.Name,
		})
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	if err = r.client.List(context.TODO(), listOps, podList); err != nil {
		return reconcile.Result{}, err
	}

	// Check if pod has a matching build, delete otherwise
	for _, pod := range podList.Items {
		reqLogger.Info("Pod", "Name", pod.Name, "Phase", pod.Status.Phase)
		if pod.ObjectMeta.DeletionTimestamp != nil {
			continue
		}
		buildid := pod.Labels["devops.example.com/buildid"]
		found := false
		for _, b := range builds {
			if strconv.Itoa(b.ID) == buildid {
				found = true
			}
		}
		if !found {
			r.client.Delete(context.TODO(), &pod)
		}
	}

	//TODO: Delete agent from Azure Pool; Or let the agent deregister with postHook (what happens is kubernetes moves the pod?)

	//Try priodic reconcile
	return reconcile.Result{RequeueAfter: period}, nil
}

func newMainAgentForPool(cr *apov1alpha1.AzureAgentPool) *corev1.Pod {
	labels := map[string]string{
		"azureagentpool": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-" + cr.Name + "-kubernetes-main",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "agent",
					Image: "quay.io/bszeti/azure-pipeline-agent-oc",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVER_URL",
							Value: "https://dev.azure.com/" + cr.Spec.Account,
						},
						{
							Name:  "ACCESS_TOKEN",
							Value: cr.Spec.AccessToken,
						},
						{
							Name:  "AGENT_POOL",
							Value: cr.Spec.AgentPool,
						},
						{
							Name:  "AGENT_NAME",
							Value: "kubernetes-main",
						},
						{
							Name:  "KUBERNETES_AGENT_TYPE",
							Value: "main",
						},
					},
				},
			},
		},
	}
}

func podForBuild(build *azuredevops.Build, cr *apov1alpha1.AzureAgentPool) *corev1.Pod {
	labels := map[string]string{
		"devops.example.com/azure-agent-pool": cr.Name,
		"devops.example.com/buildid":          strconv.Itoa(build.ID),
	}
	//Should not use annotations because the ListOptions doesn't support annotations
	// annotations := map[string]string{
	// 	"devops.example.com/azure-agent-pool": cr.Name,
	// 	"devops.example.com/buildid": strconv.Itoa(build.ID),
	// }
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-" + cr.Name + "-kubernetes-build-" + strconv.Itoa(build.ID),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "agent",
					Image: "quay.io/bszeti/azure-pipeline-agent-oc",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVER_URL",
							Value: "https://dev.azure.com/" + cr.Spec.Account,
						},
						{
							Name:  "ACCESS_TOKEN",
							Value: cr.Spec.AccessToken,
						},
						{
							Name:  "AGENT_POOL",
							Value: cr.Spec.AgentPool,
						},
						{
							Name:  "AGENT_NAME",
							Value: "agent-build-" + strconv.Itoa(build.ID),
						},
						{
							Name:  "KUBERNETES_AGENT_TYPE",
							Value: "java",
						},
						{
							Name: "BUILD_BUILDID",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.labels['devops.example.com/buildid']",
								},
							},
						},
					},
				},
			},
		},
	}
}