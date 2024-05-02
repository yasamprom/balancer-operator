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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	balancer "github.com/yasamprom/balancer-operator/api/v1"
)

var _ = Describe("Balancer Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "balancer"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		b := &balancer.Balancer{
			Spec: balancer.BalancerSpec{
				Replicas: 1,
			},
		}

		testLabels := make(map[string]string)
		testLabels["custom-label"] = "value"
		BeforeEach(func() {
			By("creating the custom resource for the Kind Balancer")
			err := k8sClient.Get(ctx, typeNamespacedName, b)
			if err != nil && errors.IsNotFound(err) {
				d := appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &b.Spec.Replicas,
						Selector: &metav1.LabelSelector{
							MatchLabels: testLabels,
						},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels:    testLabels,
								Namespace: "default",
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
				Expect(k8sClient.Create(ctx, &d)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &BalancerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
