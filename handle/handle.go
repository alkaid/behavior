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

func NewHandler() *Handler {
	return &Handler{
		Methods: map[string]*reflect.Method{},
	}
}

type HandlerPool struct {
	handlers map[string]*Handler // all handler method
}

func NewHandlerPool() *HandlerPool {
	return &HandlerPool{
		handlers: make(map[string]*Handler),
	}
}

// Register 注册代理类的反射信息
//  @receiver h
//  @param name
//  @param target
func (h *HandlerPool) Register(name string, target any) (err error) {
	if target == nil {
		err = errors.WithStack(fmt.Errorf("target is nil,name=%s", name))
		return err
	}
	if name == "" {
		name = reflect.TypeOf(target).Name()
	}
	handler := h.handlers[name]
	if handler != nil {
		logger.Log.Debug("target has already register", zap.String("type", name))
		return nil
	}
	handler = &Handler{
		Name:    name,
		Type:    reflect.TypeOf(target),
		Methods: map[string]*reflect.Method{},
	}
	methods := suitableHandlerMethods(handler.Type, func(s string) string {
		// 首字母转小写
		return strings.ToLower(s[:1]) + s[1:]
	})
	if len(methods) == 0 {
		err = errors.WithStack(fmt.Errorf("target[%s] do not have delegate methods,make sure you implement the function sign: func(eventType bcore.EventType, delta time.Duration) bcore.Result", name))
		return err
	}
	h.handlers[name] = handler
	return nil
}

func (h *HandlerPool) getMethod(targetName string, methodName string) *reflect.Method {
	handler := h.handlers[targetName]
	if handler == nil {
		return nil
	}
	return handler.Methods[methodName]
}

// SuitableGetMethod 获取反射方法,首次获取会缓存
//  @receiver h
//  @param receiverName
//  @param receiver
//  @param methodName
//  @return method
//  @return err
func (h *HandlerPool) SuitableGetMethod(receiverName string, receiver any, methodName string) (method *reflect.Method, err error) {
	if receiver == nil {
		err = errors.New("receiver is nil")
		return
	}
	if methodName == "" {
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
			return nil, err
		}
	} else {
		logger.Log.Debug("receiver method has already register", zap.String("type", receiverName), zap.String("method", methodName))
	}
	h.handlers[receiverName] = handler
	return method, err
}

// ProcessHandlerMessage handle方法反射调用
//  @receiver h
//  @param receiverName 缓存中注册的名字
//  @param receiverInstance receiver的实例
//  @param receiver receiver的 reflect.Value
//  @param methodName 方法名
//  @param args 入参
//  @return rets
//  @return error
func (h *HandlerPool) ProcessHandlerMessage(receiverName string, receiverInstance any, receiver reflect.Value, methodName string, args ...any) (rets []any, err error) {
	method := h.getMethod(receiverName, methodName)
	if method == nil {
		return nil, errors.New("method not found")
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
