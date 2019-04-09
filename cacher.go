package goche

// CacheInteractor represent the Cache implementation contract
type CacheInteractor interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
}
