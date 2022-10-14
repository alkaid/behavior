package util

import (
	"reflect"
	"strconv"

	"github.com/alkaid/behavior/logger"
	gonanoid "github.com/matoous/go-nanoid"
	"go.uber.org/zap"
)

func Float(origin any) (out float64, ok bool) {
	rv := reflect.ValueOf(origin)
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out = float64(rv.Int())
		ok = true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		out = float64(rv.Uint())
		ok = true
	case reflect.Float32, reflect.Float64:
		out = rv.Float()
		ok = true
	case reflect.String:
		tmp, err := strconv.Atoi(rv.String())
		if err == nil {
			out = float64(tmp)
			ok = true
		}
	}
	return out, ok
}

const defaultNanoIDLen = 32

// NanoID 随机唯一ID like UUID
//
//	@return string
func NanoID() string {
	id, err := gonanoid.ID(defaultNanoIDLen)
	if err != nil {
		logger.Log.Error("", zap.Error(err))
	}
	return id
}
