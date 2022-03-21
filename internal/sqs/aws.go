package sqs

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"strconv"
)

var credentials aws.CredentialsProviderFunc = func(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{AccessKeyID: "ABC", SecretAccessKey: "EFG", CanExpire: false}, nil
}

func sqsEndpointResolver(url string) aws.EndpointResolverWithOptionsFunc {
	return func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               url,
			HostnameImmutable: true,
		}, nil
	}
}

func lambdaEndpointResolver(port int) aws.EndpointResolverWithOptionsFunc {
	return func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               "http://localhost:" + strconv.Itoa(port),
			HostnameImmutable: true,
		}, nil
	}
}
