package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	ddbClient    *dynamodb.DynamoDB
	ddbInputVars ddbVars
)

/*
* type MyEvent struct {
*	Name string `json:"name"`
* }
 */

/*************************************
{
  deviceInfo: {
    deviceId: 'G030PM037162UXE3',
    type: 'button',
    remainingLife: 99.05,
    attributes: {
      projectRegion: 'us-east-1',
      projectName: 'LightSwitch',
      placementName: 'BathroomLightSwitch',
      deviceTemplateName: 'DeviceType'
    }
  },
  deviceEvent: {
    buttonClicked: { clickType: 'DOUBLE', reportedTime: '2019-12-22T04:32:28.325Z' }
  },
  placementInfo: {
    projectName: 'LightSwitch',
    placementName: 'BathroomLightSwitch',
    attributes: {},
    devices: { DeviceType: 'G030PM037162UXE3' }
  }
}
*************************************/

type ddbVars struct {
	region   string
	table    string
	tableKey string
}

func handleRequest(ctx context.Context, req events.IoTButtonEvent) (resultJSON map[string]*dynamodb.AttributeValue, err error) {
	inputTime := time.Now().UTC().Unix()
	ddbTablePrimaryKeyValue := strings.Join([]string{req.SerialNumber, strconv.FormatInt(inputTime, 10)}, "_")
	fmt.Printf("Request details:\n%v", req)

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			ddbInputVars.tableKey: {
				S: aws.String(ddbTablePrimaryKeyValue),
			},
			"clickType": {
				S: aws.String(req.ClickType),
			},
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String(ddbInputVars.table),
	}

	result, err := ddbClient.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeTransactionConflictException:
				fmt.Println(dynamodb.ErrCodeTransactionConflictException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	return result.Attributes, err
}

func main() {

	ddbInputVars = ddbVars{
		region:   os.Getenv("AWS_REGION"),
		table:    os.Getenv("DDB_TABLE"),
		tableKey: os.Getenv("DDB_TABLE_KEY"),
	}
	session := session.Must(session.NewSession())

	ddbClient = dynamodb.New(session, aws.NewConfig().WithRegion(ddbInputVars.region))

	lambda.Start(handleRequest)
}
