/**
 * Auth :   liubo
 * Date :   2019/11/11 10:27
 * Comment:
 */

package proto

import "reflect"

type FunctionEvent struct {
	eventMap map[interface{}]map[uintptr]reflect.Value
}
func (e *FunctionEvent) check(obj interface{}) {
	if e.eventMap == nil {
		e.eventMap = make(map[interface{}]map[uintptr]reflect.Value)
	}
	_, ok := e.eventMap[obj]
	if !ok {
		e.eventMap[obj] = map[uintptr]reflect.Value{}
	}
}
func (e *FunctionEvent) Add(obj interface{}, callback func()) {
	e.check(obj)
	addr := reflect.ValueOf(callback)
	e.eventMap[obj][addr.Pointer()] = addr
}
func (e *FunctionEvent) Remove(obj interface{}, callback func()) {
	e.check(obj)
	addr := reflect.ValueOf(callback)
	delete(e.eventMap[obj], addr.Pointer())
	if len(e.eventMap[obj]) == 0 {
		delete(e.eventMap, obj)
	}
}
func (e *FunctionEvent) Broadcast() {
	e.check(nil)
	for _, v := range e.eventMap {
		for _, v2 := range v {
			v2.Call(nil)
		}
	}
}

