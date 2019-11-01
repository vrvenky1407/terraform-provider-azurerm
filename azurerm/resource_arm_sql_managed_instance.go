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
		Create: resourceArmSqlManagedInstanceCreate,
		Read:   resourceArmSqlManagedInstanceServerRead,
		Update: resourceArmSqlManagedInstanceUpdate,
		Delete: resourceArmSqlManagedInstanceDelete,
		// TOOD: timeouts

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateMsSqlServerName,
				// TODO: confirm validation
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"administrator_login": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // TODO: is it?
				// TODO: validation
			},

			"administrator_login_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				// TODO: validation
			},

			"sku_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"GP_Gen4",
					"GP_Gen5",
				}, false),
			},

			"storage_size_in_gb": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validate.IntBetweenAndDivisibleBy(32, 8000, 32),
			},

			"subnet_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"tags": tagsSchema(),

			"vcores": {
				Type:     schema.TypeInt,
				Required: true,
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

			// Optional
			"license_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(sql.LicenseIncluded),
				ValidateFunc: validation.StringInSlice([]string{
					string(sql.BasePrice),
					string(sql.LicenseIncluded),
				}, false),
			},

			// COmputed
			"fully_qualified_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceArmSqlManagedInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Sql.ManagedInstancesClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))
	adminUsername := d.Get("administrator_login").(string)
	adminPassword := d.Get("administrator_login_password").(string)
	licenseType := d.Get("license_type").(string)
	skuName := d.Get("sku_name").(string)
	storageSizeInGb := d.Get("storage_size_in_gb").(int)
	subnetId := d.Get("subnet_id").(string)
	t := d.Get("tags").(map[string]interface{})
	vCores := d.Get("vcores").(int)

	// TODO: lock on the subnet id
	// TODO: requires import

	parameters := sql.ManagedInstance{
		Location: utils.String(location),
		Tags:     tags.Expand(t),
		ManagedInstanceProperties: &sql.ManagedInstanceProperties{
			LicenseType:                sql.ManagedInstanceLicenseType(licenseType),
			AdministratorLogin:         utils.String(adminUsername),
			AdministratorLoginPassword: utils.String(adminPassword),
			SubnetID:                   utils.String(subnetId),
			StorageSizeInGB:            utils.Int32(int32(storageSizeInGb)),
			VCores:                     utils.Int32(int32(vCores)),

			//Collation:
			//TimezoneID:
		},
		Sku: &sql.Sku{
			Name: utils.String(skuName),
		},
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, name, parameters)
	if err != nil {
		// TODO: should this be here?
		if response.WasConflict(future.Response()) {
			return fmt.Errorf("SQL Server names need to be globally unique and %q is already in use.", name)
		}

		return fmt.Errorf("Error creating SQL Managed Instance %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for creation of SQL Managed Instance %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return err
	}

	d.SetId(*resp.ID)

	return resourceArmSqlManagedInstanceServerRead(d, meta)
}

func resourceArmSqlManagedInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Sql.ManagedInstancesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := sqlSvc.ParseManagedInstanceResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	name := id.Name

	// TODO: lock on the subnet id

	parameters := sql.ManagedInstanceUpdate{
		ManagedInstanceProperties: &sql.ManagedInstanceProperties{},
	}

	// TODO: is this required?
	if d.HasChange("administrator_login") {
		adminUsername := d.Get("administrator_login").(string)
		parameters.ManagedInstanceProperties.AdministratorLogin = utils.String(adminUsername)
	}

	if d.HasChange("administrator_login_password") {
		adminPassword := d.Get("administrator_login_password").(string)
		parameters.ManagedInstanceProperties.AdministratorLoginPassword = utils.String(adminPassword)
	}

	if d.HasChange("license_type") {
		licenseType := d.Get("license_type").(string)
		parameters.ManagedInstanceProperties.LicenseType = sql.ManagedInstanceLicenseType(licenseType)
	}

	if d.HasChange("sku_name") {
		parameters.Sku = &sql.Sku{
			Name: utils.String(d.Get("sku").(string)),
		}
	}

	if d.HasChange("") {
		storageSizeInGb := d.Get("storage_size_in_gb").(int)
		parameters.ManagedInstanceProperties.StorageSizeInGB = utils.Int32(int32(storageSizeInGb))
	}

	if d.HasChange("subnet_id") {
		subnetId := d.Get("subnet_id").(string)
		parameters.ManagedInstanceProperties.SubnetID = utils.String(subnetId)
	}

	if d.HasChange("tags") {
		t := d.Get("tags").(map[string]interface{})
		parameters.Tags = tags.Expand(t)
	}

	if d.HasChange("vcores") {
		vCores := d.Get("vcores").(int)
		parameters.ManagedInstanceProperties.VCores = utils.Int32(int32(vCores))
	}

	future, err := client.Update(ctx, resourceGroup, name, parameters)
	if err != nil {
		return fmt.Errorf("Error updating SQL Managed Instance %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for update of SQL Managed Instance %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

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

	if sku := resp.Sku; sku != nil {
		d.Set("sku_name", sku.Name)
	}

	if props := resp.ManagedInstanceProperties; props != nil {
		d.Set("administrator_login", props.AdministratorLogin)
		d.Set("fully_qualified_domain_name", props.FullyQualifiedDomainName)
		d.Set("license_type", string(props.LicenseType))
		d.Set("subnet_id", props.SubnetID)

		storageSizeInGb := 0
		if props.StorageSizeInGB != nil {
			storageSizeInGb = int(*props.StorageSizeInGB)
		}
		d.Set("storage_size_in_gb", storageSizeInGb)

		vCores := 0
		if props.VCores != nil {
			vCores = int(*props.VCores)
		}
		d.Set("vcores", vCores)
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
