package planner

import "fmt"

func recovered(v interface{}) string {
	switch v := v.(type) {
	case error:
		return v.Error()

	case string:
		return v
	}

	return fmt.Sprintf("%+v", v)
}
