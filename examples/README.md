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


fuzzy-simple
------------

Basic demonstration of using fuzzy filtering.

Displays and filters a list of subdirectories of ~/ in Alfred, and allows you to open the folders or browse them in Alfred.


fuzzy-custom
------------

Custom implementation of [`fuzzy.Interface`][fuzzy-if].

Displays and filters a list of GitHub repos tagged `alfred-workflow`.


update
------

Demonstration of how to enable your workflow to update itself from GitHub releases.

A good template for new workflows.


[installer]: https://gist.github.com/deanishe/35faae3e7f89f629a94e
[fuzzy-if]: https://godoc.org/github.com/deanishe/awgo/fuzzy#Interface
