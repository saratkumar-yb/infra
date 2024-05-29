package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Schedule struct represents a schedule item in DynamoDB
type Schedule struct {
	InstanceID   string `json:"instance_id"`
	StartTime    string `json:"start_time"`
	StopTime     string `json:"stop_time"`
	Timezone     string `json:"timezone"`
	AWSRegion    string `json:"aws_region"`
	FriendlyName string `json:"friendly_name"`
}

// CloudProvider interface defines methods for managing schedules
type CloudProvider interface {
	CreateSchedule(instanceID, region, startTime, stopTime, timezone, friendlyName string)
	ListSchedules()
	DeleteSchedule(instanceID string)
}

// AWSProvider implements the CloudProvider interface for AWS
type AWSProvider struct{}

func (a AWSProvider) CreateSchedule(instanceID, region, startTime, stopTime, timezone, friendlyName string) {
	tz, err := resolveTimezone(timezone)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2Svc := ec2.New(sess, aws.NewConfig().WithRegion(region))

	if friendlyName == "" {
		name, err := getInstanceName(ec2Svc, instanceID)
		if err != nil {
			log.Fatalf("ERROR: Failed to get instance name: %v", err)
		}
		if name == "" {
			log.Fatalf("ERROR: Instance %s does not have a Name tag. Please provide --friendly-name.", instanceID)
		}
		friendlyName = name
	}

	svc := dynamodb.New(sess)

	schedule := Schedule{
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
		schedule := Schedule{}

		err = dynamodbattribute.UnmarshalMap(i, &schedule)
		if err != nil {
			log.Fatalf("ERROR: Got error unmarshalling: %v", err)
		}

		fmt.Printf("Instance ID: %s, Start Time: %s, Stop Time: %s, Timezone: %s, AWS Region: %s, Friendly Name: %s\n",
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

var tableName string

func init() {
	tableName = os.Getenv("TABLE_NAME")
	if tableName == "" {
		log.Fatalf("ERROR: TABLE_NAME environment variable is not set")
	}
}

var timezoneMapping = map[string]string{
	"EST":  "America/New_York",
	"PST":  "America/Los_Angeles",
	"CST":  "America/Chicago",
	"MST":  "America/Denver",
	"IST":  "Asia/Kolkata",
	"BST":  "Europe/London",
	"CET":  "Europe/Berlin",
	"EET":  "Europe/Helsinki",
	"GMT":  "Europe/London",
	"HST":  "Pacific/Honolulu",
	"AKST": "America/Anchorage",
	"AST":  "America/Halifax",
	"NST":  "America/St_Johns",
	"PDT":  "America/Los_Angeles",
	"EDT":  "America/New_York",
	"CDT":  "America/Chicago",
	"MDT":  "America/Denver",
	"CEST": "Europe/Berlin",
	"EEST": "Europe/Helsinki",
	"WET":  "Europe/Lisbon",
	"WEST": "Europe/Lisbon",
}

func resolveTimezone(abbreviation string) (string, error) {
	if tz, exists := timezoneMapping[strings.ToUpper(abbreviation)]; exists {
		return tz, nil
	}
	return "", fmt.Errorf("unsupported timezone abbreviation: %s", abbreviation)
}

func getInstanceName(ec2Svc *ec2.EC2, instanceID string) (string, error) {
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected 'create-schedule', 'list-schedules' or 'delete-schedule' subcommands")
		os.Exit(1)
	}

	var cloudType string

	switch os.Args[1] {
	case "create-schedule":
		createCmd := flag.NewFlagSet("create-schedule", flag.ExitOnError)
		instanceID := createCmd.String("instance-id", "", "The ID of the instance")
		startTime := createCmd.String("start-time", "", "The start time in format HH:MM")
		stopTime := createCmd.String("stop-time", "", "The stop time in format HH:MM")
		timezone := createCmd.String("timezone", "UTC", "The timezone for the schedule (e.g., 'EST', 'PST', 'IST')")
		awsRegion := createCmd.String("aws-region", "", "The AWS region for the instance")
		friendlyName := createCmd.String("friendly-name", "", "A friendly name for the instance")
		createCmd.StringVar(&cloudType, "cloud-type", "aws", "Cloud provider type (aws, gcp, azure)")
		createCmd.Parse(os.Args[2:])

		if *instanceID == "" || *startTime == "" || *stopTime == "" || *timezone == "" || *awsRegion == "" {
			createCmd.PrintDefaults()
			os.Exit(1)
		}

		var provider CloudProvider
		switch cloudType {
		case "aws":
			provider = AWSProvider{}
		default:
			fmt.Println("Unsupported cloud provider type. Please specify 'aws'.")
			os.Exit(1)
		}

		provider.CreateSchedule(*instanceID, *awsRegion, *startTime, *stopTime, *timezone, *friendlyName)

	case "list-schedules":
		listCmd := flag.NewFlagSet("list-schedules", flag.ExitOnError)
		listCmd.StringVar(&cloudType, "cloud-type", "aws", "Cloud provider type (aws, gcp, azure)")
		listCmd.Parse(os.Args[2:])

		var provider CloudProvider
		switch cloudType {
		case "aws":
			provider = AWSProvider{}
		default:
			fmt.Println("Unsupported cloud provider type. Please specify 'aws'.")
			os.Exit(1)
		}

		provider.ListSchedules()

	case "delete-schedule":
		deleteCmd := flag.NewFlagSet("delete-schedule", flag.ExitOnError)
		instanceID := deleteCmd.String("instance-id", "", "The ID of the instance to delete")
		deleteCmd.StringVar(&cloudType, "cloud-type", "aws", "Cloud provider type (aws, gcp, azure)")
		deleteCmd.Parse(os.Args[2:])

		if *instanceID == "" {
			deleteCmd.PrintDefaults()
			os.Exit(1)
		}

		var provider CloudProvider
		switch cloudType {
		case "aws":
			provider = AWSProvider{}
		default:
			fmt.Println("Unsupported cloud provider type. Please specify 'aws'.")
			os.Exit(1)
		}

		provider.DeleteSchedule(*instanceID)

	default:
		fmt.Println("expected 'create-schedule', 'list-schedules' or 'delete-schedule' subcommands")
		os.Exit(1)
	}
}
