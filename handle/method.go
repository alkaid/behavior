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
)

// isHandlerMethod decide a method is suitable handler method
func isHandlerMethod(method reflect.Method) bool {
	mt := method.Type
	// Method must be exported.
	if method.PkgPath != "" {
		return false
	}
	// Method needs two or three ins: receiver, EventType, delta
	if mt.NumIn() != 3 {
		return false
	}
	if t1 := mt.In(1); !t1.Implements(typeOfEventType) {
		return false
	}
	if t1 := mt.In(2); !t1.Implements(typeOfDuration) {
		return false
	}
	return true
}

func suitableHandlerMethods(typ reflect.Type, nameFunc func(string) string) map[string]reflect.Method {
	methods := make(map[string]reflect.Method)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mn := method.Name
		if isHandlerMethod(method) {
			// rewrite remote name
			if nameFunc != nil {
				mn = nameFunc(mn)
			}
			methods[mn] = method
		}
	}
	return methods
}
