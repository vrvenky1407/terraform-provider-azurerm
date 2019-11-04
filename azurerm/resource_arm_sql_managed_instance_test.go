package azurerm

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

// TODO: a test scaling vcores, storage and then together
// TODO: a test updating the password, which presumably needs the public endpoint?

func TestAccAzureRMSQLManagedInstance_basic(t *testing.T) {
	resourceName := "azurerm_sql_managed_instance.test"
	ri := tf.AccRandTimeInt()
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMSQLManagedInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMSQLManagedInstance_basic(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"administrator_login_password"},
			},
		},
	})
}

func TestAccAzureRMSQLManagedInstance_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	resourceName := "azurerm_sql_managed_instance.test"
	ri := tf.AccRandTimeInt()
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMSQLManagedInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMSQLManagedInstance_basic(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				Config:      testAccAzureRMSQLManagedInstance_basic(ri, location),
				ExpectError: testRequiresImportError("azurerm_sql_managed_instance"),
			},
		},
	})
}

func TestAccAzureRMSQLManagedInstance_disappears(t *testing.T) {
	resourceName := "azurerm_sql_managed_instance.test"
	ri := tf.AccRandTimeInt()
	config := testAccAzureRMSQLManagedInstance_basic(ri, testLocation())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMSQLManagedInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
					testCheckAzureRMSQLManagedInstanceDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureRMSQLManagedInstance_complete(t *testing.T) {
	resourceName := "azurerm_sql_managed_instance.test"
	ri := tf.AccRandTimeInt()
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMSQLManagedInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMSQLManagedInstance_complete(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"administrator_login_password"},
			},
		},
	})
}

func TestAccAzureRMSQLManagedInstance_proxyOverride(t *testing.T) {
	resourceName := "azurerm_sql_managed_instance.test"
	ri := tf.AccRandTimeInt()
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMSQLManagedInstanceDestroy,
		Steps: []resource.TestStep{
			{
				// Redirect
				Config: testAccAzureRMSQLManagedInstance_proxyOverrideRedirect(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"administrator_login_password"},
			},
			{
				// Proxy
				Config: testAccAzureRMSQLManagedInstance_basic(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"administrator_login_password"},
			},
			{
				// Redirect
				Config: testAccAzureRMSQLManagedInstance_proxyOverrideRedirect(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"administrator_login_password"},
			},
		},
	})
}

func TestAccAzureRMSQLManagedInstance_publicDataEndpointEnabled(t *testing.T) {
	resourceName := "azurerm_sql_managed_instance.test"
	ri := tf.AccRandTimeInt()
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMSQLManagedInstanceDestroy,
		Steps: []resource.TestStep{
			{
				// Enabled
				Config: testAccAzureRMSQLManagedInstance_publicDataEndpointEnabled(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"administrator_login_password"},
			},
			{
				// Disabled
				Config: testAccAzureRMSQLManagedInstance_basic(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"administrator_login_password"},
			},
			{
				// Enabled
				Config: testAccAzureRMSQLManagedInstance_publicDataEndpointEnabled(ri, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSQLManagedInstanceExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"administrator_login_password"},
			},
		},
	})
}

func testCheckAzureRMSQLManagedInstanceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		sqlServerName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for SQL Managed Instance: %s", sqlServerName)
		}

		conn := testAccProvider.Meta().(*ArmClient).Mssql.ManagedInstancesClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := conn.Get(ctx, resourceGroup, sqlServerName)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: SQL Managed Instance %s (resource group: %s) does not exist", sqlServerName, resourceGroup)
			}
			return fmt.Errorf("Bad: Get SQL Managed Instance: %v", err)
		}

		return nil
	}
}

func testCheckAzureRMSQLManagedInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).Mssql.ManagedInstancesClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_sql_managed_instance" {
			continue
		}

		sqlServerName := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(ctx, resourceGroup, sqlServerName)

		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return fmt.Errorf("Bad: Get SQL Managed Instance: %+v", err)
		}

		return fmt.Errorf("SQL Managed Instance %s still exists", sqlServerName)

	}

	return nil
}

func testCheckAzureRMSQLManagedInstanceDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		serverName := rs.Primary.Attributes["name"]

		client := testAccProvider.Meta().(*ArmClient).Mssql.ManagedInstancesClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := client.Delete(ctx, resourceGroup, serverName)
		if err != nil {
			return fmt.Errorf("Error deleting: %+v", err)
		}

		if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("Error waiting for deletion: %+v", err)
		}

		return nil
	}
}

func testAccAzureRMSQLManagedInstance_basic(rInt int, location string) string {
	template := testAccAzureRMSQLManagedInstance_template(rInt, location)
	return fmt.Sprintf(`
%s
 
resource "azurerm_sql_managed_instance" "test" {
  name                         = "acctestsqlserver%d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  subnet_id					   = azurerm_subnet.test.id
  administrator_login          = "mradministrator"
  administrator_login_password = "P@s$w0rd1234!1234!"
  license_type				   = "BasePrice"
  storage_size_in_gb           = 32
  sku_name                     = "GP_Gen5"
  vcores                       = 4
}
`, template, rInt)
}

func testAccAzureRMSQLManagedInstance_complete(rInt int, location string) string {
	template := testAccAzureRMSQLManagedInstance_template(rInt, location)
	return fmt.Sprintf(`
%s
 
resource "azurerm_sql_managed_instance" "test" {
  name                         = "acctestsqlserver%d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  subnet_id					   = azurerm_subnet.test.id
  administrator_login          = "mradministrator"
  administrator_login_password = "P@s$w0rd1234!1234!"
  license_type				   = "BasePrice"
  storage_size_in_gb           = 32
  sku_name                     = "GP_Gen5"
  vcores                       = 4

  tags = {
	environment = "staging"
	database    = "test"
  }
}
`, template, rInt)
}

func testAccAzureRMSQLManagedInstance_proxyOverrideRedirect(rInt int, location string) string {
	template := testAccAzureRMSQLManagedInstance_template(rInt, location)
	return fmt.Sprintf(`
%s
 
resource "azurerm_sql_managed_instance" "test" {
  name                         = "acctestsqlserver%d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  subnet_id					   = azurerm_subnet.test.id
  administrator_login          = "mradministrator"
  administrator_login_password = "P@s$w0rd1234!1234!"
  license_type				   = "BasePrice"
  storage_size_in_gb           = 32
  sku_name                     = "GP_Gen5"
  vcores                       = 4
  proxy_override               = "Redirect"
}
`, template, rInt)
}

func testAccAzureRMSQLManagedInstance_publicDataEndpointEnabled(rInt int, location string) string {
	template := testAccAzureRMSQLManagedInstance_template(rInt, location)
	return fmt.Sprintf(`
%s
 
resource "azurerm_sql_managed_instance" "test" {
  name                         = "acctestsqlserver%d"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  subnet_id					   = azurerm_subnet.test.id
  administrator_login          = "mradministrator"
  administrator_login_password = "P@s$w0rd1234!1234!"
  license_type				   = "BasePrice"
  storage_size_in_gb           = 32
  sku_name                     = "GP_Gen5"
  vcores                       = 4
  public_data_endpoint_enabled = true
}
`, template, rInt)
}

func testAccAzureRMSQLManagedInstance_template(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctest-vnet-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  address_space       = ["10.0.0.0/16"]
}
  
resource "azurerm_subnet" "test" {
  name                 = "subnet-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.0.0/24"  
  route_table_id       = azurerm_route_table.test.id
}

resource "azurerm_route_table" "test" {
  name                          = "routetable-%d"
  location                      = azurerm_resource_group.test.location
  resource_group_name           = azurerm_resource_group.test.name
  disable_bgp_route_propagation = false

  route {
    name           = "RouteToAzureSqlMiMngSvc"
    address_prefix = "0.0.0.0/0"
    next_hop_type  = "Internet"
  }
}

resource "azurerm_subnet_route_table_association" "test" {
  subnet_id      = azurerm_subnet.test.id
  route_table_id = azurerm_route_table.test.id
}
`, rInt, location, rInt, rInt, rInt)
}
