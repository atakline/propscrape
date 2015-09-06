package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"gopkg.in/xmlpath.v2"
)

// type comboOption []string
//
// //  Scanning of multiple entries with flag:
// func (o *comboOption) String() string {
// 	return fmt.Sprintf("%s", *o)
// }
// func (o *comboOption) Set(value string) error {
// 	*o = append(*o, value)
// 	return nil
// }

type entry struct {
	id   string
	data string
}

var (
	//emails    comboOption
	url       = "http://www.homegate.ch/homegate-war/view/component/marketplace/search/objectlist.seam?level1=rent&level2=search&a=s037&aa=AllBuyOrRent&incsubs=1"
	store     string
	itemquery = "//table[@id='objectList']/tbody/tr"
	idquery   = "td[@class='tdTitle']/h2/a[1]"
	dataquery = "td[@class='tdTitle']/h2/a[1]/@href"
	help      = false
)

var hgcbucket = []byte("hgcentries")

func parseArgs() {
	//flag.Var(&emails, "email", "Email addresses to notify of changes")
	flag.StringVar(&url, "url", url, "Url to scan")
	flag.StringVar(&store, "store", "store.db", "Storage file of previously scanned items")
	flag.StringVar(&itemquery, "itemquery", itemquery, "XPath query for repeating item of interest")
	flag.StringVar(&idquery, "idquery", idquery, "XPath query of ID under repeated item. The text of this entry is expected to be our ID, and the 'href' attribute the entry value.")
	flag.StringVar(&dataquery, "dataquery", dataquery, "XPath query of value under repeated item")
	flag.BoolVar(&help, "h", false, "Show this help")
	flag.BoolVar(&help, "-help", help, "Show this help")

	flag.Parse()
	if help {
		usage()
		flag.PrintDefaults()
		os.Exit(0)
	}
}

func usage() {
	fmt.Printf("Requests a property listing page from 'url' and extracts individual\n")
	fmt.Printf("entries from it. Compares entries with those present in 'store'\n")
	fmt.Printf("and emits any new ones to stdout, then stores the current listing.\n")
	fmt.Printf("\n")
	fmt.Printf("The 'itemquery' is a CSS selector that identifies a repeating\n")
	fmt.Printf("element to scrape for a description and link.\n")
	fmt.Printf("The 'idquery' is a CSS selector that identifies an element\n")
	fmt.Printf("under 'itemquery' that has a text and href, which are used \n")
	fmt.Printf("as id and link to emit.\n")
	fmt.Printf("(The current selector syntax does not support attribute and\n")
	fmt.Printf("content lookup - the above will be improved if a better library\n")
	fmt.Printf("is found.)\n\n")
}

// Uses goquery to retrieve a page and select relevant elements from it.
// Builds an entry map of the contents of this single page.
// Will fail if
func scrape() (map[string]entry, error) {

	r := make(map[string]entry)

	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to retrieve document: %v\n", err)
	}
	doc1, err := xmlpath.ParseHTML(res.Body)
	if err != nil {
		log.Fatalf("Failed to parse document: %v\n", err)
	}
	xp1, err := xmlpath.Compile(itemquery)
	if err != nil {
		log.Fatalf("Failed to parse item query: %v\n", err)
	}
	xp2, err := xmlpath.Compile(idquery)
	if err != nil {
		log.Fatalf("Failed to parse id query: %v\n", err)
	}
	xp3, err := xmlpath.Compile(dataquery)
	if err != nil {
		log.Fatalf("Failed to parse data query: %v\n", err)
	}
	topmatches := xp1.Iter(doc1)
	// TODO: add a debug mode that shows what was or was not matched
	for topmatches.Next() {
		n := topmatches.Node()
		n2 := xp2.Iter(n)
		n3 := xp3.Iter(n)
		for n2.Next() {
			sn2 := n2.Node().String()
			if !n3.Next() {
				log.Printf("No href for %s\n", sn2)
				continue
			}
			sn3 := n3.Node().String()
			//fmt.Printf("XXX %s %s\n", sn2, sn3)
			id := strings.TrimSpace(sn2)
			data := strings.TrimSpace(sn3)
			r[id] = entry{id, data}
		}
	}
	return r, nil
}

// Example of reading the contents of a boltdb bucket.
func readStored(file string) (map[string]entry, error) {
	r := make(map[string]entry)

	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(hgcbucket)
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", hgcbucket)
		}
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			// fmt.Printf("Read %v = %v\n", string(k), string(v))
			r[string(k)] = entry{string(k), string(v)}
		}
		return nil
	})
	return r, err
}

// Example of writing a boltdb bucket.
func writeEntries(file string, all map[string]entry) error {

	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(hgcbucket)
		bucket, err := tx.CreateBucketIfNotExists(hgcbucket)
		if err != nil {
			return err
		}
		for _, v := range all {
			// v.id should be same a k
			//fmt.Printf("WRiting %v = %v\n", v.id, v.data)
			err = bucket.Put([]byte(v.id), []byte(v.data))
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

// Outputs entries in news missing from olds.
// Prepends the identifier.
func emitMissing(olds, news map[string]entry, id string) {

	for k := range news {
		_, exists := olds[k]
		if !exists {
			fmt.Printf("%s%s | %s\n", id, k, news[k].data)
		}
	}
}

func main() {
	parseArgs()
	items, err := readStored(store)
	if err != nil {
		// This is ok, will happen on first run ever.
		log.Printf("Failed to read old entries: %v\n", err)
	}
	news, err := scrape()
	emitMissing(items, news, "> ")
	err = writeEntries(store, news)
	if err != nil {
		log.Fatalf("Failed to store entries: %v\n", err)
	}
}
