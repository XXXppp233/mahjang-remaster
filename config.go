package main

import (
	"context"
	"encoding/json"
	"fmt"
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

type Assets struct {
	CharactersMap struct {
		Urls  []string `json:"Urls"`
		Index []struct {
			Head     map[string][]string `json:"head"`
			Critical map[string][]string `json:"critical"`
			Confirm  map[string][]string `json:"confirm"`
			Full     map[string][]string `json:"full"`
		} `json:"Index"`
	} `json:"CharactersMap"`
}

var (
	GlobalConf Config
	assets     Assets
	HOST       string
)

func init() {
	HOST = getEnv("HOST", "http://localhost:8080")
	if HOST == "" {
		return
	}
	if !strings.HasPrefix(HOST, "https://") && !strings.HasPrefix(HOST, "http://") {
		HOST = "https://" + HOST
	}
	fmt.Println("HOST:", HOST)
	ret := loadS3Assets()
	if ret != 0 {
		loadAssets()
	}
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

func loadS3Assets() int {
	S3Endpoint := getEnv("S3_ENDPOINT", "")
	S3AccessKeyID := getEnv("S3_ACCESS_KEY_ID", "")
	S3SecretAccessKey := getEnv("S3_SECRET_ACCESS_KEY", "")
	S3Region := getEnv("S3_REGION", "us-east-1")
	S3Bucket := getEnv("S3_BUCKET", "mahjong")
	ap := getEnv("S3_ACCESS_POINT", "")
	// Character storage is optional for local and test deployments.
	if strings.TrimSpace(S3Endpoint) == "" || strings.TrimSpace(ap) == "" {
		return 1
	}
	cred := credentials.NewStaticCredentialsProvider(S3AccessKeyID, S3SecretAccessKey, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(S3Region),
		config.WithCredentialsProvider(cred),
	)
	if err != nil {
		return 2
		//		panic("unable to load AWS SDK config, " + err.Error())
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
		return 3
		//		panic("unable to list S3 objects, " + err.Error())
	}

	for _, object := range result.Contents {
		key := *object.Key
		if strings.HasSuffix(key, ".webp") || strings.HasSuffix(key, ".webm") {
			GlobalConf.CharactersMap = append(GlobalConf.CharactersMap, "https://"+ap+"/"+key)
		}
	}
	return 0
}

func loadAssets() {
	file, _ := os.ReadFile("FrontEnd/dist/assets.json")
	err := json.Unmarshal(file, &assets)
	if err != nil {
		panic(err)
	}
	for u := range assets.CharactersMap.Urls {
		for org, charas := range assets.CharactersMap.Index[u].Head {
			for _, chara := range charas {
				GlobalConf.CharactersMap = append(GlobalConf.CharactersMap, HOST+"/head/"+org+"/"+chara+".webp")
			}
		}
		for org, charas := range assets.CharactersMap.Index[u].Critical {
			for _, chara := range charas {
				GlobalConf.CharactersMap = append(GlobalConf.CharactersMap, HOST+"/critical/"+org+"/"+chara+".webm")
			}
		}
		for org, charas := range assets.CharactersMap.Index[u].Confirm {
			for _, chara := range charas {
				GlobalConf.CharactersMap = append(GlobalConf.CharactersMap, HOST+"/confirm/"+org+"/"+chara+".webm")
			}
		}
		for org, charas := range assets.CharactersMap.Index[u].Full {
			for _, chara := range charas {
				GlobalConf.CharactersMap = append(GlobalConf.CharactersMap, HOST+"/full/"+org+"/"+chara+".webp")
			}
		}
	}
}
