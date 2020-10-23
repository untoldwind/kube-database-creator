package platforms

import (
	"fmt"
	"net/url"
)

type Platform interface {
	CheckExists(databaseName string) (bool, error)
	Create(databaseName string, adminUsername string, adminPassword string) error
}

func NewPlatform(databaseURLStr string) (Platform, error) {
	databaseURL, err := url.Parse(databaseURLStr)

	if err != nil {
		return nil, err
	}

	switch databaseURL.Scheme {
	case "postgres":
		return newPostgresPlatform(databaseURLStr)
	default:
		return nil, fmt.Errorf("Unsupported database type: %s", databaseURL.Scheme)
	}
}
