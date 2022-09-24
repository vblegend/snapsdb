package snapsdb

import "time"

type Option func(*dbOptions)

func WithDataPath(dataPath string) Option {
	return func(s *dbOptions) {
		s.dataPath = dataPath
	}
}

func WithDataRetention(value time.Duration) Option {
	return func(s *dbOptions) {
		s.retention = value
	}
}
