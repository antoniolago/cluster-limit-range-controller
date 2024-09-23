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
    "k8s.io/apimachinery/pkg/api/errors"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	lag0v1 "lag0.com.br/cluster-limit-range-controller/api/v1"
)

// ClusterLimitRangeReconciler reconciles a ClusterLimitRange object
type ClusterLimitRangeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=lag0.com.br,resources=clusterlimitranges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=lag0.com.br,resources=clusterlimitranges/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=lag0.com.br,resources=clusterlimitranges/finalizers,verbs=update

// Reconcile function
func (r *ClusterLimitRangeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // Fetch the ClusterLimitRange instance
    var clr lag0v1.ClusterLimitRange
    if err := r.Get(ctx, req.NamespacedName, &clr); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil // Resource not found, ignore
        }
        return ctrl.Result{}, err
    }

    // Create sets for ignoredNamespaces and applyNamespaces
    ignoredNamespaces := sets.NewString(clr.Spec.IgnoredNamespaces...)
    applyNamespaces := sets.NewString(clr.Spec.ApplyNamespaces...)

    // List all namespaces
    var namespaces corev1.NamespaceList
    if err := r.List(ctx, &namespaces); err != nil {
        log.Error(err, "unable to list namespaces")
        return ctrl.Result{}, err
    }

    // Map to keep track of applied LimitRanges
    existingLimitRanges := make(map[string]corev1.LimitRange)

    // Iterate through namespaces and handle LimitRanges
    for _, ns := range namespaces.Items {
        namespaceName := ns.Name

        // Skip namespaces that are in the ignoredNamespaces list
        if ignoredNamespaces.Has(namespaceName) {
            log.Info("Skipping namespace in ignoredNamespaces", "namespace", namespaceName)
            continue
        }

        // If applyNamespaces is specified, only apply to these namespaces
        if len(applyNamespaces) > 0 && !applyNamespaces.Has(namespaceName) {
            log.Info("Skipping namespace not in applyNamespaces", "namespace", namespaceName)
            continue
        }

        // Check if LimitRange already exists in this namespace
        var limitRangeList corev1.LimitRangeList
        if err := r.List(ctx, &limitRangeList, client.InNamespace(namespaceName)); err != nil {
            log.Error(err, "unable to list LimitRanges for namespace", "namespace", namespaceName)
            return ctrl.Result{}, err
        }

        // Handle LimitRange creation
        if len(limitRangeList.Items) == 0 {
            limitRange := &corev1.LimitRange{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "default-limits",
                    Namespace: namespaceName,
                },
				Spec: corev1.LimitRangeSpec{
					Limits: translateClusterLimits(clr.Spec.Limits),
				},
            }

            if err := r.Create(ctx, limitRange); err != nil {
                log.Error(err, "unable to create LimitRange for namespace", "namespace", namespaceName)
                return ctrl.Result{}, err
            }

            log.Info("Created LimitRange for namespace", "namespace", namespaceName)
        } else {
            existingLimitRanges[namespaceName] = limitRangeList.Items[0] // Keep track of existing LimitRanges
        }
    }

    // Handle deletions of LimitRange objects for namespaces no longer in applyNamespaces
    for _, ns := range existingLimitRanges {
        namespaceName := ns.Namespace
        if !applyNamespaces.Has(namespaceName) {
            log.Info("Deleting LimitRange for namespace no longer in applyNamespaces", "namespace", namespaceName)
            if err := r.Delete(ctx, &ns); err != nil {
                log.Error(err, "unable to delete LimitRange for namespace", "namespace", namespaceName)
                return ctrl.Result{}, err
            }
        }
    }

    return ctrl.Result{}, nil
}

// isNamespaceIgnored checks if a namespace is in the ignored list
func isNamespaceIgnored(namespace string, ignoredNamespaces []string) bool {
    for _, ignored := range ignoredNamespaces {
        if namespace == ignored {
            return true
        }
    }
    return false
}

// isNamespaceInList checks if a namespace is in the apply list
func isNamespaceInList(namespace string, applyNamespaces []string) bool {
    for _, apply := range applyNamespaces {
        if namespace == apply {
            return true
        }
    }
    return false
}

// translateClusterLimits translates the ClusterLimitRange spec into a LimitRange spec
func translateClusterLimits(clusterLimits []lag0v1.LimitRangeItem) []corev1.LimitRangeItem {
    var limitRangeItems []corev1.LimitRangeItem
    for _, cl := range clusterLimits {
        limitRangeItems = append(limitRangeItems, corev1.LimitRangeItem{
            Type:               corev1.LimitType(cl.Type),
            Max:                cl.Max,
            Min:                cl.Min,
            Default:            cl.Default,
            DefaultRequest:     cl.DefaultRequest,
            MaxLimitRequestRatio: cl.MaxLimitRequestRatio,
        })
    }
    return limitRangeItems
}

// SetupWithManager sets up the controller with the Manager
func (r *ClusterLimitRangeReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&lag0v1.ClusterLimitRange{}).
        Complete(r)
}

