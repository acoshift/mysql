package mystmt

// Arg marks value as argument to replace with ? when build query
func Arg(v interface{}) interface{} {
	switch v.(type) {
	default:
		return arg{v}
	case arg:
	case notArg:
	case defaultValue:
	}
	return v
}

type arg struct {
	value interface{}
}

// NotArg marks value as non-argument
func NotArg(v interface{}) interface{} {
	if _, ok := v.(notArg); ok {
		return v
	}
	return notArg{v}
}

type notArg struct {
	value interface{}
}

// Default use for insert default value
var Default interface{} = defaultValue{}

type defaultValue struct{}
