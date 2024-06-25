package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/patrickmn/go-cache"
)

type Schedule struct {
	InstanceID string `json:"instance_id"`
	StartTime  string `json:"start_time"`
	StopTime   string `json:"stop_time"`
	Timezone   string `json:"timezone"`
	AWSRegion  string `json:"aws_region"`
}

const gracePeriod = 10 * time.Minute // Define a 10-minute grace period

var locationCache = cache.New(5*time.Minute, 10*time.Minute)

func getLocation(timezone string) (*time.Location, error) {
	if loc, found := locationCache.Get(timezone); found {
		return loc.(*time.Location), nil
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	locationCache.Set(timezone, loc, cache.DefaultExpiration)
	return loc, nil
}

func handler(ctx context.Context) {
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		log.Fatalf("ERROR: TABLE_NAME environment variable is not set")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	dynamoSvc := dynamodb.New(sess)

	// Scan DynamoDB table for schedules
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynamoSvc.Scan(input)
	if err != nil {
		log.Fatalf("ERROR: Failed to scan DynamoDB table: %v", err)
	}

	for _, item := range result.Items {
		schedule := Schedule{}
		err = dynamodbattribute.UnmarshalMap(item, &schedule)
		if err != nil {
			log.Fatalf("ERROR: Failed to unmarshal schedule: %v", err)
		}

		loc, err := getLocation(schedule.Timezone)
		if err != nil {
			log.Printf("ERROR: Failed to load location for timezone %s: %v", schedule.Timezone, err)
			continue
		}

		sess = session.Must(session.NewSession(&aws.Config{
			Region: aws.String(schedule.AWSRegion),
		}))

		ec2Svc := ec2.New(sess)

		now := time.Now().In(loc)

		// Skip weekends
		if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			fmt.Printf("Skipping instance %s in region %s because today is a weekend.\n", schedule.InstanceID, schedule.AWSRegion)
			continue
		}

		currentDate := now.Format("2006-01-02")

		startTime, err := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", currentDate, schedule.StartTime), loc)
		if err != nil {
			log.Printf("ERROR: Failed to parse start time %s: %v", schedule.StartTime, err)
			continue
		}

		stopTime, err := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", currentDate, schedule.StopTime), loc)
		if err != nil {
			log.Printf("ERROR: Failed to parse stop time %s: %v", schedule.StopTime, err)
			continue
		}

		if withinGracePeriod(now, startTime) {
			instanceState, err := getInstanceState(ec2Svc, schedule.InstanceID)
			if err != nil {
				log.Printf("ERROR: Failed to get instance state for %s: %v", schedule.InstanceID, err)
				continue
			}

			if instanceState != "running" {
				_, err := ec2Svc.StartInstances(&ec2.StartInstancesInput{
					InstanceIds: []*string{aws.String(schedule.InstanceID)},
				})
				if err != nil {
					log.Printf("ERROR: Failed to start instance %s: %v", schedule.InstanceID, err)
				} else {
					log.Printf("INFO: Started instance %s", schedule.InstanceID)
				}
			} else {
				log.Printf("INFO: Instance %s is already running", schedule.InstanceID)
			}
		} else if withinGracePeriod(now, stopTime) {
			instanceState, err := getInstanceState(ec2Svc, schedule.InstanceID)
			if err != nil {
				log.Printf("ERROR: Failed to get instance state for %s: %v", schedule.InstanceID, err)
				continue
			}

			if instanceState != "stopped" {
				_, err := ec2Svc.StopInstances(&ec2.StopInstancesInput{
					InstanceIds: []*string{aws.String(schedule.InstanceID)},
				})
				if err != nil {
					log.Printf("ERROR: Failed to stop instance %s: %v", schedule.InstanceID, err)
				} else {
					log.Printf("INFO: Stopped instance %s", schedule.InstanceID)
				}
			} else {
				log.Printf("INFO: Instance %s is already stopped", schedule.InstanceID)
			}
		}
	}
}

func withinGracePeriod(currentTime, targetTime time.Time) bool {
	return currentTime.After(targetTime.Add(-gracePeriod)) && currentTime.Before(targetTime.Add(gracePeriod))
}

func getInstanceState(ec2Svc *ec2.EC2, instanceID string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	}
	result, err := ec2Svc.DescribeInstances(input)
	if err != nil {
		return "", err
	}

	if len(result.Reservations) > 0 && len(result.Reservations[0].Instances) > 0 {
		return *result.Reservations[0].Instances[0].State.Name, nil
	}

	return "", fmt.Errorf("instance %s not found", instanceID)
}

func main() {
	lambda.Start(handler)
}
