package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

/*************************************
Sample Event Data to configure structs for 1-click
{
    "deviceEvent": {
      "buttonClicked": {
        "clickType": "SINGLE",
        "reportedTime": "2018-05-04T23:26:33.747Z"
      }
    },
    "deviceInfo": {
      "attributes": {
        "key3": "value3",
        "key1": "value1",
        "key4": "value4"
      },
      "type": "button",
      "deviceId": " G030PMXXXXXXXXXX ",
      "remainingLife": 5.00
    },
    "placementInfo": {
      "projectName": "test",
      "placementName": "myPlacement",
      "attributes": {
        "location": "Seattle",
        "equipment": "printer"
      },
      "devices": {
        "myButton": " G030PMXXXXXXXXXX "
      }
    }
  }
*************************************/

type ddbVars struct {
	region   string
	table    string
	tableKey string
}

type iotButtonEvent struct {
	DeviceInfo  iotButtonDeviceInfo  `json:"deviceInfo"`
	DeviceEvent iotButtonDeviceEvent `json:"deviceEvent"`
}

type iotButtonDeviceEvent struct {
	ButtonClicked iotButtonClicked `json:"buttonClicked"`
}

type iotButtonClicked struct {
	ClickType    string `json:"clickType"`
	ReportedTime string `json:"reportedTime"`
}

type iotButtonDeviceInfo struct {
	SerialNumber  string  `json:"deviceId"`
	RemainingLife float64 `json:"remainingLife"`
}

func handleRequest(req iotButtonEvent) (resultJSON map[string]*dynamodb.AttributeValue, err error) {
	inputTime := time.Now().UTC().Unix()
	serial := req.DeviceInfo.SerialNumber
	clickType := req.DeviceEvent.ButtonClicked.ClickType
	timestamp := req.DeviceEvent.ButtonClicked.ReportedTime
	batteryLife := req.DeviceInfo.RemainingLife
	ddbTablePrimaryKeyValue := strings.Join([]string{serial, strconv.FormatInt(inputTime, 10)}, "_")
	prettyJSON, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		fmt.Printf("Could not unmarshal json. Error:\n%v", err)
	}
	fmt.Printf("Request details:\n%s", string(prettyJSON))

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			ddbInputVars.tableKey: {
				S: aws.String(ddbTablePrimaryKeyValue),
			},
			"clickType": {
				S: aws.String(clickType),
			},
			"reportedTime": {
				S: aws.String(timestamp),
			},
			"remainingLife": {
				N: aws.String(strconv.FormatFloat(batteryLife, 'f', -1, 64)),
			},
			"deviceId": {
				S: aws.String(serial),
			},
		},
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String(ddbInputVars.table),
	}

	fmt.Printf("DDB input: \n%v", input)

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
	fmt.Printf("%v", result)
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
