package goche

// Bytes represent the array of byte types for data marshal
type Bytes []byte

// // Unmarshal the bytes to given param
// func (b Bytes) Unmarshal(item interface{}) error {
// }

// CacheInteractor represent the Cache implementation contract
type CacheInteractor interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
}
