package repo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ATenderholt/lambda-router/internal/domain"
	"github.com/ATenderholt/lambda-router/pkg/database"
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type FunctionRepository struct {
	db database.Database
}

func NewFunctionRepository(db database.Database) domain.FunctionRepository {
	return FunctionRepository{db}
}

func (f FunctionRepository) GetAllLatestFunctions(ctx context.Context) ([]domain.Function, error) {
	logger.Infof("Querying for latest version of all Functions.")

	rows, err := f.db.QueryContext(
		ctx,
		`SELECT name, max(version), runtime, handler FROM lambda_function GROUP BY name`,
	)

	var results []domain.Function

	switch {
	case err == sql.ErrNoRows:
		logger.Warn("No Functions were found.")
		return results, nil
	case err != nil:
		e := Error{"unable to get all latest Functions", err}
		logger.Error(e)
		return results, e
	}

	for rows.Next() {
		var function domain.Function

		err := rows.Scan(
			&function.FunctionName,
			&function.Version,
			&function.Runtime,
			&function.Handler,
		)

		if err != nil {
			e := RowError{
				Op:   "GetAllLatestFunctions",
				Row:  len(results),
				Base: err,
			}
			logger.Error(e)
			return results, e
		}

		results = append(results, function)
	}

	logger.Infof("Found %d Functions.", len(results))
	return results, nil
}

func (f FunctionRepository) GetEnvironmentForFunction(ctx context.Context, function domain.Function) (*aws.Environment, error) {
	logger.Infof("Querying environment for Function %s.", function.FunctionName)

	variables := make(map[string]string)
	results, err := f.db.QueryContext(
		ctx,
		`SELECT key, value FROM lambda_function_environment WHERE function_id=?`,
		function.ID,
	)
	switch {
	case err == sql.ErrNoRows:
		logger.Infof("No Environment was found for Function %s", function.FunctionName)
	case err != nil:
		e := Error{"unable to get Environment for Function " + function.FunctionName, err}
		logger.Error(e)
		return nil, e
	}

	for results.Next() {
		var key, value string
		err = results.Scan(&key, &value)
		if err != nil {
			e := RowError{
				"GetEnvironmentForFunction " + function.FunctionName,
				len(variables),
				err,
			}
			logger.Error(e)
			return nil, e
		}

		variables[key] = value
	}

	return &aws.Environment{Variables: variables}, nil
}

func (f FunctionRepository) GetLayersForFunction(ctx context.Context, function domain.Function) ([]domain.LambdaLayer, error) {
	logger.Infof("Querying for Layers of Function %s.", function.FunctionName)
	var layers []domain.LambdaLayer

	rows, err := f.db.QueryContext(
		ctx,
		`SELECT layer_name, layer_version, code_size FROM lambda_function_layer AS lfl
					JOIN lambda_layer AS ll
						ON lfl.layer_name = ll.name AND lfl.layer_version = ll.version
				WHERE lfl.function_id = ?
		`,
		function.ID,
	)

	switch {
	case err == sql.ErrNoRows:
		logger.Infof("No layers were found for Function %s.", function.FunctionName)
		return layers, nil
	case err != nil:
		e := Error{"Error when querying for Layers of Function " + function.FunctionName, err}
		logger.Error(e)
		return nil, e
	}

	for rows.Next() {
		var layer domain.LambdaLayer
		err := rows.Scan(&layer.Name, &layer.Version, &layer.CodeSize)
		if err != nil {
			e := RowError{
				"GetLayersForFunction " + function.FunctionName,
				len(layers),
				err,
			}
			logger.Error(e)
			return layers, e
		}

		layers = append(layers, layer)
	}

	logger.Infof("Found %d Layers for Function %s.", len(layers), function.FunctionName)
	return layers, nil
}

func (f FunctionRepository) GetLatestFunctionByName(ctx context.Context, name string) (*domain.Function, error) {
	logger.Infof("Querying for Latest Function %s.", name)

	var function domain.Function
	err := f.db.QueryRowContext(
		ctx,
		`SELECT id, name, version, description, handler, role, dead_letter_arn, memory_size,
					runtime, timeout, code_sha256, code_size, last_modified_on
				FROM lambda_function WHERE name = ? ORDER BY version DESC LIMIT 1`,
		name,
	).Scan(
		&function.ID,
		&function.FunctionName,
		&function.Version,
		&function.Description,
		&function.Handler,
		&function.Role,
		&function.DeadLetterArn,
		&function.MemorySize,
		&function.Runtime,
		&function.Timeout,
		&function.CodeSha256,
		&function.CodeSize,
		&function.LastModified,
	)

	switch {
	case err == sql.ErrNoRows:
		logger.Infof("Found 0 instances of Function %s.", name)
		return nil, err
	case err != nil:
		e := Error{
			"unable to query for function " + name,
			err,
		}
		logger.Error(e)
		return nil, e
	}

	logger.Infof("Found Function: %+v", function)
	logger.Infof("Setting Function %s version to $LATEST", name)

	function.Version = "$LATEST"

	environment, err := f.GetEnvironmentForFunction(ctx, function)
	if err != nil {
		return nil, err
	}

	function.Environment = environment

	return &function, nil
}

