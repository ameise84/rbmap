package rbmap

import "errors"

var (
	ErrorInvalid = errors.New("iterator is invalid")
	ErrorEndIter = errors.New("end iterator can't op")
)
