package handle

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// Handler represents a message.Message's handler's meta information.
type Handler struct {
	Type       reflect.Type // low-level type of method
	Method     *reflect.Method
	MethodType MethodType
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
//  @param target 插桩:该实例仅用于反射获取方法元数据,实际调用时并不会用该实例作为方法的receiver
func (h *HandlerPool) Register(name string, target any) (err error) {
	if target == nil {
		err = errors.WithStack(fmt.Errorf("target is nil,name=%s", name))
		return err
	}
	if name == "" {
		name = reflect.TypeOf(target).Name()
	}
	targetType := reflect.TypeOf(target)
	handles := suitableHandlerMethods(targetType, nil) // func(s string) string {
	// 	// 首字母转小写
	// 	return strings.ToLower(s[:1]) + s[1:]
	// },

	if len(handles) == 0 {
		err = errors.WithStack(fmt.Errorf("target[%s] do not have delegate handles,make sure you implement the function sign: func(eventType bcore.EventType, delta time.Duration) bcore.Result", name))
		return err
	}
	for methodName, handler := range handles {
		h.handlers[fmt.Sprintf("%s.%s", name, methodName)] = handler
	}
	return nil
}

func (h *HandlerPool) GetHandle(targetName string, methodName string) *Handler {
	handler := h.handlers[fmt.Sprintf("%s.%s", targetName, methodName)]
	if handler == nil {
		return nil
	}
	return handler
}

// ProcessHandlerMessage handle方法反射调用
//  @receiver h
//  @param receiverName 缓存中注册的名字
//  @param receiverInstance receiver的实例
//  @param methodName 方法名
//  @param args 入参
//  @return methodType 方法类型
//  @return rets
//  @return error
func (h *HandlerPool) ProcessHandlerMessage(receiverName string, receiver reflect.Value, methodName string, args ...any) (methodType MethodType, rets []any, err error) {
	handle := h.GetHandle(receiverName, methodName)
	if handle == nil {
		return 0, nil, errors.New("handle not found")
	}
	return h.ProcessHandler(handle, receiver, args...)
}

// ProcessHandler handle方法反射调用
//  @receiver h
//  @param handler
//  @param receiver receiver的 reflect.Value
//  @param args
//  @return methodType 方法类型
//  @return rets
//  @return err
func (h *HandlerPool) ProcessHandler(handler *Handler, receiver reflect.Value, args ...any) (methodType MethodType, rets []any, err error) {
	pargs := []reflect.Value{receiver}
	for _, arg := range args {
		pargs = append(pargs, reflect.ValueOf(arg))
	}
	rets, err = Pcall(handler.Method, pargs)
	if err != nil {
		return 0, nil, err
	}
	return handler.MethodType, rets, nil
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
