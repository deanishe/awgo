//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-05-23
//

package workflow

import "fmt"

func ExamplePadLeft() {
	fmt.Println(PadLeft("wow", "-", 5))
	// Output: --wow
}

func ExamplePadRight() {
	fmt.Println(PadRight("wow", "-", 5))
	// Output: wow--
}

func ExamplePad() {
	fmt.Println(Pad("wow", "-", 5))
	// Output: -wow-
}

func ExamplePad_longer() {
	fmt.Println(Pad("wow", "-", 10))
	// Output: ---wow----
}
