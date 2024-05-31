package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/saratkumar-yb/infra/cli/aws"
	"github.com/saratkumar-yb/infra/cli/helpers"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
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

		var provider helpers.CloudProvider
		switch cloudType {
		case "aws":
			tableName, err := helpers.GetTableName()
			if err != nil {
				log.Fatalf("ERROR: %v", err)
			}
			provider = aws.AWSProvider{TableName: tableName}
		default:
			fmt.Println("Unsupported cloud provider type. Please specify 'aws'.")
			os.Exit(1)
		}

		provider.CreateSchedule(*instanceID, *awsRegion, *startTime, *stopTime, *timezone, *friendlyName)

	case "list-schedules":
		listCmd := flag.NewFlagSet("list-schedules", flag.ExitOnError)
		listCmd.StringVar(&cloudType, "cloud-type", "aws", "Cloud provider type (aws, gcp, azure)")
		listCmd.Parse(os.Args[2:])

		var provider helpers.CloudProvider
		switch cloudType {
		case "aws":
			tableName, err := helpers.GetTableName()
			if err != nil {
				log.Fatalf("ERROR: %v", err)
			}
			provider = aws.AWSProvider{TableName: tableName}
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

		var provider helpers.CloudProvider
		switch cloudType {
		case "aws":
			tableName, err := helpers.GetTableName()
			if err != nil {
				log.Fatalf("ERROR: %v", err)
			}
			provider = aws.AWSProvider{TableName: tableName}
		default:
			fmt.Println("Unsupported cloud provider type. Please specify 'aws'.")
			os.Exit(1)
		}

		provider.DeleteSchedule(*instanceID)

	case "get-schedule":
		getCmd := flag.NewFlagSet("get-schedule", flag.ExitOnError)
		instanceID := getCmd.String("instance-id", "", "The ID of the instance")
		friendlyName := getCmd.String("friendly-name", "", "The friendly name of the instance")
		getCmd.StringVar(&cloudType, "cloud-type", "aws", "Cloud provider type (aws, gcp, azure)")
		getCmd.Parse(os.Args[2:])

		if *instanceID == "" && *friendlyName == "" {
			getCmd.PrintDefaults()
			os.Exit(1)
		}

		var provider helpers.CloudProvider
		switch cloudType {
		case "aws":
			tableName, err := helpers.GetTableName()
			if err != nil {
				log.Fatalf("ERROR: %v", err)
			}
			provider = aws.AWSProvider{TableName: tableName}
		default:
			fmt.Println("Unsupported cloud provider type. Please specify 'aws'.")
			os.Exit(1)
		}

		if *instanceID != "" {
			provider.GetScheduleByInstanceID(*instanceID)
		} else {
			provider.GetScheduleByFriendlyName(*friendlyName)
		}

	default:
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("yb_infra CLI")
	fmt.Println("Usage:")
	fmt.Println("  yb_infra [command]")
	fmt.Println("")
	fmt.Println("Available Commands:")
	fmt.Println("  create-schedule   Create a new schedule for an instance")
	fmt.Println("  list-schedules    List all schedules")
	fmt.Println("  delete-schedule   Delete a schedule by instance ID")
	fmt.Println("  get-schedule      Get a schedule by instance ID or friendly name")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --cloud-type      Cloud provider type (aws, gcp, azure) (default \"aws\")")
	fmt.Println("")
	fmt.Println("Use \"yb_infra [command] --help\" for more information about a command.")
}
