package controllers

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/aws/aws-sdk-go/aws"
)

// AWS clients

const (
	bucketName          = "wetalkmore"
	region              = "eu-west-1"
	similarityThreshold = 70.0
)

var rekogClient *rekognition.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				"",
				"",
				"")),
	)
	if err != nil {
		log.Fatalf("Unable to load AWS config: %v", err)
	}
	rekogClient = rekognition.NewFromConfig(cfg)
}

func compareFaces(sourceKey, targetKey string) (float32, error) {
	input := &rekognition.CompareFacesInput{

		SourceImage: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String("wetalkmore"),
				Name:   aws.String(sourceKey),
			},
		},
		TargetImage: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String("wetalkmore"),
				Name:   aws.String(targetKey),
			},
		},
		SimilarityThreshold: aws.Float32(similarityThreshold),
	}

	result, err := rekogClient.CompareFaces(context.Background(), input)
	if err != nil {
		return 0.0, fmt.Errorf("failed to compare faces: %v", err)
	}

	if len(result.FaceMatches) > 0 {
		similarity := *result.FaceMatches[0].Similarity
		return similarity, nil // Match found above threshold
	}
	return 0.0, nil // No match
}

func compareFacesBytes(sourceImageBytes, targetImageBytes []byte) (float32, error) {
	// Call CompareFaces API
	input := &rekognition.CompareFacesInput{
		SourceImage: &types.Image{
			Bytes: sourceImageBytes,
		},
		TargetImage: &types.Image{
			Bytes: targetImageBytes,
		},
		SimilarityThreshold: aws.Float32(similarityThreshold),
	}

	result, err := rekogClient.CompareFaces(context.Background(), input)
	if err != nil {
		return 0.0, fmt.Errorf("failed to compare faces: %v", err)
	}

	if len(result.FaceMatches) > 0 {
		similarity := *result.FaceMatches[0].Similarity
		return similarity, nil // Match found above threshold
	}
	return 0.0, nil // No match
}

func detectFaces(sourceImageBytes []byte) (int, error) {
	input := &rekognition.DetectFacesInput{
		Image: &types.Image{
			Bytes: sourceImageBytes,
		},
	}

	result, err := rekogClient.DetectFaces(context.Background(), input)
	if err != nil {
		return 0, fmt.Errorf("failed to detect faces: %v", err)
	}

	return len(result.FaceDetails), nil // Number of faces detected
}
