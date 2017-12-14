# bmfont
BMFont binary file reader

## Usage
```golang
package main

import (
	"io/ioutil"
	"log"

	"github.com/mogaika/bmfont"
)

func main() {
	fb, err := ioutil.ReadFile(`fnt.fnt`)
	if err != nil {
		panic(err)
	}
	f, err := bmfont.NewFontFromBuf(fb)
	if err != nil {
		panic(err)
	}
	log.Printf("%#+v", f.Info)
	log.Printf("%#+v", f.Common)
	log.Printf("%#+v", f.Pages)
	for _, ch := range f.Chars {
		log.Printf("%#+v", ch)

	}
}
```
