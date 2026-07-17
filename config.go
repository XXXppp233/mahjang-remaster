package main

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3ObjectStorage struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
}

type Config struct {
	LoginRequire   bool
	CharactersMap  []string
	ShowIPLocation bool
}

var GlobalConf Config

func init() {

	loadS3Assets()
	// LOGIN_REQUIRE, default false
	loginRequireStr := getEnv("LOGIN_REQUIRE", "false")
	GlobalConf.LoginRequire = strings.ToLower(loginRequireStr) == "true"
	// SHOWIPLOCATION, default true
	showIPLocationStr := getEnv("SHOWIPLOCATION", "true")
	GlobalConf.ShowIPLocation = strings.ToLower(showIPLocationStr) == "true"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
func loadS3Assets() {
	S3Endpoint := getEnv("S3_ENDPOINT", "")
	S3AccessKeyID := getEnv("S3_ACCESS_KEY_ID", "")
	S3SecretAccessKey := getEnv("S3_SECRET_ACCESS_KEY", "")
	S3Region := getEnv("S3_REGION", "us-east-1")
	S3Bucket := getEnv("S3_BUCKET", "mahjong")
	ap := getEnv("S3_ACCESS_POINT", "")
	// Character storage is optional for local and test deployments.
	if strings.TrimSpace(S3Endpoint) == "" || strings.TrimSpace(ap) == "" {
		return
	}
	cred := credentials.NewStaticCredentialsProvider(S3AccessKeyID, S3SecretAccessKey, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(S3Region),
		config.WithCredentialsProvider(cred),
	)
	if err != nil {
		panic("unable to load AWS SDK config, " + err.Error())
	}
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(S3Endpoint)
		o.UsePathStyle = true
	})
	input := &s3.ListObjectsV2Input{
		Bucket: &S3Bucket,
	}
	result, err := s3Client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		panic("unable to list S3 objects, " + err.Error())
	}

	for _, object := range result.Contents {
		key := *object.Key
		if strings.HasSuffix(key, ".webp") || strings.HasSuffix(key, ".webm") {
			GlobalConf.CharactersMap = append(GlobalConf.CharactersMap, "https://"+ap+"/"+key)
		}
	}

}
