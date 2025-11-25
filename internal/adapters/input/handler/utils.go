package handler

import (
	"net/url"
	"strconv"
)

//Functions to search queryParams and transform

func getQueryInt(q url.Values, key string) *int {
	valStr := q.Get(key)
	if valStr == "" {
		return nil
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return nil
	}
	return &val
}

func getQueryFloat(q url.Values, key string) *float64 {
	valStr := q.Get(key)
	if valStr == "" {
		return nil
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return nil
	}
	return &val
}

func getQueryBool(q url.Values, key string) *bool {
	valStr := q.Get(key)
	if valStr == "" {
		return nil
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return nil
	}
	return &val
}

func getQueryString(q url.Values, key string) *string {
	valStr := q.Get(key)
	if valStr == "" {
		return nil
	}
	return &valStr
}
