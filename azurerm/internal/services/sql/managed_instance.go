package sql

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type ManagedInstanceResourceID struct {
	Name          string
	ResourceGroup string

	ID azure.ResourceID
}

func ParseManagedInstanceResourceID(id string) (*ManagedInstanceResourceID, error) {
	parsedId, err := azure.ParseAzureResourceID(id)
	if err != nil {
		return nil, err
	}

	resourceGroup := parsedId.ResourceGroup
	if resourceGroup == "" {
		return nil, fmt.Errorf("%q is missing a Resource Group", id)
	}

	instanceName := parsedId.Path["managedInstances"]
	if instanceName == "" {
		return nil, fmt.Errorf("%q is missing the `managedInstances` segment", id)
	}

	resourceID := ManagedInstanceResourceID{
		Name:          instanceName,
		ResourceGroup: resourceGroup,
		ID:            *parsedId,
	}
	return &resourceID, nil
}
