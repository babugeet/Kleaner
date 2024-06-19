package controller

import (
	"context"
	"fmt"

	ldhctlrv1alpha1 "github.com/babugeet/test-webserver-operator/api/v1alpha1"
	"github.com/babugeet/test-webserver-operator/variables"
	net "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetIngress(ctx context.Context, client client.Client, namespace types.NamespacedName) (*net.Ingress, error) {
	ingress := &net.Ingress{}
	err := client.Get(ctx, namespace, ingress)
	return ingress, err
}

func WebServerIngress(ws *ldhctlrv1alpha1.Webserver, serviceName string) *net.Ingress {
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
									Name: serviceName,
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

func (r *WebserverReconciler) ReconcileIngress(ctx context.Context, Webserver *ldhctlrv1alpha1.Webserver) (ctrl.Result, error) {
	// 1. Create ingress instance
	Ingress := WebServerIngress(Webserver, variables.ServiceName)

	// 2. Set the owner  reference
	if err := controllerutil.SetControllerReference(Webserver, Ingress, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	// 3. Get the existing ingress, if not found create one
	Ing, err := GetIngress(ctx, r.Client, types.NamespacedName{Name: Ingress.Name, Namespace: Ingress.Namespace})
	if err != nil && client.IgnoreNotFound(err) == nil {
		variables.Found = false
		fmt.Println("Creating ingress")
		fmt.Println(err)
		if err := r.Client.Create(ctx, Ingress); err != nil {
			fmt.Println(err.Error())
			return ctrl.Result{}, err
		}
	} else {
		variables.Found = true
	}

	// 4. Update the status of the wss
	if variables.Found {
		Webserver.Status.WebUrl = Ing.Spec.Rules[0].HTTP.Paths[0].Path
	} else {
		Webserver.Status.WebUrl = "Ingress not found"
	}
	err = r.Status().Update(ctx, Webserver)
	fmt.Println("status updated")
	if err != nil {
		fmt.Println("Unable to update the status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, err
}
