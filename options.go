package snapsdb

import "time"

type Option func(*dbOptions)

/* data storage path, do not duplicate with other libraries. default("./data") */
func WithDataPath(dataPath string) Option {
	return func(s *dbOptions) {
		s.dataPath = dataPath
	}
}

/* Data storage strategy, how long to save.  default(TimestampOf7Day) */
func WithDataRetention(value time.Duration) Option {
	return func(s *dbOptions) {
		s.retention = value
	}
}

/* When the map key is a string, the time format of the key. default("2006-01-02 15:04:05") */
func WithTimeKeyFormat(value string) Option {
	return func(s *dbOptions) {
		s.timekeyformat = value
	}
}
