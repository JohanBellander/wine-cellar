package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var (
	client     *s3.Client
	bucketName string
	publicURL  string
)

// Init initializes the R2 client
func Init() {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName = os.Getenv("R2_BUCKET_NAME")
	publicURL = os.Getenv("R2_PUBLIC_URL") // e.g., https://images.yourdomain.com

	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		log.Println("R2 storage not configured - image uploads will use base64 fallback")
		return
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Printf("Failed to load R2 config: %v", err)
		return
	}

	client = s3.NewFromConfig(cfg)
	log.Println("R2 storage initialized successfully")
}

// IsConfigured returns true if R2 is properly configured
func IsConfigured() bool {
	return client != nil && bucketName != ""
}

// UploadImage uploads an image to R2 and returns the public URL
func UploadImage(fileBytes []byte, contentType string, userID uint) (string, error) {
	if !IsConfigured() {
		return "", fmt.Errorf("R2 storage not configured")
	}

	// Generate a unique filename
	ext := ".jpg" // default
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	}
	
	filename := fmt.Sprintf("wines/%d/%s%s", userID, uuid.New().String(), ext)

	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:       aws.String(bucketName),
		Key:          aws.String(filename),
		Body:         bytes.NewReader(fileBytes),
		ContentType:  aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	// Return the public URL
	if publicURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(publicURL, "/"), filename), nil
	}
	
	// Fallback to bucket URL if no custom domain
	return fmt.Sprintf("https://%s.r2.dev/%s", bucketName, filename), nil
}

// DeleteImage deletes an image from R2
func DeleteImage(imageURL string) error {
	if !IsConfigured() || imageURL == "" {
		return nil
	}

	// Extract the key from the URL
	var key string
	if publicURL != "" && strings.HasPrefix(imageURL, publicURL) {
		key = strings.TrimPrefix(imageURL, strings.TrimSuffix(publicURL, "/")+"/")
	} else if strings.Contains(imageURL, ".r2.dev/") {
		parts := strings.SplitN(imageURL, ".r2.dev/", 2)
		if len(parts) == 2 {
			key = parts[1]
		}
	}

	if key == "" {
		return nil // Not an R2 URL, skip
	}

	_, err := client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}

	return nil
}

// GetExtensionFromContentType returns the file extension for a content type
func GetExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}

// GetContentTypeFromFilename returns the content type based on file extension
func GetContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

// UploadFromReader uploads from an io.Reader
func UploadFromReader(r io.Reader, contentType string, userID uint) (string, error) {
	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return UploadImage(fileBytes, contentType, userID)
}
