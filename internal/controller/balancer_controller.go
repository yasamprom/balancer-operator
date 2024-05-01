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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	balancer "github.com/yasamprom/balancer-operator/api/v1"
	"github.com/yasamprom/balancer-operator/internal/model"
	appsv1 "k8s.io/api/apps/v1"
)

// BalancerReconciler reconciles a Balancer object
type BalancerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Uc     model.Usecases
}

//+kubebuilder:rbac:groups=apps.yasamprom.com,resources=balancers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.yasamprom.com,resources=balancers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.yasamprom.com,resources=balancers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Balancer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *BalancerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	logger := log.Log.WithValues("balancer", req.NamespacedName)
	logger.Info("Reconcile...")

	var app balancer.Balancer

	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch Balancer")
		return ctrl.Result{}, err
	}

	logger.Info("Application is being created...")
	d := r.getDeploymentForBalancer(ctx, &app)

	err := r.Client.Create(ctx, d)
	if err != nil {
		logger.Error(err, "failed to create balancer")
		return ctrl.Result{}, err
	}
	logger.Info("Balancer created...")
	return ctrl.Result{}, nil
}

func (r *BalancerReconciler) getDeploymentForBalancer(_ context.Context, m *balancer.Balancer) *appsv1.Deployment {
	_ = m.GetLabels()

	ls := m.GetLabels()

	d := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "balancer",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &m.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "balancer:v1",
						Name:  "balancer-deployment",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "http",
						}},
					}},
				},
			},
		},
	}
	return &d
}

// SetupWithManager sets up the controller with the Manager.
func (r *BalancerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&balancer.Balancer{}).
		Complete(r)
}
