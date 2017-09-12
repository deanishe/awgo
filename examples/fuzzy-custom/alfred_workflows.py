#!/usr/bin/env python
# encoding: utf-8
#
# Copyright (c) 2017 Dean Jackson <deanishe@deanishe.net>
#
# MIT Licence. See http://opensource.org/licenses/MIT
#
# Created on 2017-09-12
#

"""Fetch list of Alfred workflows on GitHub."""

from __future__ import print_function, absolute_import

import json
import os
import sys

import requests

# Base search URL
API_URL = 'https://api.github.com/search/repositories?per_page=100'
# What to search for
QUERY = 'topic:alfred-workflow'
# Where to save results
OUTPUT = os.path.join(os.path.dirname(__file__), 'workflows.json')
# Enable beta API features (needed for topics)
HEADERS = {'Accept': 'application/vnd.github.mercy-preview+json'}


def log(s, *args):
    """Simple STDERR logger."""
    if args:
        s = s % args
    print(s, file=sys.stderr)


def fetch():
    """Retrieve search results from GitHub."""
    results = []
    page_count = 0
    page_number = 1
    while True:
        if page_count and page_number > page_count:
            break

        log('fetching page %d ...', page_number)

        r = requests.get(API_URL, {'q': QUERY, 'page': page_number},
                         headers=HEADERS)
        log('[%s] %s', r.status_code, r.url)
        r.raise_for_status()

        data = r.json()
        page_number += 1

        # populate page_count
        if not page_count:
            total_results = data.get('total_count', 100)
            page_count = total_results / 100
            if total_results % 100:
                page_count += 1

            log('%d results on %d pages', total_results, page_count)

        # extract workflows
        for d in data.get('items', []):
            results.append(dict(
                name=d['name'],
                description=d['description'],
                owner=d['owner']['login'],
                url=d['html_url'],
                stars=d['stargazers_count'],
                topics=d.get('topics', []),
                lang=d.get('language') or '',
            ))

    return results


def main():
    """Fetch Alfred workflows from GitHub."""
    results = fetch()
    with open(OUTPUT, 'wb') as fp:
        json.dump(results, fp, indent=2, sort_keys=True)

    log('saved %d workflows to %s', len(results), OUTPUT)

if __name__ == '__main__':
    main()
