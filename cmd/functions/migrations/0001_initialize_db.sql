-- +goose Up
CREATE TABLE IF NOT EXISTS lambda_runtime (
    id      integer PRIMARY KEY AUTOINCREMENT,
    name	text    NOT NULL UNIQUE
);

INSERT INTO lambda_runtime (name) VALUES
('python3.6'),
('python3.7'),
('python3.8'),
('python3.9'),
('python3.10');


CREATE TABLE IF NOT EXISTS lambda_layer (
    id           integer    PRIMARY KEY AUTOINCREMENT,
    name         text       NOT NULL,
    description  text       NOT NULL,
    version      integer    NOT NULL,
    created_on   integer    NOT NULL,
    code_size	 integer    NOT NULL,
    code_sha256  text       NOT NULL
);

CREATE TABLE IF NOT EXISTS lambda_layer_runtime (
    id					integer PRIMARY KEY AUTOINCREMENT,
    lambda_layer_id		integer,
    lambda_runtime_id	integer,
    FOREIGN KEY(lambda_layer_id) REFERENCES lambda_layer(id),
    FOREIGN	KEY(lambda_runtime_id) REFERENCES lambda_runtime(id)
);

CREATE TABLE IF NOT EXISTS lambda_function (
    id					integer PRIMARY KEY AUTOINCREMENT,
    name				text    NOT NULL,
    version				integer NOT NULL,
    description			text,
    handler				text    NOT NULL,
    role				text,
    dead_letter_arn	    text,
    memory_size			integer NOT NULL,
    runtime				text    NOT NULL,
    timeout				integer NOT NULL,
    code_sha256			text    NOT NULL,
    code_size			integer NOT NULL,
    last_modified_on	integer NOT NULL
);

CREATE TABLE IF NOT EXISTS lambda_function_environment (
    id					integer PRIMARY KEY AUTOINCREMENT,
    function_id 		integer NOT NULL,
    key					text    NOT NULL,
    value				text,
    FOREIGN KEY(function_id) REFERENCES lambda_function(id)
);

CREATE UNIQUE INDEX uk_environment ON lambda_function_environment(function_id, key);

CREATE TABLE IF NOT EXISTS lambda_function_tag (
    id					integer PRIMARY KEY AUTOINCREMENT,
    function_id 		integer NOT NULL,
    key					text    NOT NULL,
    value				text,
    FOREIGN KEY(function_id) REFERENCES lambda_function(id)
);

CREATE TABLE IF NOT EXISTS lambda_function_layer (
    id					integer PRIMARY KEY AUTOINCREMENT,
    function_id 		integer NOT NULL,
    layer_name			text    NOT NULL,
    layer_version		integer NOT NULL,
    FOREIGN KEY(function_id) REFERENCES lambda_function(id),
    FOREIGN KEY(layer_name) REFERENCES lambda_layer(name),
    FOREIGN KEY(layer_version) REFERENCES lambda_layer(version)
);

CREATE TABLE IF NOT EXISTS lambda_event_source (
    id		          integer   PRIMARY KEY AUTOINCREMENT,
    uuid              text      NOT NULL,
    enabled           integer   NOT NULL,
    arn               text      NOT NULL,
    function_id       integer   NOT NULL,
    batch_size        integer   NOT NULL,
    last_modified_on  integer   NOT NULL,
    FOREIGN KEY(function_id) REFERENCES lambda_function(id)
);

CREATE UNIQUE INDEX uk_lambda_event_source on lambda_event_source(arn, function_id);
