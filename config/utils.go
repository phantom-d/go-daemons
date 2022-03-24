package config

import (
	"fmt"
	"reflect"
	"time"
)

func DynamicCall(obj interface{}, fn string, params ...interface{}) (result interface{}, err error) {
	st := reflect.TypeOf(obj)
	if _, ok := st.MethodByName(fn); !ok {
		return
	}
	method := reflect.ValueOf(obj).MethodByName(fn)
	var inputs []reflect.Value
	if len(params) > 0 {
		for _, v := range params {
			inputs = append(inputs, reflect.ValueOf(v))
		}
	}
	res := method.Call(inputs)
	if res != nil {
		if len(res) > 1 {
			respErr := res[1].Interface()
			if respErr != nil {
				err = respErr.(error)
			}
		}
		result = res[0].Interface()
	}
	return
}

func FmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
