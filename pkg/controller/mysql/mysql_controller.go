package mysql

import (
	"context"
	"encoding/json"
	appsv1 "k8s.io/api/apps/v1"
	"reflect"

	appv1 "operator/mysql/pkg/apis/app/v1"
	"operator/mysql/pkg/resources"

	mylog "github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_mysql")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MySQL Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMySQL{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("mysql-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MySQL
	err = c.Watch(&source.Kind{Type: &appv1.MySQL{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MySQL
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1.MySQL{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMySQL implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMySQL{}

// ReconcileMySQL reconciles a MySQL object
type ReconcileMySQL struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MySQL object and makes changes based on the state read
// and what is in the MySQL.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMySQL) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MySQL")

	// Fetch the MySQL mysql
	mysql := &appv1.MySQL{}
	err := r.client.Get(context.TODO(), request.NamespacedName, mysql)
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
	if mysql.DeletionTimestamp != nil {
		return reconcile.Result{}, err
	}
	mylog.Infoln("MySQL.Name: ", mysql.Name)
	mylog.Infoln("MySQL.Image: ", mysql.Spec.Image)
	mylog.Infoln("MySQL.Replicas: ", mysql.Spec.Replicas)
	mylog.Infoln("MySQL.RootPassword: ", mysql.Spec.RootPassword)
	// 查看传入数据是否符合预期
	deployment := &appsv1.Deployment{}
	// 看看这个Deployment是不是已经存在了
	if err := r.client.Get(context.TODO(), request.NamespacedName, deployment); err != nil && errors.IsNotFound(err) {
		// 不存在的话就新建
		deployment := resources.MySQLDeployment(mysql)
		if err := r.client.Create(context.TODO(), deployment); err != nil {
			return reconcile.Result{}, err
		}
		service := resources.MySQLService(mysql)
		if err := r.client.Create(context.TODO(), service); err != nil {
			return reconcile.Result{}, err
		}
		data, _ := json.Marshal(mysql.Spec)
		if mysql.Annotations != nil {
			mysql.Annotations["spec"] = string(data)
		} else {
			mysql.Annotations = map[string]string{
				"spec": string(data),
			}
		}
		if err := r.client.Update(context.TODO(), mysql); err != nil {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, nil
	}
	originalSpec := appv1.MySQLSpec{}
	if err := json.Unmarshal([]byte(mysql.Annotations["spec"]), &originalSpec); err != nil {
		return reconcile.Result{}, err
	}
	if !reflect.DeepEqual(mysql.Spec, originalSpec) {
		mysqlDeploy := resources.MySQLDeployment(mysql)
		originalDeploy := &appsv1.Deployment{}
		if err := r.client.Get(context.TODO(), request.NamespacedName, originalDeploy); err != nil {
			return reconcile.Result{}, err
		}
		originalDeploy.Spec = mysqlDeploy.Spec
		// 用我们期望的MySQL Deployment Spec去替代默认的Deployment Spec
		if err := r.client.Update(context.TODO(), originalDeploy); err != nil {
			return reconcile.Result{}, nil
		}
		mysqlService := resources.MySQLService(mysql)
		originalService := &corev1.Service{}
		if err := r.client.Get(context.TODO(), request.NamespacedName, originalService); err != nil {
			return reconcile.Result{}, err
		}
		originalService.Spec = mysqlService.Spec
		// Service同
		if err := r.client.Update(context.TODO(), originalService); err != nil {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}
