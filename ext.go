package gruby

import (
	"sync"
)

// Default extesions

var all = []any{}
var lock = sync.Mutex{}

func RegisterDefaultExtensions(grb *GRuby) error {
	instanceExtensions := grb.DefineModule("GoInstanceMethods")
	classExtensions := grb.DefineModule("GoClassMethods")

	instanceExtensions.DefineMethod("gruby_version", func(grb *GRuby, self Value) (Value, Value) {
		return Must(ToRuby(grb, "0.1.0")), nil
	}, ArgsNone())

	lock.Lock()
	all = append(all, instanceExtensions, classExtensions)
	lock.Unlock()

	grb.ObjectClass().Include(instanceExtensions)
	grb.ObjectClass().Extend(classExtensions)
	return nil
}
