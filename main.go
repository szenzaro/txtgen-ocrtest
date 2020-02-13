package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"
)

func main() {
	homPath := flag.String("h", "", "path to homeric words xlsx file")
	paraPath := flag.String("p", "", "path to paraphrase words xlsx file")
	flag.Parse()
	fmt.Println("args", os.Args)

	if *homPath == "" || *paraPath == "" || !strings.HasSuffix(*homPath, "xlsx") || !strings.HasSuffix(*paraPath, "xlsx") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	hw, err := readWords(*homPath)
	if err != nil {
		log.Fatalln(err)
	}
	pw, err := readWords(*paraPath)
	if err != nil {
		log.Fatalln(err)
	}
	pages := getPages(hw, pw)

	for _, p := range pages {
		fmt.Println("saving page ", p.n)
		savePage(p)
	}
}

func savePage(p page) {
	_ = os.Mkdir("out", 0777)
	path := fmt.Sprint("out/", p.n, ".txt")
	fmt.Println(path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, v := range p.verses {
		w.WriteString(strings.TrimSpace(v) + "\n")
	}
	w.Flush()
}

type word struct {
	chant int
	verse int
	page  int
	first bool
	last  bool
	text  string
}

type page struct {
	n      int
	verses []string
}

func getPages(hom, para []word) []page {
	pages := []page{}
	var p page
	h := hom[:]
	ph := para[:]
	for len(h) > 0 && len(h) != 0 && len(ph) != 0 {
		if h[0].first {
			p, h, ph = getNextPage(h, ph)
			pages = append(pages, p)
		}
		if len(ph) > 0 && ph[0].first {
			p, ph, h = getNextPage(ph, h)
			pages = append(pages, p)
		}
	}
	return pages
}

func getNextPage(start, follow []word) (page, []word, []word) {
	verses := []string{}
	endpage := false
	p := start[0].page
	var verse string

	fmt.Println("Fetchig page ", p)
	for !endpage {
		verse, start, endpage = getVerse(start)
		verses = append(verses, verse)
		if endpage {
			continue
		}
		verse, follow, endpage = getVerse(follow)
		verses = append(verses, verse)
	}
	fmt.Println("Done...")
	return page{n: p, verses: verses}, start, follow
}

func getVerse(words []word) (string, []word, bool) {
	n := words[0].verse
	verseText := ""
	last := false
	ws := words[:]
	for i := range ws {
		if ws[i].verse != n {
			last = ws[i-1].last
			ws = ws[i:]
			break
		}
		verseText += ws[i].text + " "
	}
	if words[0] == ws[0] {
		ws = []word{}
		last = true
	}
	return verseText, ws, last
}

func readWords(path string) ([]word, error) {
	xlFile, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, err
	}

	data := []word{}

	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows[1:] {

			chant, err := strconv.Atoi(row.Cells[3].Value)
			if err != nil {
				return nil, err
			}

			verse, err := strconv.Atoi(row.Cells[10].Value)
			if err != nil {
				return nil, err
			}

			page, err := strconv.Atoi(row.Cells[7].Value)
			if err != nil && row.Cells[7].Value != "" {
				return nil, err
			}

			w := word{
				chant: chant,
				verse: verse,
				page:  page,
				first: row.Cells[8].Value == "d",
				last:  row.Cells[8].Value == "f",
				text:  strings.ToLower(row.Cells[15].Value),
			}

			data = append(data, w)
		}
	}

	return data, nil
}
