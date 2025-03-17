package util

import "time"

func ConverterData(data string) (time.Time, error) {

	return time.Parse(time.RFC3339, data)

}