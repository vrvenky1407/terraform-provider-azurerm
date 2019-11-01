package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/2017-03-01-preview/sql"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	sqlSvc "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/sql"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmSqlManagedInstance() *schema.Resource {
	return &schema.Resource{
		// TODO: split Create and Update
		Create: resourceArmSqlManagedInstanceCreateUpdate,
		Read:   resourceArmSqlManagedInstanceServerRead,
		Update: resourceArmSqlManagedInstanceCreateUpdate,
		Delete: resourceArmSqlManagedInstanceDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateMsSqlServerName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"GP_Gen4",
								"GP_Gen5",
							}, false),
						},

						"capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"tier": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"family": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"administrator_login": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// TODO: validation
			},

			"administrator_login_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				// TODO: validation
			},

			"vcores": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  8,
				ValidateFunc: validate.IntInSlice([]int{
					8,
					16,
					24,
					32,
					40,
					64,
					80,
				}),
			},

			"storage_size_in_gb": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      32,
				ValidateFunc: validate.IntBetweenAndDivisibleBy(32, 8000, 32),
			},

			"license_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(sql.LicenseIncluded),
				ValidateFunc: validation.StringInSlice([]string{
					string(sql.BasePrice),
					string(sql.LicenseIncluded),
				}, false),
			},

			"subnet_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmSqlManagedInstanceCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Sql.ManagedInstancesClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))
	adminUsername := d.Get("administrator_login").(string)
	licenseType := d.Get("license_type").(string)
	subnetId := d.Get("subnet_id").(string)
	t := d.Get("tags").(map[string]interface{})

	// TODO: lock on the subnet id
	// TODO: requires import

	parameters := sql.ManagedInstance{
		Location: utils.String(location),
		Tags:     tags.Expand(t),
		ManagedInstanceProperties: &sql.ManagedInstanceProperties{
			LicenseType:        sql.ManagedInstanceLicenseType(licenseType),
			AdministratorLogin: utils.String(adminUsername),
			SubnetID:           utils.String(subnetId),
		},
	}

	if d.HasChange("administrator_login_password") {
		adminPassword := d.Get("administrator_login_password").(string)
		parameters.ManagedInstanceProperties.AdministratorLoginPassword = utils.String(adminPassword)
	}

	future, err := client.CreateOrUpdate(ctx, resGroup, name, parameters)
	if err != nil {
		return err
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {

		if response.WasConflict(future.Response()) {
			return fmt.Errorf("SQL Server names need to be globally unique and %q is already in use.", name)
		}

		return err
	}

	resp, err := client.Get(ctx, resGroup, name)
	if err != nil {
		return err
	}

	d.SetId(*resp.ID)

	return resourceArmSqlManagedInstanceServerRead(d, meta)
}

func resourceArmSqlManagedInstanceServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Sql.ManagedInstancesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := sqlSvc.ParseManagedInstanceResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	name := id.Name

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] SQL Managed Instance %q was not found in Resource Group %q - assuming removed & removing from state", name, resourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving SQL Managed Instance %q (Resource Group %q): %v", name, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.ManagedInstanceProperties; props != nil {
		d.Set("license_type", string(props.LicenseType))
		d.Set("administrator_login", props.AdministratorLogin)
		d.Set("fully_qualified_domain_name", props.FullyQualifiedDomainName)
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmSqlManagedInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Sql.ManagedInstancesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := sqlSvc.ParseManagedInstanceResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	name := id.Name

	future, err := client.Delete(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("Error deleting SQL Managed Instance %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for deletion of SQL Managed Instance %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	return nil
}
