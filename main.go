package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
)

var (
	ec2_svc *ec2.EC2
)

func getInstancesByTag() string {
	res, err := ec2_svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("tag:Snapper"),
				Values: []*string{
					aws.String("create"),
				},
			},
		},
	})

	if err != nil {
		return err.Error()
	}

	errorSnapshotStrings := ""

	for _, i := range res.Reservations {
		var nt string
		for _, t := range i.Instances[0].Tags {
			if *t.Key == "Name" {
				nt = *t.Value
				break
			}
		}
		fmt.Println(nt, *i.Instances[0].InstanceId, *i.Instances[0].State.Name)
		if *i.Instances[0].State.Name == "running" {
			for _, blockDevice := range i.Instances[0].BlockDeviceMappings {
				errorSnapshot := createSnapShot(nt, *blockDevice.Ebs.VolumeId, *i.Instances[0].InstanceId, *i.Instances[0].VpcId)
				if errorSnapshot != nil {
					errorSnapshotStrings = strings.Join([]string{errorSnapshotStrings, errorSnapshot.Error()}, "\n")
				}
			}
		}
	}

	if errorSnapshotStrings != "" {
		return errorSnapshotStrings
	}

	return ""
}

func createSnapShot(nt string, v string, in string, vpc string) error {

	s, err := ec2_svc.CreateSnapshot(&ec2.CreateSnapshotInput{
		Description: aws.String(strings.Join([]string{"Created by Snapper for volume id:", v, "instance id:", in, "vpc id:", vpc}, " ")),
		VolumeId:    aws.String(v),
		TagSpecifications: []*ec2.TagSpecification{
			&ec2.TagSpecification{
				ResourceType: aws.String("snapshot"),
				Tags: []*ec2.Tag{
					&ec2.Tag{
						Key:   aws.String("Name"),
						Value: aws.String(nt),
					},
				},
			},
		},
	})

	if err != nil {
		return err
	}

	fmt.Printf("Creating snapshot for %s...\n", nt)
	fmt.Println(s)

	return nil
}

func aws_initiate() error {
	var err error
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		return err
	}
	ec2_svc = ec2.New(sess)
	return nil
}

func start(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	aws_err := aws_initiate()

	if aws_err != nil {
		response, _ := json.Marshal("Error occured when creating AWS session. Check server logs.")
		return events.APIGatewayProxyResponse{
			Body:       string(response),
			StatusCode: 500,
		}, aws_err
	}

	errorString := getInstancesByTag()
	if errorString != "" {
		response, _ := json.Marshal(errorString)
		return events.APIGatewayProxyResponse{
			Body:       string(response),
			StatusCode: 500,
		}, nil
	}

	response, _ := json.Marshal("Return 200.")
	return events.APIGatewayProxyResponse{
		Body:       string(response),
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(start)
}
