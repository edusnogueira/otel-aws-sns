package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gorilla/mux"
	//"github.com/golang/protobuf/proto"
	//"google.golang.org/protobuf/encoding/protojson"

	//collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	//colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	//coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	//commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	//logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	//metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	//resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	//tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type SNSPublishAPI interface {
	Publish(ctx context.Context,
		params *sns.PublishInput,
		optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func PublishMessage(c context.Context, api SNSPublishAPI, input *sns.PublishInput) (*sns.PublishOutput, error) {
	return api.Publish(c, input)
}

func otelTesting(resp http.ResponseWriter, req *http.Request) {
	contentLength := req.ContentLength
	fmt.Printf("Content Length Received : %v\n", contentLength)

	bodyBuffer, _ := ioutil.ReadAll(req.Body)

	fmt.Printf("Content Received : %v\n", bodyBuffer)
	
	topicARN := flag.String("t", "***REMOVED***",
		"The ARN of the topic to which the user subscribes")

	flag.Parse()

	if *topicARN == "" {
		fmt.Println("You must supply a message and topic ARN")
		fmt.Println("-m MESSAGE -t TOPIC-ARN")
		return
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
		return
	}

	fmt.Println("Message ID: " + *result.MessageId)

}

func main() {

	fmt.Println("Starting the API server...")
	r := mux.NewRouter()
	r.HandleFunc("/v1/traces", otelTesting).Methods("POST")

	server := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:4318",
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
