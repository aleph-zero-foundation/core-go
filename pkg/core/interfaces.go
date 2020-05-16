package core

import (
	"io"

	"github.com/rs/zerolog"
)

// Service that can be started and stopped.
type Service interface {
	Start() error
	Stop()
}

// Orderer represents a process used for ordering data into preblocks.
type Orderer interface {
	Service
	Set(DataSource)
}

// BuildOrderer using the given config and logger.
type BuildOrderer func(config io.Reader, log zerolog.Logger) (Orderer, PreblockSource, error)

// Interpreter represents a process for interpreting preblocks into blocks.
type Interpreter interface {
	Service
	Set(PreblockSource)
}

// BuildInterpreter using the given config and logger.
type BuildInterpreter func(config io.Reader, log zerolog.Logger) (Interpreter, BlockSource, error)

// Validator represents a process that accepts data and pushes it to the orderer, while waiting for blocks from the interpreter.
type Validator interface {
	Service
	Set(BlockSource)
}

// BuildValidator using the given config and logger.
type BuildValidator func(config io.Reader, log zerolog.Logger) (Validator, DataSource, error)
