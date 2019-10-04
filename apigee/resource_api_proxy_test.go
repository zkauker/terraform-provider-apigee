package apigee

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zambien/go-apigee-edge"
)

func TestAccProxy_Updated(t *testing.T) {

	vTestAccCheckProxyConfigRequired := testAccCheckProxyConfigRequiredTF12
	vTestAccCheckProxyConfigUpdated := testAccCheckProxyConfigUpdatedTF12
	if isTerraformVersionPriorTo("0.12") {
		vTestAccCheckProxyConfigRequired = testAccCheckProxyConfigRequired
		vTestAccCheckProxyConfigUpdated = testAccCheckProxyConfigUpdated
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: vTestAccCheckProxyConfigRequired,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists("apigee_api_proxy.foo_api_proxy", "foo_proxy_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "name", "foo_proxy_terraformed"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "bundle", "test-fixtures/helloworld_proxy.zip"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "revision", "1"),
				),
			},

			resource.TestStep{
				Config: vTestAccCheckProxyConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyExists("apigee_api_proxy.foo_api_proxy", "foo_proxy_terraformed_updated"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "name", "foo_proxy_terraformed_updated"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "bundle", "test-fixtures/helloworld_proxy.zip"),
					resource.TestCheckResourceAttr(
						"apigee_api_proxy.foo_api_proxy", "revision", "1"),
				),
			},
		},
	})
}

func testAccCheckProxyDestroy(s *terraform.State) error {

	client := testAccProvider.Meta().(*apigee.EdgeClient)

	if err := proxyDestroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckProxyExists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*apigee.EdgeClient)
		if err := proxyExistsHelper(s, client, name); err != nil {
			log.Printf("Error in testAccCheckProxyExists: %s", err)
			return err
		}
		return nil
	}
}

const testAccCheckProxyConfigRequired = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed"
   bundle       = "test-fixtures/helloworld_proxy.zip"
   bundle_sha   = "${base64sha256(file("test-fixtures/helloworld_proxy.zip"))}"
}
`

const testAccCheckProxyConfigUpdated = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed_updated"
   bundle       = "test-fixtures/helloworld_proxy.zip"
   bundle_sha   = "${base64sha256(file("test-fixtures/helloworld_proxy.zip"))}"
}
`

/*
 * Make possible to test with TF 0.12: file function can load only text files. New function was introduced for the usecase below.
 * For more details see: https://github.com/hashicorp/terraform/issues/21260
 */
const testAccCheckProxyConfigRequiredTF12 = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed"
   bundle       = "test-fixtures/helloworld_proxy.zip"
   bundle_sha   = filebase64sha256("test-fixtures/helloworld_proxy.zip")
}
`

const testAccCheckProxyConfigUpdatedTF12 = `
resource "apigee_api_proxy" "foo_api_proxy" {
   name  		= "foo_proxy_terraformed_updated"
   bundle       = "test-fixtures/helloworld_proxy.zip"
   bundle_sha   = filebase64sha256("test-fixtures/helloworld_proxy.zip")
}
`

func proxyDestroyHelper(s *terraform.State, client *apigee.EdgeClient) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No proxy ID is set")
		}

		_, _, err := client.Proxies.Get("foo_proxy")

		if err != nil {
			if strings.Contains(err.Error(), "404 ") {
				return nil
			}
			return fmt.Errorf("Received an error retrieving proxy  %+v\n", err)
		}
	}

	return fmt.Errorf("Proxy still exists")
}

func proxyExistsHelper(s *terraform.State, client *apigee.EdgeClient, name string) error {

	for _, r := range s.RootModule().Resources {
		id := r.Primary.ID

		if id == "" {
			return fmt.Errorf("No proxy ID is set")
		}

		if proxyData, _, err := client.Proxies.Get(name); err != nil {
			return fmt.Errorf("Received an error retrieving proxy  %+v\n", proxyData)
		} else {
			log.Printf("Created proxy name: %s", proxyData.Name)
		}

	}
	return nil
}
