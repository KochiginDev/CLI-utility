package storage

type Storage interface {
	Get() string
	Post() string
	Put() string
	Delete() string
}
