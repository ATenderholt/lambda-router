package types

import (
	"context"
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

func (eventSource EventSource) ToCreateEventSourceMappingOutput(ctx context.Context) lambda.CreateEventSourceMappingOutput {
	id := eventSource.UUID.String()
	lastModified := time.UnixMilli(eventSource.LastModified)
	state := "Enabled"

	return lambda.CreateEventSourceMappingOutput{
		BatchSize:                      &eventSource.BatchSize,
		BisectBatchOnFunctionError:     nil,
		DestinationConfig:              nil,
		EventSourceArn:                 &eventSource.Arn,
		FilterCriteria:                 nil,
		FunctionArn:                    eventSource.Function.GetArn(ctx),
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

func (eventSource EventSource) ToGetEventSourceMappingOutput(ctx context.Context) lambda.GetEventSourceMappingOutput {
	id := eventSource.UUID.String()
	lastModified := time.UnixMilli(eventSource.LastModified)
	state := "Enabled"

	return lambda.GetEventSourceMappingOutput{
		BatchSize:                      &eventSource.BatchSize,
		BisectBatchOnFunctionError:     nil,
		DestinationConfig:              nil,
		EventSourceArn:                 &eventSource.Arn,
		FilterCriteria:                 nil,
		FunctionArn:                    eventSource.Function.GetArn(ctx),
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
