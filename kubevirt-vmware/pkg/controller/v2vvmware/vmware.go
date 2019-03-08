package v2vvmware

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	kubevirtv1alpha1 "github.com/ovirt/v2v-conversion-host/kubevirt-vmware/pkg/apis/kubevirt/v1alpha1"
	"github.com/vmware/govmomi/vim25/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
  Following code is based on https://github.com/pkliczewski/provider-pod
  modified for the needs of the controller-flow.
*/

func getClient(ctx context.Context, loginCredentials *LoginCredentials) (*Client, error) {
	c, err := NewClient(ctx, loginCredentials)
	if err != nil {
		log.Error(err, "Client creation failed.")
		return nil, err
	}
	return c, nil
}

func GetVMs(c *Client, provider string, namespace string) ([]kubevirtv1alpha1.ExternalVm, error) {
	vms, err := c.GetVMs()
	if err != nil {
		log.Error(err, "Getting VMs failed.")
		return nil, err
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Error(err, "Faild to compile regexp")
		return nil, err
	}

	eVms := make([]kubevirtv1alpha1.ExternalVm, len(vms))
	for index, vm := range vms {
		labels := map[string]string{
			"provider": provider,
		}

		var disks []kubevirtv1alpha1.ExternalDisk
		var nics []kubevirtv1alpha1.ExternalNic
		for _, device := range vm.Config.Hardware.Device {
			switch device := device.(type) {
			case *types.VirtualDisk:
				disk := kubevirtv1alpha1.ExternalDisk{
					Label:    device.GetVirtualDevice().DeviceInfo.GetDescription().Label,
					Filename: device.GetVirtualDevice().Backing.(types.BaseVirtualDeviceFileBackingInfo).GetVirtualDeviceFileBackingInfo().FileName,
					Capacity: device.CapacityInKB,
				}
				disks = append(disks, disk)
			case *types.VirtualVmxnet3:
				nic := kubevirtv1alpha1.ExternalNic{
					Label: device.GetVirtualDevice().DeviceInfo.GetDescription().Label,
					Mac:   device.MacAddress,
				}
				nics = append(nics, nic)
			}
		}
		name := strings.ToLower(vm.Summary.Config.Name)

		eVms[index] = kubevirtv1alpha1.ExternalVm{
			ObjectMeta: metav1.ObjectMeta{
				Name:      reg.ReplaceAllString(name, ""), // TODO make sure it is unique due to vmware allowing the same name
				Namespace: namespace,
				Labels:    labels,
			},
			Spec: kubevirtv1alpha1.ExternalVmSpec{
				VmName:        vm.Summary.Config.Name,
				Memory:        vm.Config.Hardware.MemoryMB,
				CPUs:          vm.Config.Hardware.NumCPU,
				CpuSockets:    vm.Config.Hardware.NumCoresPerSocket,
				GuestFullName: vm.Config.GuestFullName,
				PowerState:    string(vm.Summary.Runtime.PowerState),
				Disks:         disks,
				Nics:          nics,
			},
		}
	}

	log.Info("Retrieved list of virtual machines")
	return eVms, nil
}

func GetVM(c *Client, vmName string) (string, error) {
	vm, err := c.GetVM(vmName)
	if err != nil {
		log.Error(err, fmt.Sprintf("GetVM: failed to get details of VMWare VM '%s'", vmName))
		return "", err
	}

	raw, _ := json.Marshal(vm)

	return string(raw), nil
}
