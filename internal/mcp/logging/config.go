package logging

// LogConfig holds the logging configuration
type LogConfig struct {
	// Enabled controls whether logging is enabled
	Enabled bool

	// CallLogPath is the path to the call log file
	CallLogPath string

	// RESTErrorLogPath is the path to the REST error log file
	RESTErrorLogPath string

	// ToolErrorLogPath is the path to the tool error log file
	ToolErrorLogPath string
}

// DefaultLogConfig returns the default logging configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Enabled:          false,
		CallLogPath:      "calls.log",
		RESTErrorLogPath: "rest_errors.log",
		ToolErrorLogPath: "tool_errors.log",
	}
}
