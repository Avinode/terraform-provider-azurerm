package tests

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
)

func TestAccAzureRMSiteRecoveryReplicatedVm_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_site_recovery_replicated_vm", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMSiteRecoveryReplicatedVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMSiteRecoveryReplicatedVm_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSiteRecoveryReplicatedVmExists(data.ResourceName),
					testCheckCorrectTargetSubnet(data.ResourceName, fmt.Sprintf("snet-%d_2", data.RandomInteger)),
				),
			},
			data.ImportStep(),
		},
	})
}

func testAccAzureRMSiteRecoveryReplicatedVm_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-recovery-%d-1"
  location = "%s"
}

resource "azurerm_resource_group" "test2" {
  name     = "acctestRG-recovery-%d-2"
  location = "%s"
}

resource "azurerm_recovery_services_vault" "test" {
  name                = "acctest-vault-%d"
  location            = azurerm_resource_group.test2.location
  resource_group_name = azurerm_resource_group.test2.name
  sku                 = "Standard"

  soft_delete_enabled = false
}

resource "azurerm_site_recovery_fabric" "test1" {
  resource_group_name = azurerm_resource_group.test2.name
  recovery_vault_name = azurerm_recovery_services_vault.test.name
  name                = "acctest-fabric1-%d"
  location            = azurerm_resource_group.test.location
}

resource "azurerm_site_recovery_fabric" "test2" {
  resource_group_name = azurerm_resource_group.test2.name
  recovery_vault_name = azurerm_recovery_services_vault.test.name
  name                = "acctest-fabric2-%d"
  location            = azurerm_resource_group.test2.location
  depends_on          = [azurerm_site_recovery_fabric.test1]
}

resource "azurerm_site_recovery_protection_container" "test1" {
  resource_group_name  = azurerm_resource_group.test2.name
  recovery_vault_name  = azurerm_recovery_services_vault.test.name
  recovery_fabric_name = azurerm_site_recovery_fabric.test1.name
  name                 = "acctest-protection-cont1-%d"
}

resource "azurerm_site_recovery_protection_container" "test2" {
  resource_group_name  = azurerm_resource_group.test2.name
  recovery_vault_name  = azurerm_recovery_services_vault.test.name
  recovery_fabric_name = azurerm_site_recovery_fabric.test2.name
  name                 = "acctest-protection-cont2-%d"
}

resource "azurerm_site_recovery_replication_policy" "test" {
  resource_group_name                                  = azurerm_resource_group.test2.name
  recovery_vault_name                                  = azurerm_recovery_services_vault.test.name
  name                                                 = "acctest-policy-%d"
  recovery_point_retention_in_minutes                  = 24 * 60
  application_consistent_snapshot_frequency_in_minutes = 4 * 60
}

resource "azurerm_site_recovery_protection_container_mapping" "test" {
  resource_group_name                       = azurerm_resource_group.test2.name
  recovery_vault_name                       = azurerm_recovery_services_vault.test.name
  recovery_fabric_name                      = azurerm_site_recovery_fabric.test1.name
  recovery_source_protection_container_name = azurerm_site_recovery_protection_container.test1.name
  recovery_target_protection_container_id   = azurerm_site_recovery_protection_container.test2.id
  recovery_replication_policy_id            = azurerm_site_recovery_replication_policy.test.id
  name                                      = "mapping-%d"
}

resource "azurerm_virtual_network" "test1" {
  name                = "net-%d"
  resource_group_name = azurerm_resource_group.test.name
  address_space       = ["192.168.1.0/24"]
  location            = azurerm_site_recovery_fabric.test1.location
}

resource "azurerm_subnet" "test1" {
  name                 = "snet-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test1.name
  address_prefix       = "192.168.1.0/24"
}

resource "azurerm_virtual_network" "test2" {
  name                = "net-%d"
  resource_group_name = azurerm_resource_group.test2.name
  address_space       = ["192.168.2.0/24"]
  location            = azurerm_site_recovery_fabric.test2.location
}
resource "azurerm_subnet" "test2_1" {
  name                 = "acctest-snet-%d_1"
  resource_group_name  = "${azurerm_resource_group.test2.name}"
  virtual_network_name = "${azurerm_virtual_network.test2.name}"
  address_prefix       = "192.168.2.0/27"
}

resource "azurerm_subnet" "test2_2" {
  name                 = "snet-%d_2"
  resource_group_name  = "${azurerm_resource_group.test2.name}"
  virtual_network_name = "${azurerm_virtual_network.test2.name}"
  address_prefix       = "192.168.2.32/27"
}

resource "azurerm_subnet" "test2_3" {
  name                 = "snet-%d_3"
  resource_group_name  = "${azurerm_resource_group.test2.name}"
  virtual_network_name = "${azurerm_virtual_network.test2.name}"
  address_prefix       = "192.168.2.64/27"
}

