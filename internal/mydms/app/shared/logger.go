package shared

import "github.com/go-kit/kit/log"

// Log uses structed logging with key-val for method, err and any other key-val supplied
func Log(logger log.Logger, method string, err error, keyvals ...interface{}) {
	kvs := make([]interface{}, 4)
	std := []interface{}{"method", method, "err", err}
	for _, e := range std {
		kvs = append(kvs, e)
	}
	// append the rest of the supplied params
	kvs = append(kvs, keyvals...)
	logger.Log(keyvals...)
}
