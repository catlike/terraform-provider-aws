// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package wafregional

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceIPSet,
			TypeName: "aws_wafregional_ipset",
			Name:     "IPSet",
		},
		{
			Factory:  dataSourceRateBasedRule,
			TypeName: "aws_wafregional_rate_based_rule",
			Name:     "Rate Based Rule",
		},
		{
			Factory:  dataSourceRule,
			TypeName: "aws_wafregional_rule",
			Name:     "Rule",
		},
		{
			Factory:  dataSourceSubscribedRuleGroup,
			TypeName: "aws_wafregional_subscribed_rule_group",
			Name:     "Subscribed Rule Group",
		},
		{
			Factory:  dataSourceWebACL,
			TypeName: "aws_wafregional_web_acl",
			Name:     "Web ACL",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  resourceByteMatchSet,
			TypeName: "aws_wafregional_byte_match_set",
			Name:     "Byte Match Set",
		},
		{
			Factory:  resourceGeoMatchSet,
			TypeName: "aws_wafregional_geo_match_set",
			Name:     "Geo Match Set",
		},
		{
			Factory:  resourceIPSet,
			TypeName: "aws_wafregional_ipset",
			Name:     "IPSet",
		},
		{
			Factory:  resourceRateBasedRule,
			TypeName: "aws_wafregional_rate_based_rule",
			Name:     "Rate Based Rule",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceRegexMatchSet,
			TypeName: "aws_wafregional_regex_match_set",
			Name:     "Regex Match Set",
		},
		{
			Factory:  resourceRegexPatternSet,
			TypeName: "aws_wafregional_regex_pattern_set",
			Name:     "Regex Pattern Set",
		},
		{
			Factory:  resourceRule,
			TypeName: "aws_wafregional_rule",
			Name:     "Rule",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceRuleGroup,
			TypeName: "aws_wafregional_rule_group",
			Name:     "Rule Group",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceSizeConstraintSet,
			TypeName: "aws_wafregional_size_constraint_set",
			Name:     "Size Constraint Set",
		},
		{
			Factory:  resourceSQLInjectionMatchSet,
			TypeName: "aws_wafregional_sql_injection_match_set",
			Name:     "SQL Injection Match Set",
		},
		{
			Factory:  resourceWebACL,
			TypeName: "aws_wafregional_web_acl",
			Name:     "Web ACL",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceWebACLAssociation,
			TypeName: "aws_wafregional_web_acl_association",
			Name:     "Web ACL Association",
		},
		{
			Factory:  resourceXSSMatchSet,
			TypeName: "aws_wafregional_xss_match_set",
			Name:     "XSS Match Set",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.WAFRegional
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*wafregional.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*wafregional.Options){
		wafregional.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		withExtraOptions(ctx, p, config),
	}

	return wafregional.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*wafregional.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*wafregional.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *wafregional.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*wafregional.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
