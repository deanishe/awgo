#!/usr/bin/python
# encoding: utf-8
#
# Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
#
# MIT Licence. See http://opensource.org/licenses/MIT
#
# Created on 2018-02-13
#

"""Benchmark calling Alfred via JXA.

Test the necessity of bundling calls to Alfred.
"""

from __future__ import print_function, absolute_import

from contextlib import contextmanager
import subprocess
import sys
from time import time

# How many times to repeat each benchmark
REPS = 5
# How many values to set in Alfred
VALUES = 5

# Which techniques to benchmark
SINGLE = True
MULTI = True
MULTI_ALT = True

BUNDLE_ID = 'net.deanishe.awgo'

TPL_ONE = """\
Application('com.runningwithcrayons.Alfred').setConfiguration('{key}', {{
    toValue: '{value}',
    inWorkflow: '{bid}'
}});
"""

TPL_MANY = """\
var alfred = Application('com.runningwithcrayons.Alfred');
"""

TPL_LINE = """\
alfred.setConfiguration('{key}', {{
    toValue: '{value}',
    inWorkflow: '{bid}'
}});
"""


def log(s, *args):
    """Simple STDERR logger."""
    if args:
        s = s % args
    print(s, file=sys.stderr)


@contextmanager
def timed(title):
    """Time a section of code."""
    start = time()
    yield
    log('%s took %0.3fs', title, time() - start)


def run_script(script):
    """Execute JXA."""
    args = ['/usr/bin/osascript', '-l', 'JavaScript', '-e', script]
    subprocess.check_output(args)


def single():
    """Set variables one at a time."""
    with timed('single (%d values)' % VALUES):

        for i in range(VALUES):

            key = 'BENCH_{}'.format(i)
            value = 'VAL_SINGLE_{}'.format(i)

            script = TPL_ONE.format(key=key, value=value, bid=BUNDLE_ID)
            run_script(script)


def multiple():
    """Set variables all at once."""
    with timed('multiple (%d values)' % VALUES):

        script = [TPL_MANY]

        for i in range(VALUES):

            key = 'BENCH_{}'.format(i)
            value = 'VAL_SINGLE_{}'.format(i)

            script.append(TPL_LINE.format(key=key, value=value, bid=BUNDLE_ID))

        script = '\n'.join(script)

        run_script(script)


def multiple_alt():
    """Set variables all at once."""
    with timed('multi-alt (%d values)' % VALUES):

        script = []

        for i in range(VALUES):

            key = 'BENCH_{}'.format(i)
            value = 'VAL_SINGLE_{}'.format(i)

            script.append(TPL_ONE.format(key=key, value=value, bid=BUNDLE_ID))

        script = '\n'.join(script)

        run_script(script)


def main():
    """Run benchmarks."""
    cumone = cumall = cumalt = 0
    logs = []

    if SINGLE:
        for _ in range(REPS):
            start = time()
            single()
            cumone += time() - start

        logs.append('single: %0.1fs (%0.3fs/rep)' % (cumone, cumone / REPS))

    if MULTI:
        for _ in range(REPS):
            start = time()
            multiple()
            cumall += time() - start

        logs.append('multi: %0.1fs (%0.3fs/rep)' % (cumall, cumall / REPS))

    if MULTI_ALT:
        for _ in range(REPS):
            start = time()
            multiple_alt()
            cumalt += time() - start

        logs.append('multi-alt: %0.1fs (%0.3fs/rep)' % (cumalt, cumalt / REPS))

        log(', '.join(logs))


if __name__ == '__main__':
    main()
