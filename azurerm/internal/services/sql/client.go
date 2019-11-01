package sql

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/2017-03-01-preview/sql"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	DatabasesClient                       *sql.DatabasesClient
	DatabaseThreatDetectionPoliciesClient *sql.DatabaseThreatDetectionPoliciesClient
	ElasticPoolsClient                    *sql.ElasticPoolsClient
	FirewallRulesClient                   *sql.FirewallRulesClient
	FailoverGroupsClient                  *sql.FailoverGroupsClient
	ManagedInstancesClient                *sql.ManagedInstancesClient
	ServersClient                         *sql.ServersClient
	ServerAzureADAdministratorsClient     *sql.ServerAzureADAdministratorsClient
	VirtualNetworkRulesClient             *sql.VirtualNetworkRulesClient
}

func BuildClient(o *common.ClientOptions) *Client {
	// SQL Azure
	databasesClient := sql.NewDatabasesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&databasesClient.Client, o.ResourceManagerAuthorizer)

	databaseThreatDetectionPoliciesClient := sql.NewDatabaseThreatDetectionPoliciesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&databaseThreatDetectionPoliciesClient.Client, o.ResourceManagerAuthorizer)

	elasticPoolsClient := sql.NewElasticPoolsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&elasticPoolsClient.Client, o.ResourceManagerAuthorizer)

	failoverGroupsClient := sql.NewFailoverGroupsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&failoverGroupsClient.Client, o.ResourceManagerAuthorizer)

	firewallRulesClient := sql.NewFirewallRulesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&firewallRulesClient.Client, o.ResourceManagerAuthorizer)

	managedInstancesClient := sql.NewManagedInstancesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&managedInstancesClient.Client, o.ResourceManagerAuthorizer)

	serversClient := sql.NewServersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&serversClient.Client, o.ResourceManagerAuthorizer)

	serverAzureADAdministratorsClient := sql.NewServerAzureADAdministratorsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&serverAzureADAdministratorsClient.Client, o.ResourceManagerAuthorizer)

	virtualNetworkRulesClient := sql.NewVirtualNetworkRulesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&virtualNetworkRulesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		DatabasesClient:                       &databasesClient,
		DatabaseThreatDetectionPoliciesClient: &databaseThreatDetectionPoliciesClient,
		ElasticPoolsClient:                    &elasticPoolsClient,
		FailoverGroupsClient:                  &failoverGroupsClient,
		FirewallRulesClient:                   &firewallRulesClient,
		ManagedInstancesClient:                &managedInstancesClient,
		ServersClient:                         &serversClient,
		ServerAzureADAdministratorsClient:     &serverAzureADAdministratorsClient,
		VirtualNetworkRulesClient:             &virtualNetworkRulesClient,
	}
}
