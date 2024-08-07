package gruby

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// This is the tag to use with structures to have settings for gruby.
const tagName = "mruby"

var (
	ErrValueMustBePointer = errors.New("result must be a pointer")
	ErrUnknownType        = errors.New("unknown type")
	ErrNonStringKeys      = errors.New("keys must be strings")
)

// Decode converts the Ruby value to a Go value.
//
// The Decode process may call Ruby code and may generate Ruby garbage,
// but it collects all of its own garbage. You don't need to GC around this.
//
// See the tests (decode_test.go) for detailed and specific examples of
// how this function decodes. Basic examples are also available here and
// in the README.
//
// For primitives, the decoding process is likely what you expect. For Ruby,
// this is booleans, strings, fixnums, and floats. These map directly to
// effectively equivalent Go types: bool, string, int, float64.
// Hash and Arrays can map directly to maps and slices in Go, and Decode
// will handle this as you expect.
//
// The only remaining data type in Go is a struct. A struct in Go can map
// to any object in Ruby. If the data in Ruby is a hash, then the struct keys
// will map directly to the hash keys. If the data in Ruby is an object, then
// one of two things will be done. First: if the object responds to the
// `to_gomruby` function, then this will be called and the resulting value
// is expected to be a Hash and will be used to decode into the struct. If
// the object does NOT respond to that function, then any struct fields will
// invoke the corresponding Ruby method to attain the value.
//
// Note that with structs you can use the `mruby` tag to specify the
// Hash key or method name to call. Example:
//
//	type Foo struct {
//	    Field string `mruby:"read_field"`
//	}
func Decode(out interface{}, v Value) error {
	// The out parameter must be a pointer since we must be
	// able to write to it.
	val := reflect.ValueOf(out)
	if val.Kind() != reflect.Ptr {
		return ErrValueMustBePointer
	}

	var d decoder
	return d.decode("root", v, val.Elem())
}

type decoder struct {
	stack []reflect.Kind
}

type decodeStructGetter func(string) (Value, error)

func (d *decoder) decode(name string, v Value, result reflect.Value) error { //nolint:cyclop
	val := result

	// If we have an interface with a valid value, we use that
	// for the check.
	if result.Kind() == reflect.Interface {
		elem := result.Elem()
		if elem.IsValid() {
			val = elem
		}
	}

	// Push current onto stack unless it is an interface.
	if val.Kind() != reflect.Interface {
		d.stack = append(d.stack, val.Kind())

		// Schedule a pop
		defer func() {
			d.stack = d.stack[:len(d.stack)-1]
		}()
	}

	switch val.Kind() {
	case reflect.Bool:
		return d.decodeBool(name, v, result)
	case reflect.Float64:
		return d.decodeFloat(name, v, result)
	case reflect.Int:
		return d.decodeInt(name, v, result)
	case reflect.Interface:
		// When we see an interface, we make our own thing
		return d.decodeInterface(name, v, result)
	case reflect.Map:
		return d.decodeMap(name, v, result)
	case reflect.Ptr:
		return d.decodePtr(name, v, result)
	case reflect.Slice:
		return d.decodeSlice(name, v, result)
	case reflect.String:
		return d.decodeString(name, v, result)
	case reflect.Struct:
		return d.decodeStruct(name, v, result)
	default:
	}

	return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, name, val.Kind())
}

func (d *decoder) decodeBool(name string, v Value, result reflect.Value) error {
	switch typ := v.Type(); typ {
	case TypeFalse:
		result.Set(reflect.ValueOf(false))
	case TypeTrue:
		result.Set(reflect.ValueOf(true))
	default:
		return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, name, typ)
	}

	return nil
}

func (d *decoder) decodeFloat(name string, v Value, result reflect.Value) error {
	switch typ := v.Type(); typ {
	case TypeFloat:
		result.Set(reflect.ValueOf(ToGo[float64](v)))
	default:
		return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, name, typ)
	}

	return nil
}

func (d *decoder) decodeInt(name string, v Value, result reflect.Value) error {
	switch typ := v.Type(); typ {
	case TypeFixnum:
		result.Set(reflect.ValueOf(ToGo[int](v)))
	case TypeString:
		v, err := strconv.ParseInt(v.String(), 0, 0)
		if err != nil {
			return fmt.Errorf("failed to decode int: %w", err)
		}

		result.SetInt(v)
	default:
		return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, name, typ)
	}

	return nil
}

