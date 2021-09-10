package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSCloudfrontFieldLevelEncryptionProfile_basic(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	keyResourceName := "aws_cloudfront_public_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.provider_id", rName),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_entities.0.public_key_id", keyResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.field_patterns.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileExtendedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some other comment"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.provider_id", rName),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_entities.0.public_key_id", keyResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.0.field_patterns.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
		},
	})
}

func TestAccAWSCloudfrontFieldLevelEncryptionProfile_disappears(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudfrontFieldLevelEncryptionProfile(), resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudfrontFieldLevelEncryptionProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_field_level_encryption_profile" {
			continue
		}

		_, err := finder.FieldLevelEncryptionProfileByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err == nil {
			return fmt.Errorf("cloudfront Field Level Encryption Profile was not deleted")
		}
	}

	return nil
}

func testAccCheckCloudfrontFieldLevelEncryptionProfileExists(r string, profile *cloudfront.GetFieldLevelEncryptionProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Id is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		resp, err := finder.FieldLevelEncryptionProfileByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error retrieving Cloudfront Field Level Encryption Profile: %w", err)
		}

		*profile = *resp

		return nil
	}
}

func testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some comment"
  name    = %[1]q

  encryption_entities {
    public_key_id  = aws_cloudfront_public_key.test.id
    provider_id    = %[1]q
    field_patterns = ["DateOfBirth"]
  }
}
`, rName)
}

func testAccAWSCloudfrontFieldLevelEncryptionProfileExtendedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some other comment"
  name    = %[1]q

  encryption_entities {
    public_key_id  = aws_cloudfront_public_key.test.id
    provider_id    = %[1]q
    field_patterns = ["FirstName", "DateOfBirth"]
  }
}
`, rName)
}