func (f FunctionRepository) GetLatestVersionForFunctionName(ctx context.Context, name string) (int, error) {
	logger.Infof("Querying for Lambda Function %s", name)

	var dbName string
	var dbVersion int
	err := f.db.QueryRowContext(
		ctx,
		`SELECT name, version FROM lambda_function WHERE name = ? ORDER BY version DESC LIMIT 1`,
		name,
	).Scan(
		&dbName,
		&dbVersion,
	)

	switch {
	case err == sql.ErrNoRows:
		logger.Infof("Lambda Function %s not found, returning version = 0.", name)
		return 0, nil
	case err != nil:
		e := fmt.Errorf("error when querying function version for %s: %v", name, err)
		logger.Error(e)
		return -1, e
	}

	logger.Infof("... found version %d.", dbVersion)

	return dbVersion, nil
}

func (f FunctionRepository) GetVersionsForFunctionName(ctx context.Context, name string) ([]domain.Function, error) {
	logger.Infof("Querying for all Versions of Function %s.", name)

	var results []domain.Function
	rows, err := f.db.QueryContext(
		ctx,
		`SELECT id, name, version, description, handler, role, dead_letter_arn, memory_size,
					runtime, timeout, code_sha256, code_size, last_modified_on
				FROM lambda_function WHERE name = ?`,
		name,
	)

	if err != nil {
		e := Error{"unable to query for all versions of Function " + name, err}
		logger.Error(e)
		return nil, e
	}

	for rows.Next() {
		var function domain.Function
		err := rows.Scan(
			&function.ID,
			&function.FunctionName,
			&function.Version,
			&function.Description,
			&function.Handler,
			&function.Role,
			&function.DeadLetterArn,
			&function.MemorySize,
			&function.Runtime,
			&function.Timeout,
			&function.CodeSha256,
			&function.CodeSize,
			&function.LastModified,
		)

		if err != nil {
			progress := len(results)
			e := RowError{
				Op:   "GetVersionsForFunctionName " + name,
				Row:  progress,
				Base: err,
			}
			logger.Error(e)
			return nil, e
		}

		environment, err := f.GetEnvironmentForFunction(ctx, function)
		if err != nil {
			progress := len(results)
			e := RowError{
				"hydrate GetVersionsForFunctionName " + name,
				progress,
				err,
			}
			logger.Error(e)
			return nil, e
		}

		function.Environment = environment

		results = append(results, function)
	}

	return results, nil
}

func (f FunctionRepository) InsertFunction(ctx context.Context, function *domain.Function) (*domain.Function, error) {

	tx, err := f.db.BeginTx(ctx)
	if err != nil {
		e := Error{"unable to create transaction to insert layer " + function.FunctionName, err}
		logger.Error(e)
		return nil, e
	}

	functionId, err := tx.InsertOne(
		ctx,
		`INSERT INTO lambda_function (name, version, description, handler, role, dead_letter_arn,
					memory_size, runtime, timeout, code_sha256, code_size, last_modified_on)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		function.FunctionName,
		function.Version,
		function.Description,
		function.Handler,
		function.Role,
		function.DeadLetterArn,
		function.MemorySize,
		function.Runtime,
		function.Timeout,
		function.CodeSha256,
		function.CodeSize,
		function.LastModified,
	)

	if err != nil {
		msg := tx.Rollback("unable to insert function %s", function.FunctionName)
		e := Error{msg, err}
		logger.Error(e)
		return nil, e
	}

	layerStmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO lambda_function_layer (function_id, layer_name, layer_version) VALUES (?, ?, ?)`,
	)
	if err != nil {
		msg := tx.Rollback("unable to create statement to add layers to function %s", function.FunctionName)
		e := Error{msg, err}
		logger.Error(e)
		return nil, e
	}
	defer layerStmt.Close()

	for _, layer := range function.Layers {
		_, err := layerStmt.ExecContext(ctx, functionId, layer.Name, layer.Version)
		if err != nil {
			msg := tx.Rollback("unable to add layer %s to function %s: %v", layer.Name, function.FunctionName, err)
			e := Error{msg, err}
			logger.Error(e)
			return nil, e
		}
	}

	tagsStmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO lambda_function_tag (function_id, key, value) VALUES (?, ?, ?)`,
	)
	defer tagsStmt.Close()

	for key, value := range function.Tags {
		_, err := tagsStmt.ExecContext(ctx, functionId, key, value)
		if err != nil {
			msg := tx.Rollback("unable to add tag %s to function %s: %v", key, function.FunctionName, err)
			e := Error{msg, err}
			logger.Error(e)
			return nil, e
		}
	}

	evnStmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO lambda_function_environment (function_id, key, value) VALUES (?, ?, ?)`,
	)
	defer evnStmt.Close()

	for key, value := range function.Environment.Variables {
		_, err := evnStmt.ExecContext(ctx, functionId, key, value)
		if err != nil {
			msg := tx.Rollback("unable to add environment %s to function %s: %v", key, err)
			e := Error{msg, err}
			logger.Error(e)
			return nil, e
		}
	}

	err = tx.Commit()
	if err != nil {
		e := Error{"unable to commit when inserting function " + function.FunctionName, err}
		logger.Error(e)
		return nil, e
	}

	saved := domain.Function{
		ID:                         functionId,
		FunctionName:               function.FunctionName,
		Description:                function.Description,
		Handler:                    function.Handler,
		Role:                       function.Role,
		DeadLetterArn:              function.DeadLetterArn,
		MemorySize:                 function.MemorySize,
		Runtime:                    function.Runtime,
		Timeout:                    function.Timeout,
		CodeSha256:                 function.CodeSha256,
		CodeSize:                   function.CodeSize,
		Environment:                function.Environment,
		Tags:                       function.Tags,
		LastModified:               function.LastModified,
		LastUpdateStatus:           "",
		LastUpdateStatusReason:     nil,
		LastUpdateStatusReasonCode: "",
		Layers:                     nil,
		PackageType:                "",
		RevisionId:                 nil,
		State:                      "",
		StateReason:                nil,
		StateReasonCode:            "",
		Version:                    function.Version,
	}

	return &saved, nil
}

