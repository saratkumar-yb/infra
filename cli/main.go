package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/saratkumar-yb/infra/aws"
	"github.com/saratkumar-yb/infra/helpers"
)

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

		var provider helpers.CloudProvider
		switch cloudType {
		case "aws":
			provider = aws.AWSProvider{}
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
			provider = aws.AWSProvider{}
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
			provider = aws.AWSProvider{}
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
