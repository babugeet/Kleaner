package controller

import (
	"context"
	"fmt"

	ldhctlrv1alpha1 "github.com/babugeet/test-webserver-operator/api/v1alpha1"
	"github.com/babugeet/test-webserver-operator/variables"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *WebserverReconciler) ReconcileService(ctx context.Context, Webserver *ldhctlrv1alpha1.Webserver) (ctrl.Result, error) {
	// 1. Create the service instance
	service := WebServerService(Webserver, variables.DeploymentName)
	// 2. Apply owner reference
	if err := controllerutil.SetControllerReference(Webserver, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	// 3. Check the existing service, if not found we are creating it
	_, err := GetService(ctx, r.Client, types.NamespacedName{Name: service.Name, Namespace: service.Namespace})

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

	// 4. Updating service status
	Webserver.Status.ServiceType = service.Name
	err = r.Status().Update(ctx, Webserver)
	fmt.Println("status updated")
	if err != nil {
		fmt.Println("Unable to update the status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, err

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

func GetService(ctx context.Context, client client.Client, namespace types.NamespacedName) (*corev1.Service, error) {

	service := &corev1.Service{}
	err := client.Get(ctx, namespace, service)
	return service, err
}
