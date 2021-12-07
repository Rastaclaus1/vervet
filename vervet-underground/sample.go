package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

var ctx context.Context

func main() {
	ctx = context.Background()
	client := getS3Client(ctx)
	makeBucket(client, ctx)
	pushFile(client, ctx)
}

func pushFile(client *s3.Client, ctx context.Context) {
	bucket := getBucketName()
	// Should make folder by naming
	folder := "my-nested/service-folder/"
	filename := folder + "timestamp_spec.json"

	var in bytes.Buffer
	var out bytes.Buffer

	bufferForS3Put := bufio.NewReader(&in)
	bufferToPushFile := bufio.NewWriter(&out)

	dummyJson := map[string]interface{}{
		"testing": 123,
	}
	dummyJson["crazy-animals"] = map[string]interface{}{
		"looney-tunes":  "Bugs",
		"acme":          "Wiley",
		"angry-beavers": "Norbert",
	}
	marshaledBytes, err := json.Marshal(dummyJson)
	if err != nil {
		log.Error().Err(err).Msg("Dummy json marshal error")
		return
	}
	bytesWritten, err := bufferToPushFile.Write(marshaledBytes)
	if err != nil || bytesWritten < 1 {
		log.Error().Err(err).Msg("Dummy json bytes buffer error")
		return
	}

	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &filename,
		Body:   bufferForS3Put,
	}

	putOutput, err := client.PutObject(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("S3 push failed")
		return
	}

	log.Info().Msgf("Pushed file %s", putOutput)
}

func getS3Client(ctx context.Context) *s3.Client {
	// localstack default, will make configurable
	awsEndpoint := "http://localstack:4566"
	awsRegion := "us-east-1"

	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if awsEndpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpoint,
				SigningRegion: awsRegion,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	credentialsProvider := credentials.NewStaticCredentialsProvider(getAccessKey(),
		getSecretKey(),
		"")

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
		config.WithEndpointResolver(customResolver),
		config.WithCredentialsProvider(credentialsProvider),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load the AWS configs")
	}

	// Create the resource client
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func makeBucket(client *s3.Client, ctx context.Context) {
	bucket := getBucketName()
	bucketInput := &s3.CreateBucketInput{
		Bucket: &bucket,
	}
	createBucket, err := client.CreateBucket(ctx, bucketInput)
	if err != nil {
		return
	}
	log.Info().Msgf("created bucket %s", createBucket)
}

func getBucketName() string {
	return "dummy-bucket"
}

func getAccessKey() string {
	return "accesskey"
}

func getSecretKey() string {
	return "secretkey"
}
