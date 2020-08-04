Examples
========

Each subdirectory contains a complete, but trivial, Alfred workflow demonstrating AwGo features.

After building the executable, copy or symlink the directory to Alfred's workflow directory to try it out.

You can use [this script][installer] to simplify installing/symlinking workflows that are still in development.

If you've installed that script on your `$PATH`, you can try out the examples by running:

```sh
workflow-install -s /path/to/example
```

which will symlink the workflow to Alfred's workflow directory.


## fuzzy ##

**Alfred 4+ only**

Basic demonstration of using fuzzy filtering.

Displays and filters contents of ~/Downloads in Alfred, and allows you to open files, reveal them or browse them in Alfred.


## reading-list ##

Demonstrates customised fuzzy sorting.

The standard fuzzy sort is only concerned with match quality. This workflow has a custom implementation of [`fuzzy.Sortable`][fuzzy-if] and fuzzy filtering to keep a list of books sorted by status ("unread", "unpublished", "read").

Regular fuzzy sorting only considers match quality, so with the query
"kingkiller", the Kingkiller Chronicle series would be sorted based on where
the term "kingkiller" appears in the title, i.e. shortest title wins:

    The Doors of Stone (The Kingkiller Chronicle, #3)   [unpublished]
    The Wise Man's Fear (The Kingkiller Chronicle, #2)  [unread]
    The Name of the Wind (The Kingkiller Chronicle, #1) [read]

This custom implementation sorts by status then match quality, thus keeping
unread books before unpublished and read ones:

    The Wise Man's Fear (The Kingkiller Chronicle, #2)  [unread]
    The Doors of Stone (The Kingkiller Chronicle, #3)   [unpublished]
    The Name of the Wind (The Kingkiller Chronicle, #1) [read]


## update ##

Demonstration of how to enable your workflow to update itself from GitHub releases.

A good template for new workflows.


## workflows ##

Demonstrates AwGo's [caching API][caching].

Shows a list of repos from GitHub tagged `alfred-workflow`.


[caching]: https://godoc.org/github.com/deanishe/awgo#Cache
[installer]: https://gist.github.com/deanishe/35faae3e7f89f629a94e
[fuzzy-if]: https://godoc.org/go.deanishe.net/fuzzy#Sortable