func (f FunctionRepository) UpsertFunctionEnvironment(ctx context.Context, function *domain.Function, environment *aws.Environment) error {
	logger.Infof("Upserting Environment for Function %s", function.FunctionName)

	adds := make(map[string]string)
	removes := make(map[string]string)
	updates := make(map[string]string)

	for key, value := range environment.Variables {
		_, exists := function.Environment.Variables[key]
		if exists {
			updates[key] = value
		} else {
			adds[key] = value
		}
	}

	for key, value := range function.Environment.Variables {
		_, exists := environment.Variables[key]
		if !exists {
			removes[key] = value
		}
	}

	if len(adds) == 0 && len(removes) == 0 && len(updates) == 0 {
		logger.Infof("No changes to Environment for Function %s", function.FunctionName)
		return nil
	}

	logger.Infof("There are %d adds, %d updates, %d removes for Function %s", len(adds), len(updates), len(removes),
		function.FunctionName)

	tx, err := f.db.BeginTx(ctx)
	if err != nil {
		e := Error{"unable to begin transaction to change Environment for Function " + function.FunctionName, err}
		logger.Error(e)
		return e
	}

	addStmt, updateStmt, removeStmt, err := prepareEnvironmentUpdateStatements(ctx, tx)
	if err != nil {
		e := Error{"unable to create statements to change Environment for Function " + function.FunctionName, err}
		logger.Error(e)
		return e
	}
	defer addStmt.Close()
	defer updateStmt.Close()
	defer removeStmt.Close()

	for key, value := range adds {
		_, err = addStmt.ExecContext(ctx, function.ID, key, value)
		if err != nil {
			msg := tx.Rollback("Unable to add Environment %s=%s to Function %s: %v", key, value, function.ID, err)
			e := Error{msg, err}
			logger.Error(e)
			return e
		}
	}

	for key, value := range updates {
		_, err = updateStmt.ExecContext(ctx, value, function.ID, key)
		if err != nil {
			msg := tx.Rollback("Unable to update Environment %s=%s for Function %s: %v", key, value, function.ID, err)
			e := Error{msg, err}
			logger.Error(e)
			return e
		}
	}

	for key, value := range removes {
		_, err = removeStmt.ExecContext(ctx, function.ID, key)
		if err != nil {
			msg := tx.Rollback("Unable to update Environment %s=%s for Function %s: %v", key, value, function.ID, err)
			e := Error{msg, err}
			logger.Error(e)
			return e
		}
	}

	err = tx.Commit()

	if err != nil {
		e := Error{"Unable to commit Environment changes for Function " + function.FunctionName, err}
		logger.Error(e)
		return e
	}

	function.Environment = environment

	logger.Infof("Finished upserting Environment for Function %s", function.FunctionName)

	return nil
}

func prepareEnvironmentUpdateStatements(ctx context.Context, tx database.Transaction) (add *sql.Stmt, update *sql.Stmt, remove *sql.Stmt, err error) {
	add, e := tx.PrepareContext(ctx, `INSERT INTO lambda_function_environment (function_id, key, value) VALUES (?, ?, ?)`)
	if e != nil {
		err = Error{"Unable to create add statement to change Environment", e}
		logger.Error(err)
		return
	}

	update, e = tx.PrepareContext(ctx, `UPDATE lambda_function_environment SET value=? WHERE function_id=? AND key=?`)
	if e != nil {
		err = Error{"Unable to create update statement to change Environment", e}
		logger.Error(err)
		return
	}

	remove, e = tx.PrepareContext(ctx, `DELETE FROM lambda_function_environment WHERE function_id=? AND key=?`)
	if e != nil {
		err = Error{"Unable to create remove statement to change Environment", e}
		logger.Error(err)
		return
	}

	return
}
