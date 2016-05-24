//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//
// Created on 2016-05-24
//

package fuzzy_test

import (
	"fmt"

	"gogs.deanishe.net/deanishe/awgo/fuzzy"
)

// Contact is a very simple data model.
type Contact struct {
	Firstname string
	Lastname  string
	Email     string
}

// Name returns the full name of the Contact.
func (c *Contact) Name() string {
	return fmt.Sprintf("%s %s", c.Firstname, c.Lastname)
}

// Contacts is a collection of Contact items. This is where fuzzy.Interface
// must be implemented to enable fuzzy sorting.
type Contacts []*Contact

// Default sort.Interface methods
func (co Contacts) Len() int      { return len(co) }
func (co Contacts) Swap(i, j int) { co[i], co[j] = co[j], co[i] }

// Sort on first name
func (co Contacts) Less(i, j int) bool { return co[i].Firstname < co[j].Firstname }

// Keywords implements fuzzy.Interface. Comparisons are based on the
// the full name of the contact.
func (co Contacts) Keywords(i int) string {
	return fmt.Sprintf("%s %s", co[i].Firstname, co[i].Lastname)
}

// My imaginary friends
var contacts = Contacts{
	&Contact{"Meggan", "Siering", "meggan.siering@hotmail.de"},
	&Contact{"Dayne", "Skiles", "dayne.skiles@hotmail.de"},
	&Contact{"Seraphin", "Stracke", "seraphin.stracke@yahoo.com"},
	&Contact{"Sheryll", "Steckel", "sheryll.steckel@aol.de"},
	&Contact{"Rodrigo", "Langern", "rodrigo.langern@hotmail.de"},
	&Contact{"Kiara", "Nicolas", "kiara.nicolas@yahoo.com"},
	&Contact{"Romain", "Losekann", "romain.losekann@gmail.com"},
	&Contact{"Simmie", "Veum", "simmie.veum@gmail.com"},
	&Contact{"Anders", "Haag", "anders.haag@hotmail.de"},
	&Contact{"Karli", "Huel", "karli.huel@yahoo.com"},
	&Contact{"Laquita", "Wisozk", "laquita.wisozk@yahoo.com"},
	&Contact{"Elenora", "Trüb", "elenora.trueb@gmail.com"},
	&Contact{"Daliah", "Girschner", "daliah.girschner@gmail.com"},
	&Contact{"Jarno", "Gude", "jarno.gude@gmail.com"},
	&Contact{"Fabiola", "Schumm", "fabiola.schumm@gmail.com"},
	&Contact{"Erlene", "Vollbrecht", "erlene.vollbrecht@yahoo.com"},
	&Contact{"Kayla", "Gumprich", "kayla.gumprich@yahoo.com"},
	&Contact{"Edrie", "Legros", "edrie.legros@gmail.com"},
	&Contact{"Pearlene", "Fritsch", "pearlene.fritsch@gmail.com"},
	&Contact{"Mischa", "Witting", "mischa.witting@hotmail.de"},
}

// Fuzzy sort contacts by name.
func ExampleSort() {
	// Unsorted
	fmt.Println(contacts[0].Name())

	fuzzy.Sort(contacts, "mw")
	fmt.Println(contacts[0].Name())

	fuzzy.Sort(contacts, "meg")
	fmt.Println(contacts[0].Name())

	fuzzy.Sort(contacts, "voll")
	fmt.Println(contacts[0].Name())

	fuzzy.Sort(contacts, "ser")
	fmt.Println(contacts[0].Name())

	fuzzy.Sort(contacts, "rich")
	fmt.Println(contacts[0].Name())
	// Output:
	// Meggan Siering
	// Mischa Witting
	// Meggan Siering
	// Erlene Vollbrecht
	// Seraphin Stracke
	// Kayla Gumprich
}
