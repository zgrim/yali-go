package yali

// Go port of Lingua::YALI v0.015
// http://search.cpan.org/perldoc?Lingua%3A%3AYALI

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Yali struct {
	DataDir   string
	Ngram     int
	Classes   []string
	Freq      map[string]map[string]float32
	ModelFile map[string]string

	Mu *sync.Mutex
}

type LangTuple struct {
	Lang  string
	Score float32
}

type LangList []*LangTuple

func (p LangList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p LangList) Len() int           { return len(p) }
func (p LangList) Less(i, j int) bool { return p[i].Score < p[j].Score }

func SortLangs(lst LangList) LangList {
	sort.Sort(sort.Reverse(lst))
	return lst
}

func New(dataDir string) *Yali {
	return &Yali{
		DataDir:   dataDir,
		Ngram:     -1,
		ModelFile: make(map[string]string, 0),
		Freq:      make(map[string]map[string]float32, 0),
		Classes:   make([]string, 0),
		Mu:        &sync.Mutex{},
	}
}

func classNameFromFile(fName string) (string, error) {

	nToks := strings.SplitN(fName, ".", -1)
	if len(nToks) != 3 || nToks[2] != "gz" || nToks[1] != "yali" {
		return "", errors.New(fmt.Sprintf("invalid file name %s -- expected $class.yali.gz", fName))
	}

	return nToks[0], nil
}

func (y *Yali) LoadAllMem() error {

	wg := sync.WaitGroup{}

	assets, err := AssetDir(y.DataDir)
	if err != nil {
		return err
	}

	defer y.ComputeClasses()

	for _, f := range assets {

		className, err := classNameFromFile(f)
		if err != nil {
			fmt.Printf("skipped MEM asset file %s :: %v", f, err)
			return nil
		}

		wg.Add(1)
		go func(class string, path string) {
			if err := y.LoadModel(class, path, true); err != nil {
				fmt.Printf("ERROR loading class %s :: %v\n", class, err)
			} else {
				fmt.Printf("loaded class %s from MEM path %s\n", class, path)
			}
			wg.Done()
		}(className, f)
	}

	wg.Wait()
	return nil
}

func (y *Yali) LoadAllFS() error {

	defer y.ComputeClasses()

	wg := sync.WaitGroup{}
	visit := func(path string, f os.FileInfo, err error) error {

		if f.IsDir() {
			return nil
		}

		className, err := classNameFromFile(f.Name())
		if err != nil {
			fmt.Printf("skipped MEM asset file %s :: %v", f, err)
			return nil
		}

		wg.Add(1)
		go func() {
			if err := y.LoadModel(className, path, false); err != nil {
				fmt.Printf("ERROR loading class %s :: %v\n", className, err)
			} else {
				fmt.Printf("loaded class %s from path %s\n", className, path)
			}
			wg.Done()
		}()

		return nil
	}

	filepath.Walk(y.DataDir, visit)
	wg.Wait()

	return nil
}

func (y *Yali) IdentifyString(data string) LangList {

	actRes := make(map[string]float32, 0)

	s := bufio.NewScanner(strings.NewReader(data))

	for s.Scan() {
		t := strings.TrimSpace(s.Text())
		if len(t) == 0 {
			continue
		}

		for i := 0; i <= len(t)-y.Ngram; i++ {
			length := y.Ngram + i
			if i > len(t) || length > len(t) {
				continue
			}
			w := t[i:length]

			if _, ok := y.Freq[w]; ok {
				for lang, v := range y.Freq[w] {
					actRes[lang] += v
				}
			}
		}
	}

	// sum scores of all classifiers
	allLangs := y.Classes

	var sum float32
	for _, l := range allLangs {
		score := float32(0)
		if _, ok := actRes[l]; ok {
			score = actRes[l]
		}
		sum += score
	}

	// normalize
	res := make([]*LangTuple, 0)
	for _, l := range allLangs {
		score := float32(0)
		if _, ok := actRes[l]; ok {
			score = actRes[l] / sum
		}
		res = append(res, &LangTuple{Lang: l, Score: score})
	}

	return SortLangs(res)
}

// recompute classes after manipulation with classes
func (y *Yali) ComputeClasses() {

	y.Mu.Lock()
	defer y.Mu.Unlock()

	y.Classes = make([]string, 0)

	for k, _ := range y.ModelFile {
		y.Classes = append(y.Classes, k)
	}

	return
}

func (y *Yali) LoadModel(class string, file string, fromMem bool) error {

	y.Mu.Lock()
	defer y.Mu.Unlock()

	if _, ok := y.ModelFile[class]; ok {
		return nil
	}

	var contents []byte
	var err error

	if fromMem {
		contents, err = readGzData(file)
	} else {
		contents, err = readGzFile(file)
	}

	if err != nil {
		return err
	}

	fields := bytes.SplitN(contents, []byte("\n"), 3)

	ngram, err := strconv.Atoi(string(bytes.TrimSpace(fields[0])))
	if err != nil {
		return err
	}

	/* fields[1] seems to be ignored by upstream
	tlfields = bytes.SplitN(fields[1], []byte(" "), -1)
	totalLine, err := strconv.Atoi(string(bytes.TrimSpace(tlfields[len(tlfields)-1])))
	if err != nil {
		return err
	}
	*/

	if y.Ngram < 0 {
		y.Ngram = ngram
	}

	if ngram != y.Ngram {
		return errors.New(
			fmt.Sprintf("incompatible model for class %s: expected %d-grams, found %d-gram", class, y.Ngram, ngram))
	}

	var sum float32
	scanner := bufio.NewScanner(bytes.NewReader(fields[2]))

	var i int = 2
	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			return err
		}

		i++
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 {
			continue
		}
		toks := strings.SplitN(line, "\t", -1)
		// XXX there some (a couple or so of) weird entries, eg: yor.yali.gz line 520
		if len(toks) != 2 {
			continue
		}

		k := strings.TrimSpace(string(toks[0]))
		v := strings.TrimSpace(string(toks[1]))

		val, err := strconv.Atoi(v)
		if err != nil {
			fmt.Printf("error in file %s line %s lineNr=%d: %v\n", file, line, i, err)
			continue
		}

		if _, ok := y.Freq[k]; !ok {
			y.Freq[k] = make(map[string]float32, 0)
		}

		y.Freq[k][class] = float32(val)
		sum += float32(val)
	}

	for word, wmap := range y.Freq {
		if _, ok := wmap[class]; ok {
			y.Freq[word][class] /= sum
		}
	}

	y.ModelFile[class] = file
	return nil
}

func (y *Yali) UnloadModel(class string) {

	y.Mu.Lock()
	defer y.Mu.Unlock()

	if _, ok := y.ModelFile[class]; !ok {
		return
	}

	y.Mu.Lock()
	defer func() {
		y.Mu.Unlock()
		y.ComputeClasses()
	}()

	delete(y.ModelFile, class)

	if len(y.Classes) == 0 {
		y.Ngram = -1
	}

	return
}

func readGzData(path string) ([]byte, error) {

	buf, err := Asset(path)
	if err != nil {
		return nil, err
	}
	bio := bytes.NewReader(buf)

	fz, err := gzip.NewReader(bio)
	if err != nil {
		return nil, err
	}
	defer fz.Close()

	s, err := ioutil.ReadAll(fz)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func readGzFile(filename string) ([]byte, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	fz, err := gzip.NewReader(fi)
	if err != nil {
		return nil, err
	}
	defer fz.Close()

	s, err := ioutil.ReadAll(fz)
	if err != nil {
		return nil, err
	}
	return s, nil
}
