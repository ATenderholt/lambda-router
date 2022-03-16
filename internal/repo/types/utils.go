package types

import (
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"time"
)

func environmentOrEmpty(environment *aws.Environment) *aws.Environment {
	if environment != nil {
		return environment
	}

	emptyMap := make(map[string]string, 0)
	return &aws.Environment{Variables: emptyMap}
}

func int32OrDefault(p *int32, d int32) int32 {
	if p == nil {
		return d
	}

	return *p
}

func stringOrEmpty(p *string) string {
	if p == nil {
		return ""
	}

	return *p
}

const TimeFormat = "2006-01-02T15:04:05.999-0700"

func timeMillisToString(ms int64) string {
	return time.UnixMilli(ms).Format(TimeFormat)
}
