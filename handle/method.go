// Copyright (c) nano Author and TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package handle

import (
	"reflect"
	"time"

	"github.com/alkaid/behavior/bcore"
)

var (
	typeOfDuration  = reflect.TypeOf(time.Duration(0))
	typeOfEventType = reflect.TypeOf(bcore.EventType(0))
	typeOfResult    = reflect.TypeOf(bcore.Result(0))
	typeOfBrain     = reflect.TypeOf((*bcore.IBrain)(nil)).Elem()
	typeOfError     = reflect.TypeOf((*error)(nil)).Elem()
)

// MethodType 支持的方法签名类型
type MethodType int

const (
	MtNone                   MethodType = iota // 无
	MtFullStyle                                // 完全形态签名:(receiver) func(eventType bcore.EventType, delta time.Duration) bcore.Result
	MtSimpleAction                             // 简单任务签名:(receiver) func()
	MtSimpleActionWithErr                      // 简单任务带错误返回的签名:(receiver) func() error
	MtSimpleActionWithResult                   // 简单任务带 bcore.Result 返回的签名:(receiver) func() bcore.Result
)

const (
	numInFullStyle  = 4
	numOutFullStyle = 1
	numInSimple     = 1
)

// isHandlerMethod decide a method is suitable handler method
func isHandlerMethod(method reflect.Method) MethodType {
	mt := method.Type
	// Method must be exported.
	if method.PkgPath != "" {
		return MtNone
	}
	// 需要4个入参: receiver,Brain EventType, delta
	if mt.NumIn() == numInFullStyle {
		if t1 := mt.In(1); t1.Kind() != reflect.Interface || !t1.Implements(typeOfBrain) {
			return MtNone
		}
		if t1 := mt.In(2); t1 != typeOfEventType {
			return MtNone
		}
		if t1 := mt.In(3); t1 != typeOfDuration {
			return MtNone
		}
		// 需要1个出参: Result,error
		if mt.NumOut() != numOutFullStyle {
			return MtNone
		}
		if t1 := mt.Out(0); t1 != typeOfResult {
			return MtNone
		}
		return MtFullStyle
	}
	// 需要一个入参 receiver
	if mt.NumIn() == numInSimple {
		// 需要0个出参
		if mt.NumOut() == 0 {
			return MtSimpleAction
		}
		// 需要1个出参 bcore.Result 或 error
		if mt.NumOut() == 1 {
			t1 := mt.Out(0)
			if t1 == typeOfResult {
				return MtSimpleActionWithResult
			}
			if t1.Kind() == reflect.Interface && t1.Implements(typeOfError) {
				return MtSimpleActionWithErr
			}
		}
	}
	return MtNone
}

func suitableHandlerMethods(typ reflect.Type, nameFunc func(string) string) map[string]*Handler {
	handles := make(map[string]*Handler)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mn := method.Name
		style := isHandlerMethod(method)
		if style == MtNone {
			continue
		}
		// rewrite name
		if nameFunc != nil {
			mn = nameFunc(mn)
		}
		h := &Handler{
			Type:       typ,
			Method:     &method,
			MethodType: style,
		}
		handles[mn] = h
	}
	return handles
}
