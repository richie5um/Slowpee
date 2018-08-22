package main

import (
	"fmt"

	"github.com/fatih/color"
)

func colorize(c color.Attribute, args ...interface{}) {
	color.Set(c)
	fmt.Println(args...)
	color.Unset()
}
