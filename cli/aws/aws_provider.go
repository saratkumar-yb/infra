package aws

import (
	"fmt"
	"log"
	"os"

	"github.com/saratkumar-yb/infra/cli/helpers"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AWSProvider struct{}

var tableName = os.Getenv("TABLE_NAME")

func (a AWSProvider) CreateSchedule(instanceID, region, startTime, stopTime, timezone, friendlyName string) {
	tz, err := helpers.ResolveTimezone(timezone)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2Svc := ec2.New(sess, aws.NewConfig().WithRegion(region))

	if friendlyName == "" {
		name, err := GetInstanceName(ec2Svc, instanceID)
		if err != nil {
			log.Fatalf("ERROR: Failed to get instance name: %v", err)
		}
		if name == "" {
			log.Fatalf("ERROR: Instance %s does not have a Name tag. Please provide --friendly-name.", instanceID)
		}
		friendlyName = name
	}

	svc := dynamodb.New(sess)

	schedule := helpers.Schedule{
		InstanceID:   instanceID,
		StartTime:    startTime,
		StopTime:     stopTime,
		Timezone:     tz,
		AWSRegion:    region,
		FriendlyName: friendlyName,
	}

	av, err := dynamodbattribute.MarshalMap(schedule)
	if err != nil {
		log.Fatalf("ERROR: Got error marshalling new schedule item: %v", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatalf("ERROR: Got error calling PutItem: %v", err)
	}

	log.Printf("INFO: Successfully added schedule to table")
}

func (a AWSProvider) ListSchedules() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := svc.Scan(input)
	if err != nil {
		log.Fatalf("ERROR: Got error calling Scan: %v", err)
	}

	for _, i := range result.Items {
		schedule := helpers.Schedule{}

		err = dynamodbattribute.UnmarshalMap(i, &schedule)
		if err != nil {
			log.Fatalf("ERROR: Got error unmarshalling: %v", err)
		}

		log.Printf("INFO: Instance ID: %s, Start Time: %s, Stop Time: %s, Timezone: %s, AWS Region: %s, Friendly Name: %s",
			schedule.InstanceID, schedule.StartTime, schedule.StopTime, schedule.Timezone, schedule.AWSRegion, schedule.FriendlyName)
	}
}

func (a AWSProvider) DeleteSchedule(instanceID string) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"instance_id": {
				S: aws.String(instanceID),
			},
		},
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		log.Fatalf("ERROR: Got error calling DeleteItem: %v", err)
	}

	log.Printf("INFO: Successfully deleted schedule from table")
}

func GetInstanceName(ec2Svc *ec2.EC2, instanceID string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}
	result, err := ec2Svc.DescribeInstances(input)
	if err != nil {
		return "", err
	}

	if len(result.Reservations) > 0 && len(result.Reservations[0].Instances) > 0 {
		for _, tag := range result.Reservations[0].Instances[0].Tags {
			if *tag.Key == "Name" {
				return *tag.Value, nil
			}
		}
		return "", nil // No Name tag found
	}

	return "", fmt.Errorf("instance %s not found", instanceID)
}
