package sqs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ATenderholt/lambda-router/internal/domain"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"strings"
)

type Manager struct {
	eventRepo    domain.EventSourceRepository
	eventSources map[uuid.UUID]context.CancelFunc
	lambdaClient *lambda.Client
	sqsClient    *sqs.Client
}

func NewManager(cfg *settings.Config, eventRepo domain.EventSourceRepository) *Manager {
	sqsCfg := aws.Config{
		Region:                      "us-west-2",
		Credentials:                 credentials,
		EndpointResolverWithOptions: sqsEndpointResolver(cfg.SqsEndpoint),
		ClientLogMode:               0,
		DefaultsMode:                "",
		RuntimeEnvironment:          aws.RuntimeEnvironment{},
	}

	lambdaCfg := aws.Config{
		Region:                      "us-west-2",
		Credentials:                 credentials,
		EndpointResolverWithOptions: lambdaEndpointResolver(cfg.BasePort),
		ClientLogMode:               0,
		DefaultsMode:                "",
		RuntimeEnvironment:          aws.RuntimeEnvironment{},
	}

	return &Manager{
		eventRepo:    eventRepo,
		eventSources: make(map[uuid.UUID]context.CancelFunc),
		lambdaClient: lambda.NewFromConfig(lambdaCfg),
		sqsClient:    sqs.NewFromConfig(sqsCfg),
	}
}

func (m *Manager) StartEventSource(ctx context.Context, eventSource *domain.EventSource) error {
	parts := strings.Split(eventSource.Arn, ":")
	queueName := parts[5]

	logger.Infof("Starting consumption from Queue %s ...", queueName)

	runCtx, cancel := context.WithCancel(ctx)

	listQueuesOutput, err := m.sqsClient.ListQueues(ctx, &sqs.ListQueuesInput{QueueNamePrefix: &queueName})
	if err != nil {
		msg := fmt.Sprintf("Unable to list queues for %s: %v", queueName, err)
		logger.Error(msg)
		cancel()
		return errors.New(msg)
	}

	if len(listQueuesOutput.QueueUrls) != 1 {
		msg := fmt.Sprintf("Found %d queue urls for %s: %v", len(listQueuesOutput.QueueUrls), queueName, listQueuesOutput.QueueUrls)
		logger.Error(msg)
		cancel()
		return errors.New(msg)
	}

	queueUrl := listQueuesOutput.QueueUrls[0]
	receiveMessageInput := sqs.ReceiveMessageInput{
		QueueUrl:                &queueUrl,
		AttributeNames:          nil,
		MaxNumberOfMessages:     eventSource.BatchSize,
		MessageAttributeNames:   nil,
		ReceiveRequestAttemptId: nil,
		VisibilityTimeout:       0,
		WaitTimeSeconds:         1,
	}

	go func() {
		for {
			select {
			case <-runCtx.Done():
				return
			default:
				receiveMessageOutput, err := m.sqsClient.ReceiveMessage(ctx, &receiveMessageInput)
				if err != nil {
					logger.Errorf("Error: %v", err)
					continue
				}

				for _, message := range receiveMessageOutput.Messages {
					logger.Infof("Received %+v", message)
					payload, err := json.Marshal(message)
					if err != nil {
						logger.Errorf("Unable to marshal %v to bytes: %v", message, err)
					}

					input := lambda.InvokeInput{
						FunctionName:   &eventSource.Function.FunctionName,
						ClientContext:  nil,
						InvocationType: types.InvocationTypeEvent,
						Payload:        payload,
						Qualifier:      nil,
					}

					_, err = m.lambdaClient.Invoke(context.Background(), &input)
					if err != nil {
						logger.Errorf("Unable to invoke Function %s: %v", eventSource.Function.FunctionName, err)
						continue
					}

					_, err = m.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
						QueueUrl:      &queueUrl,
						ReceiptHandle: message.ReceiptHandle,
					})
					if err != nil {
						logger.Errorf("Unable to delete Message %s: %v", message.ReceiptHandle, err)
						continue
					}
				}
			}
		}
	}()

	m.eventSources[eventSource.UUID] = cancel

	return nil
}

func (m *Manager) StartAllEventSources(ctx context.Context) error {
	sources, err := m.eventRepo.GetAllEventSources(ctx)
	if err != nil {
		msg := fmt.Sprintf("Unable to start All Event Sources: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	for _, source := range sources {
		err = m.StartEventSource(ctx, &source)
		if err != nil {
			logger.Errorf("Unable to start Event Source %s", source.UUID)
		}
	}

	return nil
}
