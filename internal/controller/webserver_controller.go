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
	"math/rand"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1"
	ers "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	var Found bool
	log.Log.Info("Reconcile loop triggered")
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
	webDeploy, err := GetDeployment(ctx, r.Client, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace})
	// if err != nil && client.IgnoreNotFound(err) != nil {
	// 	log.Log.Error(err, "Failed to get deployment")
	// 	return ctrl.Result{}, err
	// }
	fmt.Println(err)
	if err != nil && client.IgnoreNotFound(err) == nil {
		fmt.Println("Creating deployment")
		Found = false
		fmt.Println(err)
		if err := r.Client.Create(ctx, deployment); err != nil {
			fmt.Println(err.Error())
			return ctrl.Result{}, err
		}

	} else {
		Found = true
	}
	// fmt.Println(webDeploy)
	// fmt.Println(*Webserver.Spec.Replica)
	// fmt.Println(Webserver.Spec.Replica)
	// fmt.Println(webDeploy.Spec.Replicas)
	// fmt.Println(*webDeploy.Spec.Replicas)

	if Found && *webDeploy.Spec.Replicas != *Webserver.Spec.Replica {
		fmt.Println("Desired replica count not matching to actual replicas")
		fmt.Println("Triggering update deployment")
		webDeploy.Spec.Replicas = Webserver.Spec.Replica
		err := r.Client.Update(ctx, deployment)
		if err != nil {
			return ctrl.Result{}, err
		}

	}
	// Updating status of Deployment
	Webserver.Status.Deployment = deployment.Name

	service := WebServerService(Webserver, deployment.Name)
	if err := controllerutil.SetControllerReference(Webserver, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	_, err = GetService(ctx, r.Client, types.NamespacedName{Name: service.Name, Namespace: service.Namespace})

	if err != nil && client.IgnoreNotFound(err) != nil {
		log.Log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}
	if err != nil && client.IgnoreNotFound(err) == nil {

		fmt.Println("Creating service")
		if err := r.Client.Create(ctx, service); err != nil {
			log.Log.Error(err, "Failed to create service")
		}
	}
	Webserver.Status.ServiceType = service.Name

	Ingress := WebServerIngress(Webserver, service)
	if err := controllerutil.SetControllerReference(Webserver, Ingress, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	Ing, err := GetIngress(ctx, r.Client, types.NamespacedName{Name: Ingress.Name, Namespace: Ingress.Namespace})
	if err != nil && client.IgnoreNotFound(err) == nil {
		Found = false
		fmt.Println("Creating ingress")
		fmt.Println(err)
		if err := r.Client.Create(ctx, Ingress); err != nil {
			fmt.Println(err.Error())
			return ctrl.Result{}, err
		}

	} else {
		Found = true
	}

	// fmt.Println(Ing)
	if Found {
		Webserver.Status.WebUrl = Ing.Spec.Rules[0].HTTP.Paths[0].Path
	} else {
		Webserver.Status.WebUrl = "Ingress not found"
	}
	// Webserver.Status.WebUrl = "to do"
	err = r.Status().Update(ctx, Webserver)
	fmt.Println("status updated")
	if err != nil {
		fmt.Println("Unable to update the status")
		return ctrl.Result{}, err
	}
	//  4. Get the current state of the service

	// 5. Get current state of ingress
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}
func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
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
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": fmt.Sprintf("%s-ws-test", ws.Name),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  fmt.Sprintf("%s-ws-test", ws.Name),
							Image: ws.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
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

func WebServerService(ws *ldhctlrv1alpha1.Webserver, deployName string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s", ws.Name),
			Namespace: ws.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       *ws.Spec.ServiceType.Port,
					TargetPort: intstr.IntOrString{IntVal: *ws.Spec.ServiceType.TargetPort},
				},
			},
			Selector: map[string]string{"app": deployName},
			Type:     corev1.ServiceType(ws.Spec.ServiceType.Type),
		},
	}
}

func WebServerIngress(ws *ldhctlrv1alpha1.Webserver, service *corev1.Service) *net.Ingress {
	t := net.PathType("Prefix")
	return &net.Ingress{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("ingress-ws-test-%s", ws.Name),
			Namespace: ws.Namespace,
		},
		Spec: net.IngressSpec{
			TLS: []net.IngressTLS{
				{
					Hosts: []string{ws.Spec.Ingress.Host},
				},
			},
			Rules: []net.IngressRule{{
				Host: "",
				IngressRuleValue: net.IngressRuleValue{
					HTTP: &net.HTTPIngressRuleValue{
						Paths: []net.HTTPIngressPath{{
							Path:     ws.Spec.Ingress.Path,
							PathType: &t,
							Backend: net.IngressBackend{
								Service: &net.IngressServiceBackend{
									Name: service.Name,
									Port: net.ServiceBackendPort{
										Number: *ws.Spec.ServiceType.Port,
									},
								},
							},
						}},
					},
				},
			}},
		},
	}
}
func GetDeployment(ctx context.Context, client client.Client, namespace types.NamespacedName) (*appsv1.Deployment, error) {
	fmt.Println("Namsepacedname")
	fmt.Println(namespace)
	deployment := &appsv1.Deployment{}
	err := client.Get(ctx, namespace, deployment)
	return deployment, err
}

func GetService(ctx context.Context, client client.Client, namespace types.NamespacedName) (*corev1.Service, error) {

	service := &corev1.Service{}
	err := client.Get(ctx, namespace, service)
	return service, err
}

func GetIngress(ctx context.Context, client client.Client, namespace types.NamespacedName) (*net.Ingress, error) {
	ingress := &net.Ingress{}
	err := client.Get(ctx, namespace, ingress)
	return ingress, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebserverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ldhctlrv1alpha1.Webserver{}).
		Owns(&appsv1.Deployment{}).Owns(&corev1.Service{}).
		Complete(r)
}
