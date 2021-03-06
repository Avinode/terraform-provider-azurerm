---
subcategory: "Recovery Services"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_recovery_services_replication_policy"
description: |-
    Manages a site recovery services replication policy on Azure.
---

# azurerm_recovery_services_replication_policy

~> **NOTE:** This resource has been deprecated in favour of the `azurerm_site_recovery_replication_policy` resource and will be removed in the next major version of the AzureRM Provider. The new resource shares the same fields as this one, and information on migrating across [can be found in this guide](../guides/migrating-between-renamed-resources.html).

Manages a Azure recovery vault replication policy.

## Example Usage

```hcl
resource "azurerm_resource_group" "secondary" {
  name     = "tfex-network-mapping-secondary"
  location = "East US"
}

resource "azurerm_recovery_services_vault" "vault" {
  name                = "example-recovery-vault"
  location            = "${azurerm_resource_group.secondary.location}"
  resource_group_name = "${azurerm_resource_group.secondary.name}"
  sku                 = "Standard"
}

resource "azurerm_recovery_services_replication_policy" "policy" {
  name                                                 = "policy"
  resource_group_name                                  = "${azurerm_resource_group.secondary.name}"
  recovery_vault_name                                  = "${azurerm_recovery_services_vault.vault.name}"
  recovery_point_retention_in_minutes                  = "${24 * 60}"
  application_consistent_snapshot_frequency_in_minutes = "${4 * 60}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the network mapping.

* `resource_group_name` - (Required) Name of the resource group where the vault that should be updated is located.

* `recovery_vault_name` - (Required) The name of the vault that should be updated.

* `recovery_point_retention_in_minutes` - (Required) Retain the recovery points for given time in minutes.

* `application_consistent_snapshot_frequency_in_minutes` - (Required) Specifies the frequency(in minutes) at which to create application consistent recovery points.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `id` - The ID of the Recovery Services Replication Policy.

### Timeouts

~> **Note:** Custom Timeouts are available [as an opt-in Beta in version 1.43 of the Azure Provider](/docs/providers/azurerm/guides/2.0-beta.html) and will be enabled by default in version 2.0 of the Azure Provider.

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Recovery Services Replication Policy.
* `update` - (Defaults to 30 minutes) Used when updating the Recovery Services Replication Policy.
* `read` - (Defaults to 5 minutes) Used when retrieving the Recovery Services Replication Policy.
* `delete` - (Defaults to 30 minutes) Used when deleting the Recovery Services Replication Policy.

## Import

Recovery Services Replication Policies can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_recovery_services_protection_container.mycontainer /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resource-group-name/providers/Microsoft.RecoveryServices/vaults/recovery-vault-name/replicationPolicies/policy-name
```
