package azurerm

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/2017-03-01-preview/sql"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	sqlSvc "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/sql"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

// TODO: update the api version
func resourceArmSqlManagedInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSqlManagedInstanceCreate,
		Read:   resourceArmSqlManagedInstanceServerRead,
		Update: resourceArmSqlManagedInstanceUpdate,
		Delete: resourceArmSqlManagedInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(6 * time.Hour),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(6 * time.Hour),
			Delete: schema.DefaultTimeout(6 * time.Hour),
		},

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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // TODO: is it?
				ValidateFunc: validate.NoEmptyStrings,
			},

			"administrator_login_password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validate.StringAtLeast(16),
			},

			"sku_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"BC_Gen4",
					"BC_Gen5",
					"GP_Gen4",
					"GP_Gen5",
					// TODO: do we also need "Edition" in the SKU block?
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
					4,
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
			"collation": {
				Type:     schema.TypeString,
				Optional: true,
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

			"proxy_override": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(sql.ManagedInstanceProxyOverrideProxy),
				ValidateFunc: validation.StringInSlice([]string{
					string(sql.ManagedInstanceProxyOverrideProxy),
					string(sql.ManagedInstanceProxyOverrideRedirect),
				}, false),
			},

			"public_data_endpoint_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"time_zone": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.VirtualMachineTimeZone(),
				// TODO: confirm this is correct
			},

			// Computed
			"fully_qualified_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceArmSqlManagedInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).Sql.ManagedInstancesClient
	ctx, cancel := timeouts.ForCreate(meta.(*ArmClient).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))
	adminUsername := d.Get("administrator_login").(string)
	adminPassword := d.Get("administrator_login_password").(string)
	licenseType := d.Get("license_type").(string)
	proxyOverride := d.Get("proxy_override").(string)
	publicDataEndpointEnabled := d.Get("public_data_endpoint_enabled").(bool)
	skuName := d.Get("sku_name").(string)
	storageSizeInGb := d.Get("storage_size_in_gb").(int)
	subnetId := d.Get("subnet_id").(string)
	t := d.Get("tags").(map[string]interface{})
	vCores := d.Get("vcores").(int)

	// TODO: lock on the subnet id?

	if features.ShouldResourcesBeImported() {
		existing, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing SQL Managed Instance %q (Resource Group %q): %s", name, resourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_sql_managed_instance", *existing.ID)
		}
	}

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

			PublicDataEndpointEnabled: utils.Bool(publicDataEndpointEnabled),
			ProxyOverride:             sql.ManagedInstanceProxyOverride(proxyOverride),
		},
		Sku: &sql.Sku{
			Name: utils.String(skuName),
		},
	}

	if v, ok := d.GetOk("collation"); ok {
		parameters.ManagedInstanceProperties.Collation = utils.String(v.(string))
	}

	if v, ok := d.GetOk("time_zone"); ok {
		parameters.ManagedInstanceProperties.TimezoneID = utils.String(v.(string))
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
	ctx, cancel := timeouts.ForUpdate(meta.(*ArmClient).StopContext, d)
	defer cancel()

	id, err := sqlSvc.ParseManagedInstanceResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	name := id.Name

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

	if d.HasChange("collation") {
		collation := d.Get("collation").(string)
		parameters.ManagedInstanceProperties.Collation = utils.String(collation)
	}

	if d.HasChange("license_type") {
		licenseType := d.Get("license_type").(string)
		parameters.ManagedInstanceProperties.LicenseType = sql.ManagedInstanceLicenseType(licenseType)
	}

	if d.HasChange("proxy_override") {
		proxyOverride := d.Get("proxy_override").(string)
		parameters.ManagedInstanceProperties.ProxyOverride = sql.ManagedInstanceProxyOverride(proxyOverride)
	}

	if d.HasChange("public_data_endpoint_enabled") {
		publicDataEndpointEnabled := d.Get("public_data_endpoint_enabled").(bool)
		parameters.ManagedInstanceProperties.PublicDataEndpointEnabled = utils.Bool(publicDataEndpointEnabled)
	}

	if d.HasChange("sku_name") {
		parameters.Sku = &sql.Sku{
			Name: utils.String(d.Get("sku").(string)),
		}
	}

	if d.HasChange("storage_size_in_gb") {
		storageSizeInGb := d.Get("storage_size_in_gb").(int)
		parameters.ManagedInstanceProperties.StorageSizeInGB = utils.Int32(int32(storageSizeInGb))
	}

	if d.HasChange("subnet_id") {
		// TODO: should we be locking on this?
		subnetId := d.Get("subnet_id").(string)
		parameters.ManagedInstanceProperties.SubnetID = utils.String(subnetId)
	}

	if d.HasChange("tags") {
		t := d.Get("tags").(map[string]interface{})
		parameters.Tags = tags.Expand(t)
	}

	if d.HasChange("time_zone") {
		timeZoneId := d.Get("time_zone").(string)
		parameters.ManagedInstanceProperties.TimezoneID = utils.String(timeZoneId)
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
	ctx, cancel := timeouts.ForRead(meta.(*ArmClient).StopContext, d)
	defer cancel()

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
		d.Set("collation", props.Collation)
		d.Set("fully_qualified_domain_name", props.FullyQualifiedDomainName)
		d.Set("license_type", string(props.LicenseType))
		d.Set("proxy_override", string(props.ProxyOverride))
		d.Set("public_data_endpoint_enabled", props.PublicDataEndpointEnabled)
		d.Set("subnet_id", props.SubnetID)
		d.Set("time_zone", props.TimezoneID)

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
	ctx, cancel := timeouts.ForDelete(meta.(*ArmClient).StopContext, d)
	defer cancel()

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
