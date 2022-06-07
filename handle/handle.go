package handle

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/alkaid/behavior/logger"
	"go.uber.org/zap"

	"github.com/pkg/errors"
)

// Handler represents a message.Message's handler's meta information.
type Handler struct {
	Name    string
	Type    reflect.Type // low-level type of method
	Methods map[string]*reflect.Method
}
type HandlerMethod struct {
	Name   string
	Method *reflect.Method // method stub
}

func NewHandler() *Handler {
	return &Handler{
		Methods: map[string]*reflect.Method{},
	}
}

// HandlerPool ...
type HandlerPool struct {
	handlers map[string]*Handler // all handler method
}

// NewHandlerPool ...
func NewHandlerPool() *HandlerPool {
	return &HandlerPool{
		handlers: make(map[string]*Handler),
	}
}

// GetMethod 获取反射方法,首次获取会缓存
//  @receiver h
//  @param receiverName
//  @param receiver
//  @param methodName
//  @return method
//  @return err
func (h *HandlerPool) GetMethod(receiverName string, receiver any, methodName string) (method *reflect.Method, err error) {
	if nil == receiver {
		err = errors.New("receiver is nil")
		return
	}
	if "" == methodName {
		err = errors.New("method name is nil")
		return
	}
	if receiverName == "" {
		receiverName = reflect.TypeOf(receiver).Name()
	}
	handler := h.handlers[receiverName]
	if handler == nil {
		handler = &Handler{
			Name:    receiverName,
			Type:    reflect.TypeOf(receiver),
			Methods: map[string]*reflect.Method{},
		}
	} else {
		logger.Log.Debug("receiver Type has already register", zap.String("type", receiverName))
	}
	// 设置所有要注册的method的反射信息
	ok := false
	// 保证首字母大写
	methodName = strings.ToUpper(methodName[:1]) + methodName[1:]
	method, ok = handler.Methods[methodName]
	if !ok {
		mt, ok := handler.Type.MethodByName(methodName)
		if ok {
			method = &mt
			handler.Methods[methodName] = method
		} else {
			err = errors.WithStack(fmt.Errorf("cannot find method type:%s,method:%s", receiverName, methodName))
			return
		}
	} else {
		logger.Log.Debug("receiver method has already register", zap.String("type", receiverName), zap.String("method", methodName))
	}
	h.handlers[receiverName] = handler
	return
}

// ProcessHandlerMessage handle方法发射调用
//  @receiver h
//  @param receiverName 缓存中注册的名字
//  @param receiverInstance receiver的实例
//  @param receiver receiver的 reflect.Value
//  @param methodName 方法名
//  @param args 入参
//  @return rets
//  @return error
func (h *HandlerPool) ProcessHandlerMessage(receiverName string, receiverInstance any, receiver reflect.Value, methodName string, args ...any) (rets []any, err error) {
	method, err := h.GetMethod(receiverName, receiverInstance, methodName)
	if err != nil {
		return nil, err
	}
	pargs := []reflect.Value{receiver}
	for _, arg := range args {
		pargs = append(pargs, reflect.ValueOf(arg))
	}
	return Pcall(method, pargs)
}

// Pcall 带兜底的反射调用
//  @param method
//  @param args
//  @return rets
//  @return err
func Pcall(method *reflect.Method, args []reflect.Value) (rets []any, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			if s, ok := rec.(string); ok {
				err = errors.New(s)
			} else {
				err = errors.WithStack(fmt.Errorf("reflect call internal error - %s: %v", method.Name, rec))
			}
		}
	}()
	r := method.Func.Call(args)
	for _, value := range r {
		rets = append(rets, value.Interface())
	}
	return
}
