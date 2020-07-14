/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	labv1 "apodemakeles/k8s/gc/api/v1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=lab.apodemas,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=lab.apodemas,resources=applications/status,verbs=get;update;patch

func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("application", req.NamespacedName)

	logger.Info("create a new application")
	applicaiton := &labv1.Application{}
	err := r.Get(ctx, req.NamespacedName, applicaiton)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if applicaiton.ObjectMeta.DeletionTimestamp.IsZero() {
		result, err := r.createService(ctx, req, applicaiton)

		return result, err
	} else {
		//service := &v1.Service{}
		//err := r.Get(ctx, req.NamespacedName, service)
		//if errors.IsNotFound(err){
		//	r.Log.Info("no service")
		//}else if err == nil{
		//
		//}
	}

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&labv1.Application{}).
		Complete(r)
}

func getOwnerReference(application *labv1.Application) *v12.OwnerReference {
	return v12.NewControllerRef(application, schema.GroupVersionKind{
		Group:   labv1.GroupVersion.Group,
		Version: labv1.GroupVersion.Version,
		Kind:    "Application",
	})
}

func (r *ApplicationReconciler) createService(ctx context.Context, req ctrl.Request, application *labv1.Application) (ctrl.Result, error) {
	service := &v1.Service{}
	service.Name = application.Name
	service.Namespace = application.Namespace
	service.Spec.Selector = map[string]string{
		"app": "nobody",
	}
	port := v1.ServicePort{
		Protocol:   v1.ProtocolTCP,
		Port:       80,
		TargetPort: intstr.Parse("8080"),
	}
	service.Spec.Ports = append(service.Spec.Ports, port)

	reference := getOwnerReference(application)
	*reference.Controller = false
	service.ObjectMeta.OwnerReferences = []v12.OwnerReference{
		*reference,
	}

	err := r.Client.Create(ctx, service)

	return ctrl.Result{}, err
}
