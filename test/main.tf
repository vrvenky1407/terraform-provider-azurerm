resource "azurerm_resource_group" "tfcloud-test" {
  name     = var.resource_group_name
  location = var.location

  tags ={
    environment = "codelab"
  }
}

