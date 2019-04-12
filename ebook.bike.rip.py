#!/usr/bin/python

from __future__ import print_function

import os
import re
import time
import urllib
import urllib2

CATEGORIES = [#"fiction", 
              "romance", "general", "contemporary", "suspense",
              "historical", "fantasy", "mystery-detective", "mystery",
              "thrillers", "literature-fiction", "juvenile-fiction",
              "science-fiction", "thriller", "erotica", "paranormal",
              "action-adventure", "crime", "literary", "contemporary-women",
              "adult", "history", "women-sleuths", "young-adult", "horror"]

BOOKS_PER_PAGE = 30

USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36'

REGEX_A = r'<a class="product-title" href="\/book\/([^\/]+)\/[^\.]+\.html">([^<]+)</a>\n<p class="product-author">by: ([^<]+)</p>'
REGEX_B = r'<a href="\/book\/([^\/]+)\/[^\.]+\.html" class="title-product">([^<]+)</a>\n<div class="author-product">by ([^<]+)</div>'

DOWNLOAD_URL = 'https://ebook.bike/download/{id}/{author}/{title}/{format}'

def get(url):
    for i in range(3):
        try:
            time.sleep(0.2)
            opener = urllib2.build_opener()
            opener.addheaders = [('User-Agent', USER_AGENT)]
            print(url)
            return opener.open(url).read()
        except:
            pass
    return ''

def get_index_page(category, page):
    url = 'https://ebook.bike/tag/{category}-{page}.html'
    return get(url.format(category=category, page=page))

def links_in_category(category):
    first_page = get_index_page(category, 1) 
    count = int(re.findall("([0-9]+) BOOKS FOUND</p>", first_page)[0])
    print('{cat}: found {count} books'.format(cat=category, count=count))
    total_pages = int(count / BOOKS_PER_PAGE) + 1
    
    def all_pages():
        yield first_page
        for page in range(2, total_pages+1):
            yield get_index_page(category, page)

    for page in all_pages():
        all_matches = re.findall(REGEX_A, page, re.MULTILINE) + \
                      [ (a, c, b) for (a, b, c) in re.findall(REGEX_B, page, re.MULTILINE)]
        if not all_matches:
            print('No matches found :( - your regexes are broken')
        for match in all_matches:
            yield match

def all_urls():
    for category in CATEGORIES:
        for id_, author, title in links_in_category(category):
            author = urllib.quote_plus(author)
            title = urllib.quote_plus(title)
            for f in ('epub', 'txt'):
                yield DOWNLOAD_URL.format(id=id_, author=author, title=title, format=f)

def dump_all_urls():
    with open('out.txt', "w") as fout:
        for url in all_urls():
            fout.write('{url}\n'.format(url=url))
            fout.flush()
            os.fsync(fout)

if __name__ == '__main__':
    dump_all_urls()
