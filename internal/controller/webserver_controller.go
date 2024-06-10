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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	ers "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ldhctlrv1alpha1 "github.com/babugeet/test-webserver-operator/api/v1alpha1"
)

// WebserverReconciler reconciles a Webserver object
type WebserverReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ldhctlr.linuxdatahub.com,resources=webservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ldhctlr.linuxdatahub.com,resources=webservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ldhctlr.linuxdatahub.com,resources=webservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Webserver object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *WebserverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	//
	// 1. Get the webserver instance
	Webserver := &ldhctlrv1alpha1.Webserver{}

	if err := r.Get(ctx, req.NamespacedName, Webserver); err != nil {
		if ers.IsNotFound(err) {
			log.Log.Info("Webserver resource not found. Ignoring since object must be deleted")
			// r.Log.Info().Msg("CloudScheduler resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. Set the owner reference

	deployment := WebserverDeployment(Webserver)

	if err := controllerutil.SetControllerReference(Webserver, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// 2.Get the desired state from the resource spec

	// desired := Webserver.Spec

	// 3.  Get the current state of the deployment (replace with your logic)
	_, err := GetDeployment(ctx, r.Client, req.NamespacedName)
	if err != nil && client.IgnoreNotFound(err) != nil {
		log.Log.Error(err, "Failed to get deployment")
		return ctrl.Result{}, err
	}
	if err != nil && client.IgnoreNotFound(err) == nil {
		fmt.Println("Creating deployment")
		if err := r.Client.Create(ctx, deployment); err != nil {
			fmt.Println(err.Error())
			return ctrl.Result{}, err
		}

	}

	Webserver.Status.Deployment = deployment.Name
	Webserver.Status.ServiceType = "NodePort"
	Webserver.Status.WebUrl = "to be updated"

	err = r.Update(ctx, Webserver)

	if err != nil {
		fmt.Println("Unable to update the status")
		return ctrl.Result{}, err
	}
	//  4. Get the current state of the service

	// 5. Get current state of ingress
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebserverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ldhctlrv1alpha1.Webserver{}).
		Complete(r)
}

func WebserverDeployment(ws *ldhctlrv1alpha1.Webserver) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ws-test", ws.Name),
			Namespace: ws.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ws.Spec.Replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": fmt.Sprintf("%s-ws-test", ws.Name),
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": fmt.Sprintf("%s-ws-test", ws.Name),
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  fmt.Sprintf("%s-ws-test", ws.Name),
							Image: ws.Spec.Image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

}

func GetDeployment(ctx context.Context, client client.Client, namespace types.NamespacedName) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := client.Get(ctx, namespace, deployment)
	return deployment, err
}
