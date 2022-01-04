package ebs

import (
	"github.com/aquasecurity/defsec/rules"
	"github.com/aquasecurity/defsec/rules/aws/ebs"
	"github.com/aquasecurity/tfsec/internal/app/tfsec/block"
	"github.com/aquasecurity/tfsec/internal/app/tfsec/scanner"
	"github.com/aquasecurity/tfsec/pkg/rule"
)

func init() {
	scanner.RegisterCheckRule(rule.Rule{
		BadExample: []string{`
 resource "aws_ebs_volume" "example" {
   availability_zone = "us-west-2a"
   size              = 40
 
   tags = {
     Name = "HelloWorld"
   }
 }
 `},
		GoodExample: []string{`
 resource "aws_kms_key" "ebs_encryption" {
 	enable_key_rotation = true
 }
 
 resource "aws_ebs_volume" "example" {
   availability_zone = "us-west-2a"
   size              = 40
 
   kms_key_id = aws_kms_key.ebs_encryption.arn
 
   tags = {
     Name = "HelloWorld"
   }
 }
 `},
		Links: []string{
			"https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ebs_volume#kms_key_id",
		},
		RequiredTypes: []string{
			"resource",
		},
		RequiredLabels: []string{
			"aws_ebs_volume",
		},
		Base: ebs.CheckEncryptionCustomerKey,
		CheckTerraform: func(resourceBlock block.Block, module block.Module) (results rules.Results) {

			if resourceBlock.MissingChild("kms_key_id") {
				results.Add("Resource does not use CMK", resourceBlock)
				return
			}

			kmsKeyAttr := resourceBlock.GetAttribute("kms_key_id")
			if kmsKeyAttr.IsDataBlockReference() {
				kmsData, err := module.GetReferencedBlock(kmsKeyAttr, resourceBlock)
				if err != nil {
					return
				}
				keyIdAttr := kmsData.GetAttribute("key_id")
				if keyIdAttr.IsNotNil() && keyIdAttr.StartsWith("alias/aws/") {
					results.Add("Resource explicitly uses the default CMK", keyIdAttr)
				}
			}
			return results
		},
	})
}