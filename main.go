package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strconv"
	"strings"
	"time"
)

var (
	ec2_svc *ec2.EC2
)

type SnapperInput struct {
	Option string `json:"option"`
}

func deleteInstanceSnapshots() string {

	errorSnapshotStrings := ""

	res, err := getInstances()
	if err != nil {
		return "Error getting instances with the Snapper: create tags!"
	}

	for _, i := range res.Reservations {
		if *i.Instances[0].State.Name == "running" {
			keep := ""
			for _, a := range i.Instances[0].Tags {
				if *a.Key == "Snapper" {
					keep = *a.Value
				}
			}
			if keep == "" {
				keep = "7"
			}
			keepInt, _ := strconv.ParseFloat(keep, 64)
			for _, blockDevice := range i.Instances[0].BlockDeviceMappings {
				snapshots := getSnapshots(blockDevice.Ebs.VolumeId)
				for _, snap := range snapshots {
					fmt.Printf("VolumeId: %v\nVolumeSize: %vGB\nSnapshotId: %v\nStartTime: %v\nKeepDays: %v (%v Hours)\n", *snap.VolumeId, *snap.VolumeSize, *snap.SnapshotId, *snap.StartTime, keepInt, keepInt*24)
					fmt.Printf("Time Now: %v\nTime Snapshot: %v\nTime Difference Hours: %v\n\n", time.Now().UTC(), *snap.StartTime, time.Now().UTC().Sub(*snap.StartTime).Hours())
					if time.Now().UTC().Sub(*snap.StartTime).Hours() > keepInt*24 {
						errorSnapshot := deleteSnapshot(snap.SnapshotId)
						if errorSnapshot != nil {
							errorSnapshotStrings = strings.Join([]string{errorSnapshotStrings, errorSnapshot.Error()}, "\n")
						}
					}
				}
			}
		}
	}

	return errorSnapshotStrings
}

func deleteSnapshot(s *string) error {
	_, err := ec2_svc.DeleteSnapshot(&ec2.DeleteSnapshotInput{
		SnapshotId: s,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Snapshot marked for deletion: %v\n\n", *s)
	return nil
}

func getSnapshots(v *string) []*ec2.Snapshot {
	snapshots, err := ec2_svc.DescribeSnapshots(&ec2.DescribeSnapshotsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("volume-id"),
				Values: []*string{
					v,
				},
			},
			&ec2.Filter{
				Name: aws.String("status"),
				Values: []*string{
					aws.String("completed"),
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("%v", err)
	}

	return snapshots.Snapshots
}

func createInstanceSnapshots() string {

	errorSnapshotStrings := ""

	res, err := getInstances()
	if err != nil {
		return "Error getting instances with the Snapper: create tags!"
	}

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
				errorSnapshot := createSnapshot(nt, *blockDevice.Ebs.VolumeId, *i.Instances[0].InstanceId, *i.Instances[0].VpcId)
				if errorSnapshot != nil {
					errorSnapshotStrings = strings.Join([]string{errorSnapshotStrings, errorSnapshot.Error()}, "\n")
				}
			}
		}
	}

	return errorSnapshotStrings
}

func createSnapshot(nt string, v string, in string, vpc string) error {

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

func getInstances() (*ec2.DescribeInstancesOutput, error) {

	var res *ec2.DescribeInstancesOutput
	var err error
	res, err = ec2_svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("tag:Snapper"),
				Values: []*string{
					aws.String("*"),
				},
			},
		},
	})

	if err != nil {
		return res, err
	}

	return res, nil
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

func HandleRequest(ctx context.Context, input SnapperInput) (string, error) {

	aws_err := aws_initiate()

	if aws_err != nil {
		return "Error!", aws_err
	}

	fmt.Printf("Request: %#v\n", input)

	if input.Option == "delete" {
		errorStringDeleteSnapshots := deleteInstanceSnapshots()
		if errorStringDeleteSnapshots != "" {
			return errorStringDeleteSnapshots, errors.New("Snapshots cannot be deleted!")
		}
		return "Snapshots successfully scheduled for deletion.", nil
	} else if input.Option == "create" {
		errorStringInstances := createInstanceSnapshots()
		if errorStringInstances != "" {
			return errorStringInstances, errors.New("Some snapshots cannot be created!")
		}
		return "Snapshots successfully scheduled for creation.", nil
	} else {
		return "Incorrect input!", errors.New("Incorrect input!")
	}

}

func main() {
	lambda.Start(HandleRequest)
}
