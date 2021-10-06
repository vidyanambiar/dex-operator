/*
Copyright 2021.

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

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	authv1alpha1 "github.com/identitatem/dex-operator/api/v1alpha1"
	dexapi "github.com/identitatem/dex-operator/controllers/dex"
)

// DexClientReconciler reconciles a DexClient object
type DexClientReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=auth.identitatem.io,resources=dexclients,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=auth.identitatem.io,resources=dexclients/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=auth.identitatem.io,resources=dexclients/finalizers,verbs=update
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources={clusterroles},verbs=get;list;watch;create;update;patch;delete;escalate;bind
//+kubebuilder:rbac:groups="apiextensions.k8s.io",resources={customresourcedefinitions},verbs=get;list;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DexClient object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DexClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.V(1).Info("Reconciling...")

	dexv1Client := &authv1alpha1.DexClient{}
	if err := r.Get(ctx, req.NamespacedName, dexv1Client); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("found dexclient", "DexClient.name", dexv1Client.Name, "DexClient.namespace", dexv1Client.Namespace, "DexClient.status.state", dexv1Client.Status.State)

	// If dex server and dex client are created at the same time, we may need to wait a few seconds for dex server reconciler
	// to create the mtls certs
	mTLSSecret, err := r.getMTLSSecret(dexv1Client, ctx)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		} else {
			log.Error(err, "Error getting mTLS certificate to create api client connection to gRPC server", "client", dexv1Client.Name)
			dexv1Client.Status.State = authv1alpha1.PhaseFailed
			dexv1Client.Status.Message = err.Error()
			return ctrl.Result{Requeue: true}, err
		}
	}

	// Fetch the mTLS client cert and create the grpc client

	dexApiOptions := &dexapi.Options{
		HostAndPort: fmt.Sprintf("%s.%s.%s%s", GRPC_SERVICE_NAME, dexv1Client.Namespace, "svc.cluster.local", ":5557"),
		CABuffer:    bytes.NewBuffer(mTLSSecret.Data["ca.crt"]),
		CrtBuffer:   bytes.NewBuffer(mTLSSecret.Data["client.crt"]),
		KeyBuffer:   bytes.NewBuffer(mTLSSecret.Data["client.key"]),
	}
	dexApiClient, err := dexapi.NewClientPEM(dexApiOptions)
	if err != nil {
		log.Error(err, "Failed to create api client connection to gRPC server", "client", dexv1Client.Name)
		dexv1Client.Status.State = authv1alpha1.PhaseFailed
		dexv1Client.Status.Message = err.Error()
		return ctrl.Result{Requeue: true}, err
	}

	defer dexApiClient.CloseConnection()

	// if status is not set, set it to CREATING
	if dexv1Client.Status.State == "" || dexv1Client.Status.State == authv1alpha1.PhaseCreating {
		dexv1Client.Status.State = authv1alpha1.PhaseCreating
	}
	// Now let's make the main case distinction: implementing
	// the state diagram CREATING -> ACTIVE or CREATING -> FAILED
	switch dexv1Client.Status.State {
	case authv1alpha1.PhaseCreating:
		log.Info("Creating dex client", "name", dexv1Client.Name,
			"redirectURIs", dexv1Client.Spec.RedirectURIs,
			"TrustedPeers", dexv1Client.Spec.TrustedPeers,
			"Public", dexv1Client.Spec.Public,
			"ClientID", dexv1Client.Spec.ClientID,
			"LogoURL", dexv1Client.Spec.LogoURL,
			"clientSecretRef", dexv1Client.Spec.ClientSecretRef.Name)

		// read clientSecret from secret
		dexclientclientSecret, err := r.getClientClientSecretFromRef(dexv1Client, ctx)
		if err != nil {
			log.Error(err, "Client create failed on client secret", "client", dexv1Client.Name)
			return r.updateStatus(ctx, dexv1Client, authv1alpha1.PhaseFailed, err)
		}

		// Implement dex auth client creation here
		res, err := dexApiClient.CreateClient(
			ctx,
			dexv1Client.Spec.RedirectURIs,
			dexv1Client.Spec.TrustedPeers,
			dexv1Client.Spec.Public,
			dexv1Client.Name,
			dexv1Client.Spec.ClientID,
			dexv1Client.Spec.LogoURL,
			dexclientclientSecret,
		)
		if err != nil {
			log.Error(err, "Client create failed", "client", dexv1Client.Name)
			return r.updateStatus(ctx, dexv1Client, authv1alpha1.PhaseFailed, err)
		} else {
			dexv1Client.Status.State = authv1alpha1.PhaseActive
			log.Info("Client created", "client ID", res.GetId())
		}

	case authv1alpha1.PhaseActive:
		// If the client is active but in the reconcile loop it's being updated.
		log.Info("Client update", "client ID", dexv1Client.Name)
		err := dexApiClient.UpdateClient(
			ctx,
			dexv1Client.Spec.ClientID,
			dexv1Client.Spec.RedirectURIs,
			dexv1Client.Spec.TrustedPeers,
			dexv1Client.Spec.Public,
			dexv1Client.Name,
			dexv1Client.Spec.LogoURL,
		)
		if err != nil {
			log.Error(err, "Client update failed", "client", dexv1Client.Name)
			return r.updateStatus(ctx, dexv1Client, authv1alpha1.PhaseActiveDegraded, err)
		} else {
			log.Info("Client updated", "client ID", dexv1Client.Name)
		}

	case authv1alpha1.PhaseFailed:
		log.Info("Client failed")

	default:
		// Should never reach here
		log.Info("Got an invalid state", "state", dexv1Client.Status.State)
		return ctrl.Result{}, nil
	}

	err = r.Status().Update(ctx, dexv1Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Update the object and return
	err = r.Update(ctx, dexv1Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DexClientReconciler) updateStatus(ctx context.Context, dexClient *authv1alpha1.DexClient, status string, inErr error) (ctrl.Result, error) {
	dexClient.Status.State = status
	dexClient.Status.Message = inErr.Error()
	err := r.Status().Update(ctx, dexClient)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, inErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *DexClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1alpha1.DexClient{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func (r *DexClientReconciler) getMTLSSecret(m *authv1alpha1.DexClient, ctx context.Context) (*corev1.Secret, error) {
	// each dexserver will run in its own namespace
	// the dex controller will connect to mulitple dexservers
	// given a DexClient, the MTLS secret will be in the same namespace
	// we can find this secret by convention name
	secretNamespace := m.Namespace

	resource := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: SECRET_MTLS_NAME, Namespace: secretNamespace}, resource); err != nil {
		// failed to find the secret, wait for the secret to exist
		return nil, err
	}
	// secret exists, continue reading MTLS and connect to GRPC
	return resource, nil
}

func (r *DexClientReconciler) getClientClientSecretFromRef(m *authv1alpha1.DexClient, ctx context.Context) (string, error) {
	log := ctrllog.FromContext(ctx)
	secretName := m.Spec.ClientSecretRef.Name
	secretNamespace := m.Spec.ClientSecretRef.Namespace
	log.Info("getClientClientSecretFromRef", "secretName", secretName, "secretNamespace", "secretNamespace")

	resource := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, resource); err != nil {
		return "", err
	}

	log.Info("retrieve clientSecret in ", "secretName", secretName, "secretNamespace", "secretNamespace")
	if secret, ok := resource.Data["clientSecret"]; ok {
		log.Info("found clientSecret in ", "secretName", secretName, "secretNamespace", "secretNamespace")
		return string(secret), nil
	}
	return "", fmt.Errorf("secret %s/%s doesn't contain the data clientSecret", secretNamespace, secretName)
}
