package mssql

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/2017-10-01-preview/sql"
	sqlmi "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/2018-06-01-preview/sql"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	ElasticPoolsClient     *sql.ElasticPoolsClient
	ManagedInstancesClient *sqlmi.ManagedInstancesClient
}

func BuildClient(o *common.ClientOptions) *Client {
	elasticPoolsClient := sql.NewElasticPoolsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&elasticPoolsClient.Client, o.ResourceManagerAuthorizer)

	managedInstancesClient := sqlmi.NewManagedInstancesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&managedInstancesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		ElasticPoolsClient:     &elasticPoolsClient,
		ManagedInstancesClient: &managedInstancesClient,
	}
}
