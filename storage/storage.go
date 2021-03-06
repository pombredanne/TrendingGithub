package storage

const (
	// GreyListTTL defined the TTL of a repository in seconds: 1 month and 15 days (~45 days)
	GreyListTTL = 60 * 60 * 24 * 45
)

type Storage interface {
	NewPool(url, auth string) Pool
}

type Pool interface {
	Close() error
	Get() Connection
}

type Connection interface {
	// Close closes the connection.
	Close() error

	// MarkRepositoryAsTweeted marks a single projects as "already tweeted".
	// This information will be stored in Redis as a simple set with a TTL.
	// The timestamp of the tweet will be used as value.
	MarkRepositoryAsTweeted(projectName, score string) (bool, error)

	// IsRepositoryAlreadyTweeted checks if a project was already tweeted.
	// If it is not available
	//	a) the project was not tweeted yet
	//	b) the project ttl expired and is ready to tweet again
	IsRepositoryAlreadyTweeted(projectName string) (bool, error)
}
