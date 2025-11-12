package connectors

import "errors"

var (
	// ErrInvalidConnector is returned when a connector doesn't implement the BaseConnector interface
	ErrInvalidConnector = errors.New("connector does not implement BaseConnector interface")

	// ErrConnectorNotFound is returned when a connector is not found in the store
	ErrConnectorNotFound = errors.New("connector not found")

	// ErrInvalidConfig is returned when a connector configuration is invalid
	ErrInvalidConfig = errors.New("invalid connector configuration")
)
