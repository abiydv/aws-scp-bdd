package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {}

type ec2Api struct {
	client         *ec2.Client
	instancesInput *ec2.RunInstancesInput
}

func NewEc2Api() (*ec2Api, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := ec2.NewFromConfig(awsConfig)
	return &ec2Api{client: client}, nil
}

func tagFilters(tags map[string]string) []types.Filter {
	var filters []types.Filter
	for k, v := range tags {
		filters = append(filters, types.Filter{Name: aws.String("tag:" + k), Values: []string{v}})
	}
	return filters
}

func (e *ec2Api) FindVpcId(vpcTags map[string]string) (vpcId string, err error) {
	var vpcRes *ec2.DescribeVpcsOutput
	if vpcRes, err = e.client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{
		Filters: tagFilters(vpcTags),
	}); err != nil {
		return "", err
	}
	if len(vpcRes.Vpcs) == 0 {
		return "", fmt.Errorf("no vpc found with tags: %+v", vpcTags)
	}
	return *vpcRes.Vpcs[0].VpcId, nil
}

func (e *ec2Api) FindSubnetId(vpcId string, subnetTags map[string]string) (subnetId string, err error) {
	var subnetRes *ec2.DescribeSubnetsOutput
	if subnetRes, err = e.client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{
		Filters: tagFilters(subnetTags),
	}); err != nil {
		return "", err
	}
	if len(subnetRes.Subnets) == 0 {
		return "", fmt.Errorf("no subnet found with tags: %+v", subnetTags)
	}
	return *subnetRes.Subnets[0].SubnetId, nil
}

func (e *ec2Api) FindDefaultSecurityGroupId(vpcId string) (defaultSecurityGroupId string, err error) {
	var sgRes *ec2.DescribeSecurityGroupsOutput
	if sgRes, err = e.client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{{Name: aws.String("vpc-id"), Values: []string{vpcId}}},
	}); err != nil {
		return "", err
	}
	return *sgRes.SecurityGroups[0].GroupId, nil
}

func (e *ec2Api) PrepRunInstancesInput(input *ec2.RunInstancesInput) error {
	e.instancesInput = input
	return nil
}

func (e *ec2Api) SetTags(tags []types.TagSpecification) error {
	if e.instancesInput != nil {
		e.instancesInput.TagSpecifications = tags
	}
	return nil
}

func (e *ec2Api) RunInstances(ctx context.Context) (*ec2.RunInstancesOutput, error) {
	return e.client.RunInstances(ctx, e.instancesInput)
}
