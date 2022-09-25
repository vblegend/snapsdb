package util

import "time"

/*
	Get the Unix timestamp of the current day at zero hour
*/
func GetUnixOfDay(t time.Time) int64 {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
}

/*
	Get the time of the current day at zero hour
*/
func GetTimeOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}
