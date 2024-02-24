package upload

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

// --------------------------------------------------------------------------
// Types
// --------------------------------------------------------------------------

// Upload defines an entity within the persistence store
type Upload struct {
	ID       string    `json:"id"`
	FileName string    `json:"fileName"`
	Payload  []byte    `json:"payload"`
	MimeType string    `json:"mimeType"`
	Created  time.Time `json:"created"`
}

// --------------------------------------------------------------------------
// Store
// --------------------------------------------------------------------------

// Store provides CRUD methods for uploads
type Store interface {
	Write(item Upload) (err error)
	Read(id string) (Upload, error)
	Delete(id string) (err error)
}

// NewStore create a new store instance
func NewStore(path string) Store {
	return &jsonStore{
		path: path,
	}
}

// jsonStore implements the Store interface by serializing the data into a JSON structure
// the jsonStore saves files using the following structur
// /basepath
//     /json
//	    id.json
//     /files
//         /id
//             <file>
// the folder /json is used like a database where the metadata of each upload is stored
// the "payload" is stored in the folder /files with an folder per id and the <file> payload saved within the folder
type jsonStore struct {
	// path is the filesystem base-path
	path string
}

// compile time assertion of Upload interface
var (
	_ Store = &jsonStore{}
)

const jsonPath = "json"
const filesPath = "files"

// Write saves a new payload to the jsonStore structure
func (s *jsonStore) Write(item Upload) (err error) {
	if item.ID == "" {
		return fmt.Errorf("the supplied ID is empty")
	}

	if len(item.Payload) == 0 {
		return fmt.Errorf("a empty payload was provided")
	}

	payload := make([]byte, len(item.Payload))
	copy(payload, item.Payload)
	item.Payload = nil

	// the "db-path"
	metaPath := getMetaPath(s.path)
	if err := ensurePath(metaPath); err != nil {
		return fmt.Errorf("could not ensure path for meta-data: %v", err)
	}
	metaPath = getMetaPathFile(s.path, item.ID)

	mety, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("could not marshall JSON: %v", err)
	}
	if err = ioutil.WriteFile(metaPath, mety, 0660); err != nil {
		return fmt.Errorf("could not store item metadata: %v", err)
	}

	// after writing the meta-data; save the actual payload
	filePath := getFilePath(s.path, item.ID)
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		err = os.MkdirAll(filePath, 0750)
	}
	if err != nil {
		return fmt.Errorf("could not create payload directory: %v", err)
	}

	filePathName := path.Join(s.path, filesPath, item.ID, item.FileName)
	if err = ioutil.WriteFile(filePathName, payload, 0660); err != nil {
		return fmt.Errorf("could not save payload: %v", err)
	}

	return nil
}

// Read returns a previously saved Upload item
func (s *jsonStore) Read(id string) (Upload, error) {
	item := Upload{}

	if id == "" {
		return item, fmt.Errorf("the supplied ID is empty")
	}
	// read the "db"
	metaPath := getMetaPath(s.path)
	if err := ensurePath(metaPath); err != nil {
		return item, fmt.Errorf("could not ensure path for meta-data: %v", err)
	}
	metaPath = getMetaPathFile(s.path, id)
	metaPayload, err := ioutil.ReadFile(metaPath)
	if err != nil {
		return item, fmt.Errorf("could not read meta-data file '%s': %v", id, err)
	}
	if err = json.Unmarshal(metaPayload, &item); err != nil {
		return item, fmt.Errorf("could not unmarshal JSON: %v", err)
	}

	// read the payload file
	payloadPath := getFilePath(s.path, id)
	payloadFile := path.Join(payloadPath, item.FileName)
	filePayload, err := ioutil.ReadFile(payloadFile)
	if err != nil {
		return item, fmt.Errorf("could not read payload file: %v", err)
	}
	item.Payload = filePayload
	return item, nil
}

// Delete removes the entries for the given id
func (s *jsonStore) Delete(id string) (err error) {
	if id == "" {
		return fmt.Errorf("the supplied ID is empty")
	}
	// read the "db"
	metaPath := getMetaPath(s.path)
	if err := ensurePath(metaPath); err != nil {
		return fmt.Errorf("could not ensure path for meta-data: %v", err)
	}
	metaPath = getMetaPathFile(s.path, id)
	if err = os.RemoveAll(metaPath); err != nil {
		return fmt.Errorf("could not remove the meta-data for '%s': %v", id, err)
	}

	// read the payload file
	payloadPath := getFilePath(s.path, id)
	if err = os.RemoveAll(payloadPath); err != nil {
		return fmt.Errorf("could not remove the payload for '%s': %v", id, err)
	}

	return nil
}

func getMetaPath(basePath string) string {
	return path.Join(basePath, jsonPath)
}

func getMetaPathFile(basePath, id string) string {
	return path.Join(basePath, jsonPath, id+".json")
}

func getFilePath(basePath, id string) string {
	return path.Join(basePath, filesPath, id)
}

func ensurePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0750)
	}
	return nil
}
