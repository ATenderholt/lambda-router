package repo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ATenderholt/lambda-router/internal/domain"
	"github.com/ATenderholt/lambda-router/pkg/database"
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"strings"
	"time"
)

func NewLayerRepository(db database.Database) domain.LayerRepository {
	return LayerRepositoryImpl{db}
}

type LayerRepositoryImpl struct {
	db database.Database
}

func (l LayerRepositoryImpl) InsertLayer(ctx context.Context, layer domain.LambdaLayer, dbRuntimes *map[aws.Runtime]int) (*domain.LambdaLayer, error) {

	tx, err := l.db.BeginTx(ctx)
	if err != nil {
		e := Error{"unable to create transaction to add Lambda Layer " + layer.Name, err}
		logger.Error(e)
		return nil, e
	}

	logger.Infof("Inserting Lambda Layer %+v", layer)

	createdOn := time.Now()
	layerId, err := tx.InsertOne(
		ctx,
		`INSERT INTO lambda_layer (name, description, version, created_on, code_size, code_sha256)
					VALUES (?, ?, ?, ?, ?, ?)
		`,
		layer.Name,
		layer.Description,
		layer.Version,
		createdOn.UnixMilli(),
		layer.CodeSize,
		layer.CodeSha256,
	)

	if err != nil {
		e := Error{"unable to insert layer " + layer.Name, err}
		logger.Error(e)
		return nil, e
	}

	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO lambda_layer_runtime (lambda_layer_id, lambda_runtime_id)
					VALUES (?, ?)
		`,
	)

	if err != nil {
		msg := tx.Rollback("unable to prepare statement for inserting layer runtimes for %s: %v", layer.Name, err)
		e := Error{msg, err}
		logger.Error(e)
		return nil, e
	}

	for runtimeName, runtimeId := range *dbRuntimes {
		logger.Debugf("Trying to insert runtime %s for layer %s", runtimeName, layer.Name)
		_, err := stmt.ExecContext(ctx, layerId, runtimeId)
		if err != nil {
			msg := tx.Rollback("unable to insert runtime %s for layer %s: %v", runtimeName, layer.Name, err)
			e := Error{msg, err}
			logger.Error(e)
			return nil, e
		}
	}

	err = tx.Commit()
	if err != nil {
		e := Error{"unable to commit layer " + layer.Name, err}
		logger.Error(e)
		return nil, e
	}

	result := domain.LambdaLayer{
		ID:                 layerId,
		Name:               layer.Name,
		Version:            layer.Version,
		Description:        layer.Description,
		CreatedOn:          createdOn.Format(domain.TimeFormat),
		CompatibleRuntimes: layer.CompatibleRuntimes,
		CodeSize:           layer.CodeSize,
		CodeSha256:         layer.CodeSha256,
	}

	return &result, nil
}

func (l LayerRepositoryImpl) GetLayerByName(ctx context.Context, name string) ([]domain.LambdaLayer, error) {
	logger.Infof("Querying for Layer %s by Name", name)

	var results []domain.LambdaLayer
	rows, err := l.db.QueryContext(
		ctx,
		`SELECT ll.id, ll.name, ll.description, ll.version, ll.created_on, GROUP_CONCAT(r.name) AS runtimes,
    					ll.code_size, ll.code_sha256
					FROM lambda_runtime AS r
					INNER JOIN lambda_layer_runtime llr ON r.id = llr.lambda_runtime_id
					INNER JOIN lambda_layer ll ON llr.lambda_layer_id = ll.id
					WHERE ll.name = ?
					GROUP BY llr.lambda_layer_id;
		`,
		name,
	)

	switch {
	case err == sql.ErrNoRows:
		logger.Warnf("No rows were found for Layer %s", name)
		return results, nil
	case err != nil:
		e := Error{"problem querying for all versions for layer " + name, err}
		logger.Error(e)
		return nil, e
	}

	for rows.Next() {
		var result domain.LambdaLayer
		var createdOn int64
		var runtimes string
		err := rows.Scan(&result.ID, &result.Name, &result.Description, &result.Version, &createdOn, &runtimes,
			&result.CodeSize, &result.CodeSha256)

		if err != nil {
			e := RowError{
				Op:   "GetLayerByName" + name,
				Row:  len(results),
				Base: err,
			}
			logger.Error(e)
			return results, e
		}

		result.CreatedOn = time.UnixMilli(createdOn).Format("2006-01-02T15:04:05.999-0700")
		result.CompatibleRuntimes = stringToRuntimes(runtimes)

		logger.Debugf("found row when querying lambda layer %s: %+v", name, result)
		results = append(results, result)
	}

	logger.Infof("returning layers: %+v", results)
	return results, nil
}

func (l LayerRepositoryImpl) GetLayerByNameAndVersion(ctx context.Context, name string, version int) (domain.LambdaLayer, error) {
	var result domain.LambdaLayer
	var createdOn int64
	var runtimes string

	logger.Infof("Querying for Layer by Name and Version: %s / %d", name, version)

	err := l.db.QueryRowContext(
		ctx,
		`SELECT ll.id, ll.name, ll.description, ll.version, ll.created_on, GROUP_CONCAT(r.name) AS runtimes,
       				ll.code_size, ll.code_sha256
				FROM lambda_runtime AS r
				INNER JOIN lambda_layer_runtime llr ON r.id = llr.lambda_runtime_id
				INNER JOIN lambda_layer ll ON llr.lambda_layer_id = ll.id
				WHERE ll.name = ? AND ll.version = ?
				GROUP BY llr.lambda_layer_id;
		`,
		name,
		version,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.Version,
		&createdOn,
		&runtimes,
		&result.CodeSize,
		&result.CodeSha256,
	)

	if err != nil {
		msg := fmt.Sprintf("unable to query for Layer %s with version %d", name, version)
		e := Error{msg, err}
		logger.Error(e)
		return result, e
	}

	result.CreatedOn = time.UnixMilli(createdOn).Format("2006-01-02T15:04:05.999-0700")
	result.CompatibleRuntimes = stringToRuntimes(runtimes)

	return result, nil
}

func stringToRuntimes(runtime string) []aws.Runtime {
	logger.Debugf("converting %s to list of runtimes", runtime)
	split := strings.Split(runtime, ",")
	runtimes := make([]aws.Runtime, len(split))
	for i, value := range split {
		runtimes[i] = aws.Runtime(value)
	}

	return runtimes
}

func (l LayerRepositoryImpl) GetLatestLayerVersionByName(ctx context.Context, name string) (int, error) {
	logger.Infof("Getting latest version number for Layer %s", name)

	var dbName sql.NullString
	var dbVersion sql.NullInt32
	err := l.db.QueryRowContext(
		ctx,
		`SELECT name, max(version) from lambda_layer where name = ?`,
		name,
	).Scan(&dbName, &dbVersion)

	if err != nil {
		e := Error{"unable to query latest layer version by name", err}
		logger.Error(e)
		return -1, e
	}

	if dbName.Valid && dbVersion.Valid {
		return int(dbVersion.Int32), nil
	}

	return -1, sql.ErrNoRows
}
