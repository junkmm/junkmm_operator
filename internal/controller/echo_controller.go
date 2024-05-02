/*
Copyright 2024.

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

package controller

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "junkmm.site/echo/api/v1"
)

// EchoReconciler reconciles a Echo object
type EchoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.junkmm.site,resources=echoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.junkmm.site,resources=echoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.junkmm.site,resources=echoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Echo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *EchoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	var instance appv1.Echo
	errGet := r.Get(ctx, req.NamespacedName, &instance)
	if errGet != nil {
		log.Log.Error(errGet, "Error getting instance")
		return ctrl.Result{}, client.IgnoreNotFound(errGet)
	}

	pod := NewPod(&instance)

	_, errCreate := ctrl.CreateOrUpdate(ctx, r.Client, pod, func() error {
		return ctrl.SetControllerReference(&instance, pod, r.Scheme)
	})

	if errCreate != nil {
		log.Log.Error(errCreate, "Error create pod")
		return ctrl.Result{}, nil
	}

	err := r.Status().Update(context.TODO(), &instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func NewPod(cr *appv1.Echo) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: strings.Split(cr.Spec.Command, " "),
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *EchoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.Echo{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
