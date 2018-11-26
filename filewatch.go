package phly

/*
import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"sync"
)

const (
	fw_createdoutput = "created"
	fw_changedoutput = "changed"
	fw_removedoutput = "removed"
)

// filewatch watches for file changes.
type filewatch struct {
	Paths []string `json:"paths,omitempty"`
	done  chan struct{}
	wg    *sync.WaitGroup
	err   error
}

func (n *filewatch) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/filewatch", Name: "File Watch", Purpose: "Watch for file changes. NOTE: The underlying filewatcher rcan report spurious change events. Nothing currently compensates for that."}
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: fw_createdoutput, Purpose: "Notification that the path has been created."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: fw_changedoutput, Purpose: "Notification that the path has changed."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: fw_removedoutput, Purpose: "Notification that the path has been removed."})
	return descr
}

func (n *filewatch) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	return &filewatch{}, nil
}

func (n *filewatch) Run(args RunArgs, input Pins, sender PinSender) (Flow, error) {
	err := n.Close()
	if err != nil {
		return nil, err
	}
	n.done = make(chan struct{})
	n.wg = &sync.WaitGroup{}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.New("Watcher won't start")
	}

	n.wg.Add(1)
	go watchLoop(watcher, args.stop, n, sender)

	err = nil
	for _, p := range n.Paths {
		err = MergeErrors(err, watcher.Add(p))
	}
	if err != nil {
		n.Close()
		return nil, err
	}

	return Running, nil
}

func (n *filewatch) Close() error {
	if n.done != nil {
		close(n.done)
		n.done = nil
	}
	if n.wg != nil {
		n.wg.Wait()
		n.wg = nil
	}
	return nil
}

func watchLoop(watcher *fsnotify.Watcher, stop chan struct{}, n *filewatch, sender PinSender) {
	defer n.wg.Done()
	defer watcher.Close()

	for {
		select {
		case <-n.done:
			return
		case <-stop:
			return
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				filewatchSend(fw_createdoutput, event.Name, n, sender)
			} else if event.Op&fsnotify.Write == fsnotify.Write {
				filewatchSend(fw_changedoutput, event.Name, n, sender)
			} else if event.Op&fsnotify.Rename == fsnotify.Rename {
				filewatchSend(fw_changedoutput, event.Name, n, sender)
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				filewatchSend(fw_removedoutput, event.Name, n, sender)
			}
			//		case err := <-watcher.Errors:
			//			fmt.Println("error:", err)
		}
	}
}

func filewatchSend(channel, path string, n Node, sender PinSender) {
	doc := &Doc{}
	page := doc.NewPage("")
	page.AddItem(path)
	pins := NewPins()
	pins.Add(channel, doc)
	sender.SendPins(n, pins)
}
*/
