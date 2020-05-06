package main

import (
	"context"
	"fmt"
	"gocv.io/x/gocv"
)

type EdgeContext struct {
	ctx context.Context
	cancel context.CancelFunc
}

type DetectionResult struct {
	empty bool
	detectionTime int64
	labels map[string]string
	img gocv.Mat
	detections map[string]([]*BoundingBox)
}

type BoundingBox struct {
	topLeftX     int
	topLeftY     int
	bottomRightX int
	bottomRightY int
	confidence   float32
}

func (dr DetectionResult) String() string {
	ret := fmt.Sprintf("\n%v\n", dr.labels)
	format := "detection: %s\ntopLeftX: %d\ntopLeftY: %d\nbottomRightX: %d\nbottomRightY: %d\nconf: %f\n"
	for label, boxSl := range dr.detections {
		ret += label + "\n"
		for _, b := range boxSl {
			ret += fmt.Sprintf(format, label, b.topLeftX, b.topLeftY, b.bottomRightX, b.bottomRightY, b.confidence)
		}
		ret += "\n"
	}
	return ret
}

// TODO implementing this would make comparison faster in memdb for sets
type StringMapBoolIndex struct {
	Field bool
	Lowercase bool
}

//var MapType = reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf("")).Kind()
//
//func (s *StringMapBoolIndex) FromObject(obj interface{}) (bool, [][]byte, error) {
//	v := reflect.ValueOf(obj)
//	v = reflect.Indirect(v) // Dereference the pointer if any
//
//	fv := v.FieldByName(s.Field)
//	if !fv.IsValid() {
//		return false, nil, fmt.Errorf("field '%s' for %#v is invalid", s.Field, obj)
//	}
//
//	if fv.Kind() != MapType {
//		return false, nil, fmt.Errorf("field '%s' is not a map[string]string", s.Field)
//	}
//
//	length := fv.Len()
//	vals := make([][]byte, 0, length)
//	for _, key := range fv.MapKeys() {
//		k := key.String()
//		if k == "" {
//			continue
//		}
//		val := fv.MapIndex(key).String()
//
//		if s.Lowercase {
//			k = strings.ToLower(k)
//			val = strings.ToLower(val)
//		}
//
//		// Add the null character as a terminator
//		k += "\x00" + val + "\x00"
//
//		vals = append(vals, []byte(k))
//	}
//	if len(vals) == 0 {
//		return false, nil, nil
//	}
//	return true, vals, nil
//}
//
//// WARNING: Because of a bug in FromObject, this function will never return
//// a value when using the single-argument version.
//func (s *StringMapBoolIndex) FromArgs(args ...interface{}) ([]byte, error) {
//	if len(args) > 2 || len(args) == 0 {
//		return nil, fmt.Errorf("must provide one or two arguments")
//	}
//	key, ok := args[0].(string)
//	if !ok {
//		return nil, fmt.Errorf("argument must be a string: %#v", args[0])
//	}
//	if s.Lowercase {
//		key = strings.ToLower(key)
//	}
//	// Add the null character as a terminator
//	key += "\x00"
//
//	if len(args) == 2 {
//		val, ok := args[1].(string)
//		if !ok {
//			return nil, fmt.Errorf("argument must be a string: %#v", args[1])
//		}
//		if s.Lowercase {
//			val = strings.ToLower(val)
//		}
//		// Add the null character as a terminator
//		key += val + "\x00"
//	}
//
//	return []byte(key), nil
//}