package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/labstack/echo/v4"
)

type SNSPublishAPI interface {
	Publish(ctx context.Context,
		params *sns.PublishInput,
		optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func PublishMessage(c context.Context, api SNSPublishAPI, input *sns.PublishInput) (*sns.PublishOutput, error) {
	return api.Publish(c, input)
}

func otelTesting(c echo.Context) error {
	contentLength := c.Request().ContentLength
	fmt.Printf("Content Length Received : %v\n", contentLength)

	bodyBuffer, _ := ioutil.ReadAll(c.Request().Body)

	fmt.Printf("Content Received : %v\n", bodyBuffer)

	topicARN := aws.String("***REMOVED***")
	
	flag.Parse()

	if topicARN == nil {
		fmt.Println("You must supply a message and topic ARN")
		fmt.Println("-m MESSAGE -t TOPIC-ARN")
		return c.String(http.StatusInternalServerError, "You must supply a message and topic ARN\n")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	cfg.Region = "us-east-1"
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sns.NewFromConfig(cfg)

   	enc := aws.String(base64.StdEncoding.EncodeToString(bodyBuffer))

	input := &sns.PublishInput{
		Message:  enc,
		TopicArn: topicARN,
	}

	result, err := PublishMessage(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error publishing the message:")
		fmt.Println(err)
		return c.String(http.StatusInternalServerError, "{\"message\":\"Internal Server Error\"} \n")
	}

	fmt.Println("Message ID: " + *result.MessageId)

	return c.String(http.StatusAccepted, "\n")

}

func main() {

	e := echo.New()

	fmt.Println("Starting the API server...")

	e.GET("/v1/traces", otelTesting)

	e.Logger.Fatal(e.Start(":4318"))
}
