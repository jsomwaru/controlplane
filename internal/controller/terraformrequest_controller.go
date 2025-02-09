/*
Copyright 2025.

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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlplaneiov1 "controlplane.io/terraformrequest/api/v1"
	tfexec "controlplane.io/terraformrequest/pkg/tfexec"
)

const (
	waitTime int = 5
)

// TerraformRequestReconciler reconciles a TerraformRequest object
type TerraformRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=controlplane.io.controlplane.io,resources=terraformrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmap,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controlplane.io.controlplane.io,resources=terraformrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controlplane.io.controlplane.io,resources=terraformrequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TerraformRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *TerraformRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	tfreq := &controlplaneiov1.TerraformRequest{}
	if err := r.Get(ctx, req.NamespacedName, tfreq); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second * time.Duration(waitTime)}, nil
	}

	if tfreq.Status.Applied == "" {
		job, cm := tfexec.CreateTerraformApplyJob(tfreq)
		if err := ctrl.SetControllerReference(tfreq, cm, r.Scheme); err != nil {
			log.Error(err, "Error setting ConfigMap Reference")
			return ctrl.Result{Requeue: true, RequeueAfter: time.Second * time.Duration(waitTime)}, nil
		}
		if err := ctrl.SetControllerReference(tfreq, job, r.Scheme); err != nil {
			log.Error(err, "Error setting Job Reference")
			return ctrl.Result{Requeue: true, RequeueAfter: time.Second * time.Duration(waitTime)}, nil
		}
		if err := r.Create(ctx, cm); err != nil {
			log.Error(err, "Error creating ConfigMap")
			return ctrl.Result{}, nil
		}
		if err := r.Create(ctx, job); err != nil {
			log.Error(err, "Error creating ConfigMap")
			return ctrl.Result{}, nil
		}
		tfreq.Status.Applied = "Running"
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TerraformRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplaneiov1.TerraformRequest{}).
		Named("terraformrequest").
		Complete(r)
}
