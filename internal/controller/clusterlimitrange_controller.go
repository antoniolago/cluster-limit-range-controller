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
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
    clr := &lag0v1.ClusterLimitRange{}
    if err := r.Get(ctx, req.NamespacedName, clr); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Fetch all namespaces
    namespaces := &corev1.NamespaceList{}
    if err := r.List(ctx, namespaces); err != nil {
        return ctrl.Result{}, err
    }

    // Loop through namespaces
    for _, ns := range namespaces.Items {
        namespaceName := ns.Name

        // Skip ignored namespaces
        if isNamespaceIgnored(namespaceName, clr.Spec.IgnoredNamespaces) {
            log.Info(fmt.Sprintf("Namespace %s is in the ignored list, skipping", namespaceName))
            continue
        }

        // If applyNamespaces is not empty, check if the namespace is in the list
        if len(clr.Spec.ApplyNamespaces) > 0 && !isNamespaceInList(namespaceName, clr.Spec.ApplyNamespaces) {
            log.Info(fmt.Sprintf("Namespace %s is not in the applyNamespaces list, skipping", namespaceName))
            continue
        }

        // Check if the namespace already has a LimitRange
        limitRangeList := &corev1.LimitRangeList{}
        if err := r.List(ctx, limitRangeList, &client.ListOptions{Namespace: namespaceName}); err != nil {
            return ctrl.Result{}, fmt.Errorf("error fetching LimitRanges for namespace %s: %v", namespaceName, err)
        }

        if len(limitRangeList.Items) > 0 {
            // Skip if there's already a LimitRange in this namespace
            log.Info(fmt.Sprintf("Namespace %s already has a LimitRange, skipping", namespaceName))
            continue
        }

        // No LimitRange exists, apply the new one
        limitRange := &corev1.LimitRange{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "cluster-limit-range",
                Namespace: namespaceName,
            },
            Spec: corev1.LimitRangeSpec{
                Limits: translateClusterLimits(clr.Spec.Limits),
            },
        }

        if err := r.Create(ctx, limitRange); err != nil {
            return ctrl.Result{}, fmt.Errorf("error creating LimitRange in namespace %s: %v", namespaceName, err)
        }
        log.Info(fmt.Sprintf("Applied LimitRange to namespace %s", namespaceName))
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
