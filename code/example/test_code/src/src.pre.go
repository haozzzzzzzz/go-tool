// +build ignore

package src

import (
	"fmt"
	// +pre if( USE_IMPORT == 1 ) then
	_ "fmt"
	// +pre elseif ( USE_IMPORT == 2 ) then
	_ "log"
	// +pre else
	_ "math"
	// +pre end
)

func SaySomething() {
	fmt.Println("hello")
}
