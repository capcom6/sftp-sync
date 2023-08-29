package config

import "strings"

type arrayValue []string

func (a *arrayValue) String() string {
	return strings.Join(*a, ",")
}

func (a *arrayValue) Set(value string) error {
	*a = append(*a, value)
	return nil
}
