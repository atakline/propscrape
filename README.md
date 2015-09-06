#propscrape

This is a trivial (read: simple, non-idiomatic, non-exemplary) command line utility
for repeatedly checking a listing URL, grabbing any previously missing entries from
it, and emitting the new items.

This project uses
* [xmlpath][1] to run an XPath query against HTML output
* [boltdb][2] for simple persistence

##Retrieving and building

Obtain. If you want an independent project:
    
    cd myprojectdir; export GOPATH=$(pwd)
    go get github.com/atakline/propscrape

If you want this in your own GOPATH, naturally:

    cd $GOPATH
    go get github.com/atakline/propscrape


Build:

    go build -o bin/propscrape github.com/atakline/propscrape

##Running

Try:
    ./bin/propscrape -h
for information.

##Caveats

We assume that the garget URL returns HTML with a simple container 
for repeating elements that can be uniquely identified and used as 
a root for two further subqueries that pick an id and data element 
(here: description and URL).

BoltDB is used here as a simple persistence mechanism only. This is not good
usage, and makes absolutely no use of boltdb's finer features. Do not use this
as a boltdb example.

This is a single-run command line tool. No fancy concurrency or communication here.

##Crosscompiling

As per normal Go procedure, e.g.

    GOARCH=386 GOOS=linux go build -o bin/propscrape_linux_i386 github.com/atakline/propscrape

##Other

See [XPath query syntax][3].



   [1]: http://godoc.org/gopkg.in/xmlpath.v2 "xmlpath"
   [2]: https://github.com/boltdb/bolt "boltdb"
   [3]: https://msdn.microsoft.com/en-us/library/ms256086(v=vs.110).aspx "xpath syntax"
