package server

type HostResolver func(token string) (string, error)

var (
	PlainHostResolver = func(token string) (string, error) { return token, nil }
)
