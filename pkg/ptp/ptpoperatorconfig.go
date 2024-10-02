package ptp

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	ptpv1 "github.com/openshift/ptp-operator/api/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// PtpOperatorConfigBuilder provides a struct for the PtpOperatorConfig resource containing a connection to the cluster
// and the PtpOperatorConfig definition.
type PtpOperatorConfigBuilder struct {
	// Definition of the PtpOperatorConfig used to create the object.
	Definition *ptpv1.PtpOperatorConfig
	// Object of the PtpOperatorConfig as it is on the cluster.
	Object    *ptpv1.PtpOperatorConfig
	apiClient goclient.Client
	errorMsg  string
}

// PullPtpOperatorConfig pulls an existing PtpOperatorConfig into a Builder struct.
func PullPtpOperatorConfig(apiClient *clients.Settings, name, nsname string) (*PtpOperatorConfigBuilder, error) {
	glog.V(100).Infof("Pulling existing PtpOperatorConfig %s under namespace %s from cluster", name, nsname)

	if apiClient == nil {
		glog.V(100).Info("The apiClient is empty")

		return nil, fmt.Errorf("ptpOperatorConfig 'apiClient' cannot be nil")
	}

	err := apiClient.AttachScheme(ptpv1.AddToScheme)
	if err != nil {
		glog.V(100).Info("Failed to add ptp v1 scheme to client schemes")

		return nil, err
	}

	builder := &PtpOperatorConfigBuilder{
		apiClient: apiClient.Client,
		Definition: &ptpv1.PtpOperatorConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Info("The name of the PtpOperatorConfig is empty")

		return nil, fmt.Errorf("ptpOperatorConfig 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Info("The namespace of the PtpOperatorConfig is empty")

		return nil, fmt.Errorf("ptpOperatorConfig 'nsname' cannot be empty")
	}

	if !builder.Exists() {
		glog.V(100).Info("The PtpOperatorConfig %s does not exist in namespace %s", name, nsname)

		return nil, fmt.Errorf("ptpOperatorConfig object %s does not exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return builder, nil
}

// Get returns the PtpOperatorConfig object if found.
func (builder *PtpOperatorConfigBuilder) Get() (*ptpv1.PtpOperatorConfig, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Getting PtpOperatorConfig object %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	ptpOperatorConfig := &ptpv1.PtpOperatorConfig{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, ptpOperatorConfig)

	if err != nil {
		glog.V(100).Infof(
			"PtpOperatorConfig object %s does not exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return ptpOperatorConfig, nil
}

// Exists checks whether the given PtpOperatorConfig exists on the cluster.
func (builder *PtpOperatorConfigBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if PtpOperatorConfig %s exists in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes a PtpOperatorConfig on the cluster if it does not already exist.
func (builder *PtpOperatorConfigBuilder) Create() (*PtpOperatorConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Creating PtpOperatorConfig %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	if builder.Exists() {
		return builder, nil
	}

	err := builder.apiClient.Create(context.TODO(), builder.Definition)
	if err != nil {
		return nil, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Update changes the existing PtpOperatorConfig resource on the cluster, falling back to deleting and recreating if the
// update fails when force is set.
func (builder *PtpOperatorConfigBuilder) Update(force bool) (*PtpOperatorConfigBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Updating PtpOperatorConfig %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof(
			"PtpOperatorConfig %s does not exist in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

		return nil, fmt.Errorf("cannot update non-existent ptpOperatorConfig")
	}

	builder.Definition.ResourceVersion = builder.Object.ResourceVersion

	err := builder.apiClient.Update(context.TODO(), builder.Definition)
	if err != nil {
		if force {
			glog.V(100).Infof(msg.FailToUpdateNotification("ptpOperatorConfig", builder.Definition.Name))

			err := builder.Delete()
			builder.Definition.ResourceVersion = ""

			if err != nil {
				glog.V(100).Infof(msg.FailToUpdateError("ptpOperatorConfig", builder.Definition.Name))

				return nil, err
			}

			return builder.Create()
		}

		return nil, err
	}

	builder.Object = builder.Definition

	return builder, nil
}

// Delete removes a PtpOperatorConfig from the cluster if it exists.
func (builder *PtpOperatorConfigBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof(
		"Deleting PtpOperatorConfig %s in namespace %s", builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		glog.V(100).Infof(
			"PtpOperatorConfig %s in namespace %s does not exist",
			builder.Definition.Name, builder.Definition.Namespace)

		builder.Object = nil

		return nil
	}

	err := builder.apiClient.Delete(context.TODO(), builder.Object)
	if err != nil {
		return err
	}

	builder.Object = nil

	return nil
}

// validate checks that the builder, definition, and apiClient are properly initialized and there is no errorMsg.
func (builder *PtpOperatorConfigBuilder) validate() (bool, error) {
	resourceCRD := "ptpOperatorConfig"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is uninitialized", resourceCRD)

		return false, fmt.Errorf(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiClient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
