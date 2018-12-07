// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package fuzzy

import (
	"fmt"
)

// Contact is a very simple data model.
type Contact struct {
	Firstname string
	Lastname  string
}

// Name returns the full name of the Contact.
func (c *Contact) Name() string { return fmt.Sprintf("%s %s", c.Firstname, c.Lastname) }

// Contacts is a collection of Contact items. This is where fuzzy.Sortable
// must be implemented to enable fuzzy sorting.
type Contacts []*Contact

// Default sort.Interface methods
func (co Contacts) Len() int           { return len(co) }
func (co Contacts) Swap(i, j int)      { co[i], co[j] = co[j], co[i] }
func (co Contacts) Less(i, j int) bool { return co[i].Name() < co[j].Name() }

// Keywords implements Sortable.
// Comparisons are based on the the full name of the contact.
func (co Contacts) Keywords(i int) string { return co[i].Name() }

// Fuzzy sort contacts by name.
func ExampleSort() {
	// My imaginary friends
	var c = Contacts{
		&Contact{"Meggan", "Siering"},
		&Contact{"Seraphin", "Stracke"},
		&Contact{"Sheryll", "Steckel"},
		&Contact{"Erlene", "Vollbrecht"},
		&Contact{"Kayla", "Gumprich"},
		&Contact{"Jimmy", "Johnson"},
		&Contact{"Jimmy", "Jimson"},
		&Contact{"Mischa", "Witting"},
	}
	// Unsorted
	fmt.Println(c[0].Name())

	Sort(c, "mw")
	fmt.Println(c[0].Name())

	Sort(c, "meg")
	fmt.Println(c[0].Name())

	Sort(c, "voll")
	fmt.Println(c[0].Name())

	Sort(c, "ser")
	fmt.Println(c[0].Name())

	Sort(c, "ss")
	fmt.Println(c[0].Name())

	Sort(c, "jim")
	fmt.Println(c[0].Name())

	Sort(c, "jj")
	fmt.Println(c[0].Name())

	Sort(c, "jjo")
	fmt.Println(c[0].Name())

	Sort(c, "kg")
	fmt.Println(c[0].Name())
	// Output:
	// Meggan Siering
	// Mischa Witting
	// Meggan Siering
	// Erlene Vollbrecht
	// Seraphin Stracke
	// Sheryll Steckel
	// Jimmy Jimson
	// Jimmy Jimson
	// Jimmy Johnson
	// Kayla Gumprich
}
