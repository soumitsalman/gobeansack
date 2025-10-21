package beansack

import log "github.com/sirupsen/logrus"

// NoError fails the process if err is non-nil. Exported for external callers.
func NoError(err error, args ...any) {
	if err != nil {
		log.WithError(err).Fatal(args...)
	}
}

// LogError logs an error with a formatted message.
func LogError(err error, msg string, args ...any) {
	if err != nil {
		log.WithError(err).Errorf(msg, args...)
	}
}

// LogWarning logs a warning with a formatted message.
func LogWarning(err error, msg string, args ...any) {
	if err != nil {
		log.WithError(err).Warningf(msg, args...)
	}
}
