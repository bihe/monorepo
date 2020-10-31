package logging

const (
	// LogFieldNameFunction identifies the function in structured logging
	LogFieldNameFunction = "function"

	// RequestIDKey identifies a reqest in structured logging
	RequestIDKey = "requestID"

	// ApplicationNameKey identifies the application in structured logging
	ApplicationNameKey = "appName"

	// HostIDKey identifies the host in structured logging
	HostIDKey = "hostID"

	// ErrorKey identifies errors
	ErrorKey = "err"
)

// LogConfig specifies logging settings for
type LogConfig struct {
	FilePath string
	LogLevel string
	Trace    TraceConfig
	// GrayLogServer defines the address of a log-aggregator using Graylog
	GrayLogServer string
}

// TraceConfig is used to correlate logging-entries
type TraceConfig struct {
	AppName string
	HostID  string
}
