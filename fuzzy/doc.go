// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

/*
Package fuzzy implements fuzzy sorting and filtering.

Sort() and Match() implement fuzzy search, e.g. "of" will match "OmniFocus"
and "got" will match "Game of Thrones".

Match() compares a query and a string, while Sort() sorts an object that
implements fuzzy.Sortable. Both return Result structs for each
compared string.


The algorithm is based on Forrest Smith's reverse engineering of Sublime
Text's search:
https://blog.forrestthewoods.com/reverse-engineering-sublime-text-s-fuzzy-match-4cffeed33fdb

It additionally strips diacritics from sort keys if the query is ASCII.
*/
package fuzzy
