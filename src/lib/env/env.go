package env

import "os"

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
)

var currentEnvironment Environment = func() Environment {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		panic("No environment var is set")
	}

	switch environment {
	case "production":
		return Production
	case "development":
		return Development
	default:
		panic("Invalid environment is set")
	}
}()

func Get() Environment {
	return currentEnvironment
}
