package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/cucumber/godog"
)

type ec2ApiTest struct {
	*ec2Api
	resp *string
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func (e *ec2ApiTest) addTags(tags *godog.Table) error {
	if len(tags.Rows[1:]) == 0 {
		return nil
	}
	tagsMap := map[string][]types.Tag{}
	for _, t := range tags.Rows[1:] {
		resources := strings.Split(t.Cells[2].Value, ",")
		for _, res := range resources {
			r := strings.TrimSpace(res)
			tagsMap[r] = append(tagsMap[r], types.Tag{
				Key:   aws.String(t.Cells[0].Value),
				Value: aws.String(t.Cells[1].Value),
			})
		}
	}
	tagSpec := []types.TagSpecification{}
	for resource, tags := range tagsMap {
		tagSpec = append(tagSpec, types.TagSpecification{
			ResourceType: types.ResourceType(resource),
			Tags:         tags,
		})
	}
	return e.SetTags(tagSpec)
}

func (e *ec2ApiTest) prepInput(resourceCount int) (err error) {
	err = e.PrepRunInstancesInput(&ec2.RunInstancesInput{
		MaxCount:     aws.Int32(int32(resourceCount)),
		MinCount:     aws.Int32(int32(resourceCount)),
		DryRun:       aws.Bool(true),
		EbsOptimized: aws.Bool(false),
		ImageId:      aws.String("ami-08ba52a61087f1bd6"), // latest amazon linux image
		InstanceType: types.InstanceTypeT3Micro,
	})
	return err
}

func (e *ec2ApiTest) findNetwork(subnetName string, vpcName string) (err error) {
	var vpcId, subnetId, defaultSecurityGroupId string
	if vpcId, err = e.FindVpcId(map[string]string{"Name": vpcName}); err != nil {
		return err
	}
	if defaultSecurityGroupId, err = e.FindDefaultSecurityGroupId(vpcId); err != nil {
		return err
	}
	if subnetId, err = e.FindSubnetId(vpcId, map[string]string{"Name": subnetName}); err != nil {
		return err
	}
	e.instancesInput.NetworkInterfaces = []types.InstanceNetworkInterfaceSpecification{
		{
			DeviceIndex: aws.Int32(0),
			Groups:      []string{defaultSecurityGroupId},
			SubnetId:    aws.String(subnetId),
		},
	}
	return err
}

func (e *ec2ApiTest) setAssociatePublicIpAddressFlag(enableStr string) (err error) {
	var enable bool
	if enable, err = strconv.ParseBool(enableStr); err != nil {
		return err
	}
	e.instancesInput.NetworkInterfaces[0].AssociatePublicIpAddress = &enable
	return nil
}

func (e *ec2ApiTest) responseIs(resp string) error {
	if *e.resp != resp {
		return fmt.Errorf("want %q, but got %q", resp, *e.resp)
	}
	return nil
}

func (e *ec2ApiTest) launchInstance() (err error) {
	var resp string
	apiResp, apiErr := e.RunInstances(context.TODO())
	if apiErr != nil {
		if resp, err = identifyErr(apiErr); err != nil {
			return err
		} else {
			e.resp = &resp
			return nil
		}
	}
	e.resp = apiResp.Instances[0].InstanceId
	return
}

func identifyErr(apiErr error) (resp string, err error) {
	var ae smithy.APIError
	if errors.As(apiErr, &ae) {
		switch ae.ErrorCode() {
		// Dont fail the test for DryRunOperation error
		case "DryRunOperation":
			resp = "OK"
			err = nil
		case "UnauthorizedOperation": // Missing permissions can also result in this error
			if errFromScp(ae.ErrorMessage()) {
				resp = ae.ErrorCode()
				err = nil
			} else {
				log.Printf("code: %s, message: %s, fault: %s", ae.ErrorCode(), ae.ErrorMessage(), ae.ErrorFault().String())
				err = apiErr
			}
		case "TagPolicyViolation":
			resp = ae.ErrorCode()
			err = nil
		default:
			log.Printf("code: %s, message: %s, fault: %s", ae.ErrorCode(), ae.ErrorMessage(), ae.ErrorFault().String())
			err = apiErr
		}
	}
	return
}

func errFromScp(errMsg string) (match bool) {
	var err error
	if match, err = regexp.MatchString(`.+ with an explicit deny in a service control policy\. Encoded authorization failure message: .+`, errMsg); err != nil {
		return match
	}
	return match
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	api, err := NewEc2Api()
	if err != nil {
		panic(err)
	}
	apiTest := &ec2ApiTest{ec2Api: api}

	ctx.Step(`^I want to launch ([\d]+) ec2 instance$`, apiTest.prepInput)
	ctx.Step(`^use subnet "([\w-]+)" in vpc "([\w-]+)"$`, apiTest.findNetwork)
	ctx.Step(`^set flag "AssociatePublicIpAddress" to "(true|false)"$`, apiTest.setAssociatePublicIpAddressFlag)
	ctx.Step(`^add tags:$`, apiTest.addTags)
	ctx.Step(`^I launch the ec2 instance$`, apiTest.launchInstance)
	ctx.Step(`^the response is "([a-zA-Z]+)"$`, apiTest.responseIs)
}