func (d *decoder) decodeInterface(name string, v Value, result reflect.Value) error { //nolint:cyclop
	var set reflect.Value
	redecode := true

	switch typ := v.Type(); typ {
	case TypeHash:
		var temp map[string]interface{}
		tempVal := reflect.ValueOf(temp)
		result := reflect.MakeMap(
			reflect.MapOf(
				reflect.TypeOf(""),
				tempVal.Type().Elem()))

		set = result
	case TypeArray:
		var temp []interface{}
		tempVal := reflect.ValueOf(temp)
		result := reflect.MakeSlice(
			reflect.SliceOf(tempVal.Type().Elem()), 0, 0)
		set = result
	case TypeFalse:
		fallthrough
	case TypeTrue:
		var result bool
		set = reflect.Indirect(reflect.New(reflect.TypeOf(result)))
	case TypeFixnum:
		var result int
		set = reflect.Indirect(reflect.New(reflect.TypeOf(result)))
	case TypeFloat:
		var result float64
		set = reflect.Indirect(reflect.New(reflect.TypeOf(result)))
	case TypeString:
		set = reflect.Indirect(reflect.New(reflect.TypeOf("")))
	default:
		return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, name, typ)
	}

	// Set the result to what its supposed to be, then reset
	// result so we don't reflect into this method anymore.
	result.Set(set)

	if redecode {
		// Revisit the node so that we can use the newly instantiated
		// thing and populate it.
		if err := d.decode(name, v, result); err != nil {
			return err
		}
	}

	return nil
}

func (d *decoder) decodeMap(name string, v Value, result reflect.Value) error { //nolint:funlen,cyclop
	if v.Type() != TypeHash {
		return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, name, v.Type())
	}

	// If we have an interface, then we can address the interface,
	// but not the slice itself, so get the element but set the interface
	set := result
	if result.Kind() == reflect.Interface {
		result = result.Elem()
	}

	resultType := result.Type()
	resultElemType := resultType.Elem()
	resultKeyType := resultType.Key()
	if resultKeyType.Kind() != reflect.String {
		return fmt.Errorf("%w: name=%s", ErrNonStringKeys, name)
	}

	// Make a map if it is nil
	resultMap := result
	if result.IsNil() {
		resultMap = reflect.MakeMap(
			reflect.MapOf(resultKeyType, resultElemType))
	}

	// We're going to be allocating some garbage, so set the arena
	// so it is cleared properly.
	grb := v.GRuby()
	defer grb.ArenaRestore(grb.ArenaSave())

	// Get the hash of the value
	hash := ToGo[*Hash](v)
	keysRaw, err := hash.Keys()
	if err != nil {
		return err
	}
	keys := ToGo[*Array](keysRaw)

	for i := range keys.Len() {
		// Get the key and value in Ruby. This should do no allocations.
		rbKey, err := keys.Get(i)
		if err != nil {
			return err
		}

		rbVal, err := hash.Get(rbKey)
		if err != nil {
			return err
		}

		// Make the field name
		fieldName := fmt.Sprintf("%s.<entry %d>", name, i)

		// Decode the key into the key type
		keyVal := reflect.Indirect(reflect.New(resultKeyType))
		if err := d.decode(fieldName, rbKey, keyVal); err != nil {
			return err
		}

		// Decode the value
		val := reflect.Indirect(reflect.New(resultElemType))
		if err := d.decode(fieldName, rbVal, val); err != nil {
			return err
		}

		// Set the value on the map
		resultMap.SetMapIndex(keyVal, val)
	}

	// Set the final map if we can
	set.Set(resultMap)
	return nil
}

func (d *decoder) decodePtr(name string, v Value, result reflect.Value) error {
	// Create an element of the concrete (non pointer) type and decode
	// into that. Then set the value of the pointer to this type.
	resultType := result.Type()
	resultElemType := resultType.Elem()
	val := reflect.New(resultElemType)
	if err := d.decode(name, v, reflect.Indirect(val)); err != nil {
		return err
	}

	result.Set(val)
	return nil
}

func (d *decoder) decodeSlice(name string, v Value, result reflect.Value) error {
	// If we have an interface, then we can address the interface,
	// but not the slice itself, so get the element but set the interface
	set := result
	if result.Kind() == reflect.Interface {
		result = result.Elem()
	}

	// Create the slice if it isn't nil
	resultType := result.Type()
	resultElemType := resultType.Elem()
	if result.IsNil() {
		resultSliceType := reflect.SliceOf(resultElemType)
		result = reflect.MakeSlice(
			resultSliceType, 0, 0)
	}

	// Get the hash of the value
	array := ToGo[*Array](v)

	for i := range array.Len() {
		// Get the key and value in Ruby. This should do no allocations.
		rbVal, err := array.Get(i)
		if err != nil {
			return err
		}

		// Make the field name
		fieldName := fmt.Sprintf("%s[%d]", name, i)

		// Decode the value
		val := reflect.Indirect(reflect.New(resultElemType))
		if err := d.decode(fieldName, rbVal, val); err != nil {
			return err
		}

		// Append it onto the slice
		result = reflect.Append(result, val)
	}

	set.Set(result)
	return nil
}

