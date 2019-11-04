---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_sql_managed_instance"
sidebar_current: "docs-azurerm-resource-database-sql-managed-instance"
description: |-
  Manages a SQL Managed Instance.

---

# azurerm_sql_managed_instance

Manages a SQL Managed Instance.

~> **Note:** All arguments including the administrator login and password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Example Usage

```hcl
data "azurerm_resource_group" "existing" {
  name = "networking"
}
data "azurerm_subnet" "existing" {
  name                 = "databases"
  virtual_network_name = "production-network"
  resource_group_name  = data.azurerm_resource_group.existing.name
}

resource "azurerm_sql_managed_instance" "example" {
  name                         = "example-sql-instance"
  resource_group_name          = data.azurerm_resource_group.existing.name
  location                     = data.azurerm_resource_group.existing.location
  subnet_id                    = data.azurerm_subnet.existing.id
  license_type                 = "BasePrice"
  administrator_login          = "mradministrator"
  administrator_login_password = "thisIsJpm81"

  tags {
    environment = "production"
  }
}
```
## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SQL Managed Instance, which must be globally unique within Azure. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the Resource Group where the SQL Managed Instance should exist. Changing this forces a new resource to be created.

* `location` - (Required) Specifies the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `license_type` - (Required) License of the Managed Instance. Possible values are `BasePrice` and `PriceIncluded`.

* `sku_name` - (Required) The SKU which should be used for this Managed Instance. Possible values are `BC_Gen4` and `BC_Gen5` (for a `Business Critical` SKU) `GP_Gen4` and `GP_Gen5` (for a `General Purpose` SKU).

* `storage_size_in_gb` - (Required) The amount of Storage which should be assigned to this SQL Managed Instance. This can be between 32GB and 8192GB in increments of 32GB.

* `subnet_id` - (Required) The ID of the Subnet which should be used for the SQL Managed Instance.

* `vcores` - (Required) The number of cores which should be assigned to this SQL Managed Instance. The number of cores available depends on the `sku_name` being used.

-> **NOTE:** When using a `Gen4` SKU - `vcores` can be `4`, `8`, `16` or `24`. When using a `Gen5` SKU - `vcores` can be `4`, `8`, `16`, `24`, `32`, `40`, `64` or `80`.

* `administrator_login` - (Required) The username which should be used as the Administrator for the SQL Managed Instance. Changing this forces a new resource to be created.

* `administrator_login_password` - (Required) The password associated with the `administrator_login`, which must comply with Azure's [Password Requirements](https://msdn.microsoft.com/library/ms161959.aspx)

---

* `collation` - (Optional) The collation which should be used for the SQL Managed Instance.

* `proxy_override` - (Optional) The connection type used for connectivity to the SQL Managed Instance. Possible values are `Proxy` and `Redirect`. Defaults to `Proxy`.

* `public_data_endpoint_enabled` - (Optional) Is the Public Data Endpoint enabled? Defaults to `false`.

* `tags` - (Optional) A mapping of tags to assign to the resource.

* `time_zone` - (Optional) The Time Zone which should be used for the SQL Managed Instance, [the possible values are defined here](https://jackstromberg.com/2017/01/list-of-time-zones-consumed-by-azure/).

## Attributes Reference

The following attributes are exported:

* `id` - The SQL Managed Instance ID.

* `fully_qualified_domain_name` - The fully qualified domain name of the Azure SQL Server

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 6 hours) Used when creating the SQL Managed Instance.
* `update` - (Defaults to 6 hours) Used when updating the SQL Managed Instance.
* `delete` - (Defaults to 6 hours) Used when deleting the SQL Managed Instance.

## Import

SQL Managed Instances can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_sql_managed_instance.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Sql/managedInstances/instance1
```
