package model

import (
	"fmt"
)


func testErr() error {
	return fmt.Errorf("test error")
}
