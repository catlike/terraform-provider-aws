// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDataSource_full(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_full(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDataSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dataSource types.DataSource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_data_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceExists(ctx, resourceName, &dataSource),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceDataSource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_data_source" {
				continue
			}

			_, err := tfbedrockagent.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["data_source_id"], rs.Primary.Attributes["knowledge_base_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Data Source %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataSourceExists(ctx context.Context, n string, v *types.DataSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["data_source_id"], rs.Primary.Attributes["knowledge_base_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = "xxxx"
  data_deletion_policy = "RETAIN"

  data_source_configuration {
    type = "S3"
    s3_configuration {
      bucket_arn = "arn:aws:xxxx"
    }
  }
}
`, rName)
}

func testAccDataSourceConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_data_source" "test" {
  name                 = %[1]q
  knowledge_base_id    = "xxx"
  data_deletion_policy = "RETAIN"

  data_source_configuration {
    type = "S3"
    s3_configuration {
      bucket_arn = "arn:aws:s3:::xxxxx"
    }
  }
  vector_ingestion_configuration {
    chunking_configuration {
      chunking_strategy = "FIXED_SIZE"
      fixed_size_chunking_configuration {
        max_tokens         = 3
        overlap_percentage = 80
      }
    }
  }
}
`, rName)
}
