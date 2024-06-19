package controller

import (
	"context"
	"fmt"

	ldhctlrv1alpha1 "github.com/babugeet/test-webserver-operator/api/v1alpha1"
	"github.com/babugeet/test-webserver-operator/variables"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

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

func GetDeployment(ctx context.Context, client client.Client, namespace types.NamespacedName) (*appsv1.Deployment, error) {
	fmt.Println("Namsepacedname")
	fmt.Println(namespace)
	deployment := &appsv1.Deployment{}
	err := client.Get(ctx, namespace, deployment)
	return deployment, err
}

func (r *WebserverReconciler) ReconcileDeployment(ctx context.Context, Webserver *ldhctlrv1alpha1.Webserver) (ctrl.Result, error) {

	// 2. Set the owner reference

	deployment := WebserverDeployment(Webserver)

	if err := controllerutil.SetControllerReference(Webserver, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// 3.  Get the current state of the deployment (replace with your logic)
	webDeploy, err := GetDeployment(ctx, r.Client, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace})
	// if err != nil && client.IgnoreNotFound(err) != nil {
	// 	log.Log.Error(err, "Failed to get deployment")
	// 	return ctrl.Result{}, err
	// }
	fmt.Println(err)
	if err != nil && client.IgnoreNotFound(err) == nil {
		fmt.Println("Creating deployment")
		variables.Found = false
		fmt.Println(err)
		if err := r.Client.Create(ctx, deployment); err != nil {
			fmt.Println(err.Error())
			return ctrl.Result{}, err
		}
	} else {
		variables.Found = true
	}
	//  Added logic to keep the replica count as required as part of the reconcile loop
	if variables.Found && *webDeploy.Spec.Replicas != *Webserver.Spec.Replica {
		fmt.Println("Desired replica count not matching to actual replicas")
		fmt.Println("Triggering update deployment")
		webDeploy.Spec.Replicas = Webserver.Spec.Replica
		err := r.Client.Update(ctx, deployment)
		if err != nil {
			return ctrl.Result{}, err
		}

	}
	//4. Updating status of Deployment
	Webserver.Status.Deployment = deployment.Name
	err = r.Status().Update(ctx, Webserver)
	fmt.Println("status updated")
	if err != nil {
		fmt.Println("Unable to update the status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, err

}
