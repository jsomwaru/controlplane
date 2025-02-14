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
	kbatch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	waitTime    int = 5
	jobOwnerKey     = ".metadata.controller"
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

func (r *TerraformRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	tfreq := &controlplaneiov1.TerraformRequest{}
	if err := r.Get(ctx, req.NamespacedName, tfreq); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(waitTime)}, nil
	}

	log.Info("tfreq status", "status", tfreq.Status.Applied)

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
	// Check if job is still running
	var childJobs kbatch.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		log.Error(err, "Not able to find jobs")
	}

	for _, job := range childJobs.Items {
		if job.Status.Failed >= 1 {
			tfreq.Status.Applied = "Failed"
		}
		if job.Status.Active >= 1 {
			continue
		}
		if job.Status.Succeeded >= 1 {
			tfreq.Status.Applied = "Applied"
		}
	}

	if err := r.Status().Update(ctx, tfreq); err != nil {
		log.Error(err, "Unable to update TerraformRequest Status")
		return ctrl.Result{}, err
	}

	if tfreq.Status.Applied == "Running" {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TerraformRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kbatch.Job{}, jobOwnerKey, func(rawObj client.Object) []string {
		job := rawObj.(*kbatch.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != controlplaneiov1.GroupVersion.String() || owner.Kind != "TerraformRequest" {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplaneiov1.TerraformRequest{}).
		Named("terraformrequest").
		Complete(r)
}
