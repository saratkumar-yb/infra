package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Schedule struct {
	InstanceID string `json:"instance_id"`
	StartTime  string `json:"start_time"`
	StopTime   string `json:"stop_time"`
	Timezone   string `json:"timezone"`
	AWSRegion  string `json:"aws_region"`
}

var tableName string

func init() {
	tableName = os.Getenv("TABLE_NAME")
	if tableName == "" {
		fmt.Println("TABLE_NAME environment variable is not set")
		os.Exit(1)
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
	// Add more mappings as needed
}

func resolveTimezone(abbreviation string) (string, error) {
	if tz, exists := timezoneMapping[strings.ToUpper(abbreviation)]; exists {
		return tz, nil
	}
	return "", fmt.Errorf("unsupported timezone abbreviation: %s", abbreviation)
}

func createSchedule(instanceID, startTime, stopTime, timezone, awsRegion string) {
	tz, err := resolveTimezone(timezone)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	schedule := Schedule{
		InstanceID: instanceID,
		StartTime:  startTime,
		StopTime:   stopTime,
		Timezone:   tz,
		AWSRegion:  awsRegion,
	}

	av, err := dynamodbattribute.MarshalMap(schedule)
	if err != nil {
		fmt.Println("Got error marshalling new schedule item:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Successfully added schedule to table")
}

func listSchedules() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := svc.Scan(input)
	if err != nil {
		fmt.Println("Got error calling Scan:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, i := range result.Items {
		schedule := Schedule{}

		err = dynamodbattribute.UnmarshalMap(i, &schedule)
		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Instance ID: %s, Start Time: %s, Stop Time: %s, Timezone: %s, AWS Region: %s\n", schedule.InstanceID, schedule.StartTime, schedule.StopTime, schedule.Timezone, schedule.AWSRegion)
	}
}

func deleteSchedule(instanceID string) {
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
		fmt.Println("Got error calling DeleteItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Successfully deleted schedule from table")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected 'create-schedule', 'list-schedules' or 'delete-schedule' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create-schedule":
		createCmd := flag.NewFlagSet("create-schedule", flag.ExitOnError)
		instanceID := createCmd.String("instance-id", "", "The ID of the instance")
		startTime := createCmd.String("start-time", "", "The start time in format HH:MM")
		stopTime := createCmd.String("stop-time", "", "The stop time in format HH:MM")
		timezone := createCmd.String("timezone", "UTC", "The timezone for the schedule (e.g., 'EST', 'PST', 'IST')")
		awsRegion := createCmd.String("aws-region", "", "The AWS region for the instance")
		createCmd.Parse(os.Args[2:])

		if *instanceID == "" || *startTime == "" || *stopTime == "" || *timezone == "" || *awsRegion == "" {
			createCmd.PrintDefaults()
			os.Exit(1)
		}

		createSchedule(*instanceID, *startTime, *stopTime, *timezone, *awsRegion)

	case "list-schedules":
		listSchedules()

	case "delete-schedule":
		deleteCmd := flag.NewFlagSet("delete-schedule", flag.ExitOnError)
		instanceID := deleteCmd.String("instance-id", "", "The ID of the instance to delete")
		deleteCmd.Parse(os.Args[2:])

		if *instanceID == "" {
			deleteCmd.PrintDefaults()
			os.Exit(1)
		}

		deleteSchedule(*instanceID)

	default:
		fmt.Println("expected 'create-schedule', 'list-schedules' or 'delete-schedule' subcommands")
		os.Exit(1)
	}
}