func (d *decoder) decodeString(name string, v Value, result reflect.Value) error {
	switch typ := v.Type(); typ {
	case TypeFixnum:
		result.Set(reflect.ValueOf(
			strconv.FormatInt(int64(ToGo[int](v)), 10)).Convert(result.Type()))
	case TypeString:
		result.Set(reflect.ValueOf(v.String()).Convert(result.Type()))
	default:
		return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, name, typ)
	}

	return nil
}

func (d *decoder) decodeStruct(name string, v Value, result reflect.Value) error { //nolint:funlen,cyclop,gocognit
	var get decodeStructGetter

	// We're going to be allocating some garbage, so set the arena
	// so it is cleared properly.
	grb := v.GRuby()
	defer grb.ArenaRestore(grb.ArenaSave())

	// Depending on the type, we need to generate a getter
	switch typ := v.Type(); typ {
	case TypeHash:
		get = decodeStructHashGetter(grb, ToGo[*Hash](v))
	case TypeObject:
		get = decodeStructObjectMethods(grb, v)
	default:
		return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, name, typ)
	}

	// This slice will keep track of all the structs we'll be decoding.
	// There can be more than one struct if there are embedded structs
	// that are squashed.
	structs := make([]reflect.Value, 1, 5) //nolint:mnd
	structs[0] = result

	// Compile the list of all the fields that we're going to be decoding
	// from all the structs.
	fields := make(map[*reflect.StructField]reflect.Value)
	for len(structs) > 0 {
		structVal := structs[0]
		structs = structs[1:]

		structType := structVal.Type()
		for i := range structType.NumField() {
			fieldType := structType.Field(i)

			if fieldType.Anonymous {
				fieldKind := fieldType.Type.Kind()
				if fieldKind != reflect.Struct {
					return fmt.Errorf("%w: name=%s type=%+v", ErrUnknownType, fieldType.Name, fieldKind)
				}

				// We have an embedded field. We "squash" the fields down
				// if specified in the tag.
				squash := false
				tagParts := strings.Split(fieldType.Tag.Get(tagName), ",")
				for _, tag := range tagParts[1:] {
					if tag == "squash" {
						squash = true
						break
					}
				}

				if squash {
					structs = append(
						structs, result.FieldByName(fieldType.Name))
					continue
				}
			}

			// Normal struct field, store it away
			fields[&fieldType] = structVal.Field(i)
		}
	}

	var (
		decodedFields    = make([]string, 0, len(fields))
		decodedFieldsVal = []reflect.Value{}
		usedKeys         = make(map[string]struct{})
	)

	for fieldType, field := range fields {
		if !field.IsValid() {
			// This should never happen
			panic("field is not valid")
		}

		// If we can't set the field, then it is unexported or something,
		// and we just continue onwards.
		if !field.CanSet() {
			continue
		}

		fieldName := strings.ToLower(fieldType.Name)

		tagValue := fieldType.Tag.Get(tagName)
		tagParts := strings.SplitN(tagValue, ",", 2) //nolint:mnd
		if len(tagParts) >= 2 && tagParts[1] == "decodedFields" {
			decodedFieldsVal = append(decodedFieldsVal, field)
			continue
		}

		if tagParts[0] != "" {
			fieldName = tagParts[0]
		}

		// We move the arena for every value here so we don't
		// generate too much intermediate garbage.
		idx := grb.ArenaSave()

		// Get the Ruby string value
		value, err := get(fieldName)
		if err != nil {
			grb.ArenaRestore(idx)
			return err
		}

		// Track the used key
		usedKeys[fieldName] = struct{}{}

		// Create the field name and decode. We range over the elements
		// because we actually want the value.
		fieldName = fmt.Sprintf("%s.%s", name, fieldName)
		err = d.decode(fieldName, value, field)
		grb.ArenaRestore(idx)
		if err != nil {
			return err
		}

		decodedFields = append(decodedFields, fieldType.Name)
	}

	if len(decodedFieldsVal) > 0 {
		// Sort it so that it is deterministic
		sort.Strings(decodedFields)

		for _, v := range decodedFieldsVal {
			v.Set(reflect.ValueOf(decodedFields))
		}
	}

	return nil
}

// decodeStructHashGetter is a decodeStructGetter that reads values from
// a hash.
func decodeStructHashGetter(grb *GRuby, h *Hash) decodeStructGetter {
	return func(key string) (Value, error) {
		rbKey := ToRuby(grb, key)
		return h.Get(rbKey)
	}
}

// decodeStructObjectMethods is a decodeStructGetter that reads values from
// an object by calling methods.
func decodeStructObjectMethods(_ *GRuby, v Value) decodeStructGetter {
	return func(key string) (Value, error) {
		return v.Call(key)
	}
}
