package domain

import (
	"context"
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type RuntimeRepository interface {
	RuntimeExistsByName(ctx context.Context, runtime aws.Runtime) (bool, error)
	RuntimeIDsByNames(ctx context.Context, runtimes []aws.Runtime) (map[aws.Runtime]int, error)
}
