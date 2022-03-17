package domain

import (
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/google/uuid"
	"time"
)

type EventSource struct {
	ID           int64
	UUID         uuid.UUID
	Enabled      bool
	Arn          string
	Function     *Function
	BatchSize    int32
	LastModified int64
}

func (eventSource EventSource) ToCreateEventSourceMappingOutput(cfg *settings.Config) lambda.CreateEventSourceMappingOutput {
	id := eventSource.UUID.String()
	lastModified := time.UnixMilli(eventSource.LastModified)
	state := "Enabled"

	return lambda.CreateEventSourceMappingOutput{
		BatchSize:                      &eventSource.BatchSize,
		BisectBatchOnFunctionError:     nil,
		DestinationConfig:              nil,
		EventSourceArn:                 &eventSource.Arn,
		FilterCriteria:                 nil,
		FunctionArn:                    eventSource.Function.GetArn(cfg),
		FunctionResponseTypes:          nil,
		LastModified:                   &lastModified,
		LastProcessingResult:           nil,
		MaximumBatchingWindowInSeconds: nil,
		MaximumRecordAgeInSeconds:      nil,
		MaximumRetryAttempts:           nil,
		ParallelizationFactor:          nil,
		Queues:                         nil,
		SelfManagedEventSource:         nil,
		SourceAccessConfigurations:     nil,
		StartingPosition:               "",
		StartingPositionTimestamp:      nil,
		State:                          &state,
		StateTransitionReason:          nil,
		Topics:                         nil,
		TumblingWindowInSeconds:        nil,
		UUID:                           &id,
	}
}

func (eventSource EventSource) ToGetEventSourceMappingOutput(cfg *settings.Config) lambda.GetEventSourceMappingOutput {
	id := eventSource.UUID.String()
	lastModified := time.UnixMilli(eventSource.LastModified)
	state := "Enabled"

	return lambda.GetEventSourceMappingOutput{
		BatchSize:                      &eventSource.BatchSize,
		BisectBatchOnFunctionError:     nil,
		DestinationConfig:              nil,
		EventSourceArn:                 &eventSource.Arn,
		FilterCriteria:                 nil,
		FunctionArn:                    eventSource.Function.GetArn(cfg),
		FunctionResponseTypes:          nil,
		LastModified:                   &lastModified,
		LastProcessingResult:           nil,
		MaximumBatchingWindowInSeconds: nil,
		MaximumRecordAgeInSeconds:      nil,
		MaximumRetryAttempts:           nil,
		ParallelizationFactor:          nil,
		Queues:                         nil,
		SelfManagedEventSource:         nil,
		SourceAccessConfigurations:     nil,
		StartingPosition:               "",
		StartingPositionTimestamp:      nil,
		State:                          &state,
		StateTransitionReason:          nil,
		Topics:                         nil,
		TumblingWindowInSeconds:        nil,
		UUID:                           &id,
	}
}
