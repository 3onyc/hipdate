package file

import (
	"encoding/csv"
	"errors"
	"github.com/3onyc/hipdate/shared"
	"github.com/3onyc/hipdate/sources"
	"gopkg.in/fsnotify.v1"
	"log"
	"os"
	"path"
	"sync"
)

var (
	MissingFilepathError = errors.New("path not specified")
)

type FileSource struct {
	cce chan *shared.ChangeEvent
	wg  *sync.WaitGroup
	sc  chan bool
	p   string
	lf  [][]string
	w   *fsnotify.Watcher
}

func NewFileSource(
	opt shared.OptionMap,
	cce chan *shared.ChangeEvent,
	wg *sync.WaitGroup,
	sc chan bool,
) (
	sources.Source,
	error,
) {
	p, ok := opt["path"]
	if !ok {
		return nil, MissingFilepathError
	}

	return &FileSource{
		cce: cce,
		wg:  wg,
		sc:  sc,
		p:   p,
	}, nil
}

func (fs *FileSource) eventHandler(
	cfe chan fsnotify.Event,
	ce chan error,
) {
	for {
		select {
		case e := <-ce:
			log.Println("ERROR [source:file] Watcher:", e)
		case fe := <-cfe:
			if fe.Name != fs.p {
				continue
			}

			if fe.Op&fsnotify.Write == fsnotify.Write && fe.Op&fsnotify.Chmod != fsnotify.Chmod {
				r, err := fs.processFile()
				if err != nil {
					log.Println("CRITICAL [source:file]", err)
				}

				fs.processRecords("remove", fs.lf)
				fs.processRecords("add", r)
				fs.lf = r
			}
		case <-fs.sc:
			fs.Stop()
			return
		}
	}
}

func (fs *FileSource) Stop() {
	log.Println("INFO [source:file] Stopping watcher ...")
	if err := fs.w.Remove(path.Dir(fs.p)); err != nil {
		log.Println("ERROR [source:file] watcher:", err)
	}

	fs.wg.Done()
}

func (fs *FileSource) Start() {
	fs.wg.Add(1)

	log.Println("INFO [source:file] Loading file source...")
	if err := fs.Initialise(); err != nil {
		log.Println("ERROR [source:file]", err)
	}

	log.Println("INFO [source:file] Starting watcher ...")
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("ERROR [source:file]", err)
		return
	}

	fs.w = w
	fs.w.Add(path.Dir(fs.p))

	fs.eventHandler(fs.w.Events, fs.w.Errors)
}

func (fs *FileSource) Initialise() error {
	r, err := fs.processFile()
	if err != nil {
		return err
	}
	fs.lf = r
	fs.processRecords("add", r)

	return nil
}

func (fs *FileSource) processFile() ([][]string, error) {
	f, err := os.Open(fs.p)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	rec, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (fs *FileSource) processRecords(e string, r [][]string) {
	for _, l := range r {
		h := shared.Host(l[0])
		for _, u := range l[1:] {
			ep, err := shared.NewEndpointFromUrl(u)
			if err != nil {
				log.Printf("WARN [source:file] Couldn't parse URL %s, skipping (%s)", ep, err)
				continue
			}

			ce := shared.NewChangeEvent(e, h, *ep)
			fs.cce <- ce
		}
	}
}

func init() {
	sources.SourceMap["file"] = NewFileSource
}
