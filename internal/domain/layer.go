package domain

import (
	"context"
	"github.com/ATenderholt/rainbow-functions/settings"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"path/filepath"
	"strconv"
	"strings"
)

type LambdaLayer struct {
	ID                 int64
	Name               string
	Version            int
	Description        string
	CreatedOn          string
	CompatibleRuntimes []types.Runtime
	CodeSize           int64
	CodeSha256         string
}

type LayerRepository interface {
	InsertLayer(ctx context.Context, layer LambdaLayer, dbRuntimes *map[types.Runtime]int) (*LambdaLayer, error)
	GetLayerByName(ctx context.Context, name string) ([]LambdaLayer, error)
	GetLayerByNameAndVersion(ctx context.Context, name string, version int) (LambdaLayer, error)
	GetLatestLayerVersionByName(ctx context.Context, name string) (int, error)
}

func (layer LambdaLayer) GetDestPath(cfg *settings.Config) string {
	fileName := strconv.Itoa(layer.Version) + ".zip"
	return filepath.Join(cfg.DataPath(), "lambda", "layers", layer.Name, fileName)
}

func (layer LambdaLayer) GetArn(cfg *settings.Config) *string {
	result := "arn:aws:lambda:" + cfg.Region + ":" + cfg.AccountNumber + ":layer:" + layer.Name
	return &result
}

func (layer LambdaLayer) GetVersionArn(cfg *settings.Config) *string {
	arn := layer.GetArn(cfg)
	result := *arn + ":" + strconv.Itoa(layer.Version)
	return &result
}

func LayerFromArn(arn string) LambdaLayer {
	parts := strings.Split(arn, ":")
	version, err := strconv.Atoi(parts[7])

	if err != nil {
		panic(err)
	}

	return LambdaLayer{
		Name:    parts[6],
		Version: version,
	}
}

func (layer LambdaLayer) ToPublishLayerVersionOutput(cfg *settings.Config) *lambda.PublishLayerVersionOutput {
	return &lambda.PublishLayerVersionOutput{
		CompatibleArchitectures: []types.Architecture{},
		CompatibleRuntimes:      layer.CompatibleRuntimes,
		Content: &types.LayerVersionContentOutput{
			CodeSize:   layer.CodeSize,
			CodeSha256: &layer.CodeSha256,
		},
		CreatedDate:     &layer.CreatedOn,
		Description:     &layer.Description,
		LayerArn:        layer.GetArn(cfg),
		LayerVersionArn: layer.GetVersionArn(cfg),
		LicenseInfo:     nil,
		Version:         int64(layer.Version),
	}
}

func (layer LambdaLayer) ToLayerVersionsListItem(cfg *settings.Config) types.LayerVersionsListItem {
	return types.LayerVersionsListItem{
		CompatibleArchitectures: []types.Architecture{},
		CompatibleRuntimes:      layer.CompatibleRuntimes,
		CreatedDate:             &layer.CreatedOn,
		Description:             &layer.Description,
		LayerVersionArn:         layer.GetVersionArn(cfg),
		LicenseInfo:             nil,
		Version:                 int64(layer.Version),
	}
}
