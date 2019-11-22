package storage

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/koshatul/mailhog/common/data"
)

// Maildir is a maildir storage backend
type Maildir struct {
	Path     string
	Hostname string
}

// CreateMaildir creates a new maildir storage backend
func CreateMaildir(path, hostname string) *Maildir {
	if len(path) == 0 {
		dir, err := ioutil.TempDir("", "mailhog")
		if err != nil {
			panic(err)
		}
		path = dir
	}
	if _, err := os.Stat(path); err != nil {
		err := os.MkdirAll(path, 0770)
		if err != nil {
			panic(err)
		}
	}
	log.Println("Maildir path is", path)
	return &Maildir{
		Path:     path,
		Hostname: hostname,
	}
}

// Store stores a message and returns its storage ID
func (maildir *Maildir) Store(m *data.Message) (string, error) {
	b, err := ioutil.ReadAll(m.Raw.Bytes())
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(filepath.Join(maildir.Path, string(m.ID)), b, 0660)
	return string(m.ID), err
}

// Count returns the number of stored messages
func (maildir *Maildir) Count() int {
	// FIXME may be wrong, ../. ?
	// and handle error?
	dir, err := os.Open(maildir.Path)
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	n, _ := dir.Readdirnames(0)
	return len(n)
}

// Search finds messages matching the query
func (maildir *Maildir) Search(kind, query string, start, limit int) ([]data.Message, int, error) {
	query = strings.ToLower(query)
	var filteredMessages = make([]data.Message, 0)

	var matched int

	err := filepath.Walk(maildir.Path, func(path string, info os.FileInfo, err error) error {
		if limit > 0 && len(filteredMessages) >= limit {
			return errors.New("reached limit")
		}

		if info.IsDir() {
			return nil
		}

		msg, err := maildir.Load(info.Name())
		if err != nil {
			log.Println(err)
			return nil
		}

		switch kind {
		case "to":
			for _, t := range msg.To {
				if strings.Contains(strings.ToLower(t.Mailbox+"@"+t.Domain), query) {
					if start > matched {
						matched++
						break
					}
					filteredMessages = append(filteredMessages, *msg)
					break
				}
			}
		case "from":
			if strings.Contains(strings.ToLower(msg.From.Mailbox+"@"+msg.From.Domain), query) {
				if start > matched {
					matched++
					break
				}
				filteredMessages = append(filteredMessages, *msg)
			}
		case "containing":
			if strings.Contains(strings.ToLower(msg.Raw.Data), query) {
				if start > matched {
					matched++
					break
				}
				filteredMessages = append(filteredMessages, *msg)
			}
		}

		return nil
	})

	if err != nil {
		log.Println(err)
	}

	return filteredMessages, len(filteredMessages), nil
}

// List lists stored messages by index
func (maildir *Maildir) List(start, limit int) ([]data.Message, error) {
	log.Println("Listing messages in", maildir.Path)
	messages := make([]data.Message, 0)

	dir, err := os.Open(maildir.Path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	n, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}
	sort.Slice(n, func(i, j int) bool {
		return !n[i].ModTime().Before(n[j].ModTime())
	})

	if start < 0 {
		start = 0
	}

	if len(n) < start+limit {
		limit = len(n) - start
	}

	if len(n) > start && start > 0 {
		n = n[start:]
	}

	if len(n) > limit {
		n = n[:limit]
	}

	for _, fileinfo := range n {
		b, err := ioutil.ReadFile(filepath.Join(maildir.Path, fileinfo.Name()))
		if err != nil {
			return nil, err
		}
		msg := data.FromBytes(b)
		// FIXME domain
		m := msg.Parse("mailhog.example")
		m.ID = data.MessageID(fileinfo.Name())
		m.Created = fileinfo.ModTime()
		messages = append(messages, *m)
	}

	log.Printf("Found %d messages", len(messages))
	return messages, nil
}

// DeleteOne deletes an individual message by storage ID
func (maildir *Maildir) DeleteOne(id string) error {
	return os.Remove(filepath.Join(maildir.Path, id))
}

// DeleteAll deletes all in memory messages
func (maildir *Maildir) DeleteAll() error {
	err := os.RemoveAll(maildir.Path)
	if err != nil {
		return err
	}
	return os.Mkdir(maildir.Path, 0770)
}

// Load returns an individual message by storage ID
func (maildir *Maildir) Load(id string) (*data.Message, error) {
	b, err := ioutil.ReadFile(filepath.Join(maildir.Path, id))
	if err != nil {
		return nil, err
	}
	// FIXME domain
	m := data.FromBytes(b).Parse("mailhog.example")
	m.ID = data.MessageID(id)
	return m, nil
}
