package v2vvmware

import (
	"context"
	"errors"
	"strings"

	corev1 "k8s.io/api/core/v1"

	kubevirtv1alpha1 "github.com/ovirt/v2v-conversion-host/kubevirt-vmware/pkg/apis/kubevirt/v1alpha1"
	"github.com/ovirt/v2v-conversion-host/kubevirt-vmware/pkg/controller/utils"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func getConnectionSecret(r *ReconcileV2VVmware, request reconcile.Request, instance *kubevirtv1alpha1.V2VVmware) (*corev1.Secret, error) {
	if instance.Spec.Connection == "" {
		return nil, errors.New("the Spec.Connection is required in a V2VVmware object. References a Secret by name")
	}

	secret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Connection, Namespace: request.Namespace}, secret)
	return secret, err
}

func getLoginCredentials(connectionSecret *corev1.Secret) *LoginCredentials {
	data := connectionSecret.Data

	credentials := &LoginCredentials{
		username: strings.TrimSpace(string(data["username"])),
		password: strings.TrimSpace(string(data["password"])),
		host:     strings.TrimSpace(string(data["url"])),
	}

	log.Info("VMWare credentials retrieved from a Secret", credentials.username, credentials.host)
	return credentials
}

// read whole list at once
func readVmsList(r *ReconcileV2VVmware, request reconcile.Request, connectionSecret *corev1.Secret, provider string) error {
	log.Info("Getting list of vms")

	updateStatusPhase(r, request, PhaseConnecting)
	client, err := getClient(context.Background(), getLoginCredentials(connectionSecret))
	if err != nil {
		updateStatusPhase(r, request, PhaseConnectionFailed)
		log.Error(err, "Faild to get client")
		return err
	}
	defer client.Logout()

	updateStatusPhase(r, request, PhaseLoadingVmsList)
	vms, err := GetVMs(client, provider, request.Namespace)
	if err != nil {
		updateStatusPhase(r, request, PhaseLoadingVmsListFailed)
		log.Error(err, "Faild to get vms")
		return err
	}

	instance, err := getInstance(r, request)
	for _, vm := range vms {
		log.Info("Create vm", "vm", vm)
		if err := controllerutil.SetControllerReference(instance, &vm, r.scheme); err != nil {
			log.Error(err, "Faild to set controller ref")
			return err
		}
		// Check if this Vm already exists
		found := &kubevirtv1alpha1.ExternalVm{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: vm.Name, Namespace: vm.Namespace}, found)
		if err != nil && apierrors.IsNotFound(err) {
			log.Info("Creating a new VM", "VM.Namespace", vm.Namespace, "VM.Name", vm.Name)
			err = r.client.Create(context.TODO(), &vm)
			if err != nil {
				// TODO check whether we can skip and continue with other vms
				updateStatusPhase(r, request, PhaseLoadingVmsListFailed)
				log.Error(err, "Faild to create vm", "Vm details", vm)
				return err
			}
		} else if err != nil {
			// TODO check whether we can skip and continue with other vms
			updateStatusPhase(r, request, PhaseLoadingVmsListFailed)
			log.Error(err, "Faild to get the vm")
			return err
		}
	}

	updateStatusPhase(r, request, PhaseConnectionSuccessful)
	return nil
}

func getInstance(r *ReconcileV2VVmware, request reconcile.Request) (*kubevirtv1alpha1.V2VVmware, error) {
	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Error(err, "Failed to get V2VVmware object")
		return nil, err
	}
	return instance, nil
}

func updateStatusPhase(r *ReconcileV2VVmware, request reconcile.Request, phase string) {
	log.Info("Updating provider with", "status phase", phase)
	updateStatusPhaseRetry(r, request, phase, utils.MaxRetryCount)
}

func updateStatusPhaseRetry(r *ReconcileV2VVmware, request reconcile.Request, phase string, retryCount int) {
	// reload instance to workaround issues with parallel writes
	// TODO those updates expose race condition. Once we introduce ExternalVm we would reduce the need for it.
	instance := &kubevirtv1alpha1.V2VVmware{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		log.Error(err, "Failed to get V2VVmware object to update status info", phase)
		if retryCount > 0 {
			utils.SleepBeforeRetry()
			updateStatusPhaseRetry(r, request, phase, retryCount-1)
		}
		return
	}

	instance.Status.Phase = phase
	err = r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, "Failed to update V2VVmware status", phase)
		if retryCount > 0 {
			utils.SleepBeforeRetry()
			updateStatusPhaseRetry(r, request, phase, retryCount-1)
		}
	}
}