resource "azurerm_site_recovery_network_mapping" "test" {
  resource_group_name         = azurerm_resource_group.test2.name
  recovery_vault_name         = azurerm_recovery_services_vault.test.name
  name                        = "mapping-%d"
  source_recovery_fabric_name = azurerm_site_recovery_fabric.test1.name
  target_recovery_fabric_name = azurerm_site_recovery_fabric.test2.name
  source_network_id           = azurerm_virtual_network.test1.id
  target_network_id           = azurerm_virtual_network.test2.id
}

resource "azurerm_network_interface" "test" {
  name                = "vm-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "vm-%d"
    subnet_id                     = azurerm_subnet.test1.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_virtual_machine" "test" {
  name                = "vm-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  vm_size = "Standard_B1s"

  storage_image_reference {
    publisher = "OpenLogic"
    offer     = "CentOS"
    sku       = "7.5"
    version   = "latest"
  }

  storage_os_disk {
    name              = "disk-%d"
    os_type           = "Linux"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Premium_LRS"
  }

  os_profile {
    admin_username = "testadmin"
    admin_password = "Password1234!"
    computer_name  = "vm-%d"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
  network_interface_ids = [azurerm_network_interface.test.id]
}

resource "azurerm_storage_account" "test" {
  name                     = "acct%d"
  location                 = azurerm_resource_group.test.location
  resource_group_name      = azurerm_resource_group.test.name
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_site_recovery_replicated_vm" "test" {
  name                                      = "repl-%d"
  resource_group_name                       = azurerm_resource_group.test2.name
  recovery_vault_name                       = azurerm_recovery_services_vault.test.name
  source_vm_id                              = azurerm_virtual_machine.test.id
  source_recovery_fabric_name               = azurerm_site_recovery_fabric.test1.name
  recovery_replication_policy_id            = azurerm_site_recovery_replication_policy.test.id
  source_recovery_protection_container_name = azurerm_site_recovery_protection_container.test1.name

  target_resource_group_id                = azurerm_resource_group.test2.id
  target_recovery_fabric_id               = azurerm_site_recovery_fabric.test2.id
  target_recovery_protection_container_id = azurerm_site_recovery_protection_container.test2.id

  managed_disk {
    disk_id                    = azurerm_virtual_machine.test.storage_os_disk[0].managed_disk_id
    staging_storage_account_id = azurerm_storage_account.test.id
    target_resource_group_id   = azurerm_resource_group.test2.id
    target_disk_type           = "Premium_LRS"
    target_replica_disk_type   = "Premium_LRS"
  }

  network_interface {
    source_network_interface_id = azurerm_network_interface.test.id
    target_subnet_name          = "snet-%d_2"
  }

  depends_on = [
    azurerm_site_recovery_protection_container_mapping.test,
    azurerm_site_recovery_network_mapping.test,
  ]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.Locations.Secondary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func testCheckAzureRMSiteRecoveryReplicatedVmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		state, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		resourceGroupName := state.Primary.Attributes["resource_group_name"]
		vaultName := state.Primary.Attributes["recovery_vault_name"]
		fabricName := state.Primary.Attributes["source_recovery_fabric_name"]
		protectionContainerName := state.Primary.Attributes["source_recovery_protection_container_name"]
		replicationName := state.Primary.Attributes["name"]

		client := acceptance.AzureProvider.Meta().(*clients.Client).RecoveryServices.ReplicationMigrationItemsClient(resourceGroupName, vaultName)

		resp, err := client.Get(ctx, fabricName, protectionContainerName, replicationName)
		if err != nil {
			return fmt.Errorf("Bad: Get on replicationVmClient: %+v", err)
		}

		if resp.Response.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: fabric: %q does not exist", fabricName)
		}

		return nil
	}
}

func testCheckCorrectTargetSubnet(resourceName string, expectedTargetSubnet string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		state, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		keyMatcher := regexp.MustCompile(`network_interface\.[0-9]*\.target_subnet_name`)

		for key, value := range state.Primary.Attributes {
			log.Printf("Testing state key %q", key)
			if keyMatcher.MatchString(key) {
				if value == expectedTargetSubnet {
					return nil
				} else {
					return fmt.Errorf("Bad target subnet: %q (expected %q)", value, expectedTargetSubnet)
				}
			}
		}
		return fmt.Errorf("Target subnet not found in state.")
	}
}

func testCheckAzureRMSiteRecoveryReplicatedVmDestroy(s *terraform.State) error {
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_site_recovery_replicated_vm" {
			continue
		}

		resourceGroupName := rs.Primary.Attributes["resource_group_name"]
		vaultName := rs.Primary.Attributes["recovery_vault_name"]
		fabricName := rs.Primary.Attributes["source_recovery_fabric_name"]
		protectionContainerName := rs.Primary.Attributes["source_recovery_protection_container_name"]
		replicationName := rs.Primary.Attributes["name"]

		client := acceptance.AzureProvider.Meta().(*clients.Client).RecoveryServices.ReplicationMigrationItemsClient(resourceGroupName, vaultName)

		resp, err := client.Get(ctx, fabricName, protectionContainerName, replicationName)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Replicated VM still exists:\n%#v", resp.Properties)
		}
	}

	return nil
}
