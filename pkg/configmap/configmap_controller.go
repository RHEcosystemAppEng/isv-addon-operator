/*
Copyright 2022.

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

package configmap

import (
	"context"
	"fmt"
	"github.com/isv-addon-operator/api/v1alpha1"
	"github.com/isv-addon-operator/pkg/isv"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ConfigMapReconciler reconciles a ConfigMap object
type ConfigMapReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",namespace=system,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",namespace=system,resources=configmaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",namespace=system,resources=configmaps/finalizers,verbs=update

func (r *ConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if err := r.deleteISVAddonCr(ctx, req.Namespace); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete ISVAddon CR: %w", err)
	}

	logger.Info("Successfully deleted ISVAddon CR")
	return ctrl.Result{}, nil
}

func (r *ConfigMapReconciler) deleteISVAddonCr(ctx context.Context, namespace string) error {
	logger := log.FromContext(ctx).WithValues("Reconcile Step", "DeleteISVAddonCr")
	logger.Info("Getting ISVAddon CR")

	isvAddonCrs := v1alpha1.ISVAddonList{}
	if err := r.List(ctx, &isvAddonCrs); err != nil {
		return err
	}

	if len(isvAddonCrs.Items) > 1 {
		logger.Info(fmt.Sprintf("In namespace %s there are multiple (%v) ISVAddon CRs.", namespace, len(isvAddonCrs.Items)))
	}

	for _, addonCr := range isvAddonCrs.Items {
		err := r.Delete(ctx, &addonCr)
		if err != nil || !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ConfigMap{}).
		WithEventFilter(configMapFilter()).
		Complete(r)
}

func configMapFilter() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isAddonConfigMap(e.ObjectNew)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isAddonConfigMap(e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

func isAddonConfigMap(object client.Object) bool {
	return object.GetName() == isv.CommonISVInstance.GetAddonID() && object.GetNamespace() == isv.CommonISVInstance.GetAddonCRNamespace()
}
