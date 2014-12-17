package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/nightlyone/lockfile"
)

type TemplateManager interface {
	RegisterTemplate(filepath, installpath string) error
	RemoveTemplate(installpath string) error
	WhiteList(installpath string, vers []WebotsVersion) error
	BlackList(installpath string, vers []WebotsVersion) error

	ApplyTemplates(basepath string, v WebotsVersion) error
}

type Template struct {
	Installpath, Datapath string
	Whitelist, Blacklist  map[string]bool
}

type HashTemplateManager struct {
	byPath   map[string]Template
	basepath string
	lock     lockfile.Lockfile
}

func NewHasHTemplateManager(basepath string) (*HashTemplateManager, error) {
	res := &HashTemplateManager{
		byPath:   make(map[string]Template),
		basepath: basepath,
	}
	var err error
	err = os.MkdirAll(basepath, 0755)
	if err != nil {
		return nil, err
	}
	res.lock, err = lockfile.New(path.Join(basepath, "global.lock"))
	if err != nil {
		return nil, err
	}

	if err := res.tryLock(); err != nil {
		return nil, err
	}
	defer res.unlock()

	err = res.load()
	if err != nil && err != io.EOF {
		return nil, err
	}
	return res, nil

}

func (m *HashTemplateManager) tryLock() error {
	if err := m.lock.TryLock(); err != nil {
		return fmt.Errorf("COuld not lock %s: %s", m.lock, err)
	}
	return nil
}

func (m *HashTemplateManager) unlock() {
	if err := m.lock.Unlock(); err != nil {
		panic(err)
	}
}

func (m *HashTemplateManager) load() error {
	f, err := os.Open(path.Join(m.basepath, "data.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&m.byPath)
	if err != nil {
		return err
	}

	return nil
}

func (m *HashTemplateManager) save() error {
	f, err := os.Create(path.Join(m.basepath, "data.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)

	err = enc.Encode(&m.byPath)
	if err != nil {
		return err
	}

	return nil
}

func (m *HashTemplateManager) RegisterTemplate(filepath, installpath string) error {
	if err := m.tryLock(); err != nil {
		return err
	}
	defer m.unlock()

	if _, ok := m.byPath[installpath]; ok == true {
		return fmt.Errorf("Template to %s is already installed", installpath)
	}

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	var content bytes.Buffer
	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(&content, hash), f)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%x.data", hash.Sum([]byte(installpath)))

	dest, err := os.Create(path.Join(m.basepath, filename))
	_, err = io.Copy(dest, &content)
	if err != nil {
		return err
	}

	m.byPath[installpath] = Template{
		Datapath:    filename,
		Installpath: installpath,
		Blacklist:   make(map[string]bool),
		Whitelist:   make(map[string]bool),
	}

	return m.save()
}

func (m *HashTemplateManager) WhiteList(installpath string, vers []WebotsVersion) error {
	if err := m.tryLock(); err != nil {
		return err
	}
	defer m.unlock()

	_, ok := m.byPath[installpath]
	if ok == false {
		return fmt.Errorf("Unknown template %s", installpath)
	}
	for _, v := range vers {
		m.byPath[installpath].Whitelist[v.String()] = true
	}

	return m.save()
}
func (m *HashTemplateManager) BlackList(installpath string, vers []WebotsVersion) error {
	if err := m.tryLock(); err != nil {
		return err
	}
	defer m.unlock()

	_, ok := m.byPath[installpath]
	if ok == false {
		return fmt.Errorf("Unknown template %s", installpath)
	}
	for _, v := range vers {
		m.byPath[installpath].Blacklist[v.String()] = true
	}

	return m.save()
}

func (m *HashTemplateManager) uninstallTemplate(basepath string, t Template) error {
	absTarget := path.Join(basepath, t.Installpath)
	_, err := os.Lstat(absTarget)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	//check that it is a target
	dataTarget, err := os.Readlink(absTarget)
	if err != nil {
		return err
	}
	if dataTarget != path.Join(m.basepath, t.Datapath) {
		return fmt.Errorf("Synlink %s installed is not a template (points to %s )", absTarget, dataTarget)
	}

	return os.Remove(absTarget)
}

func (m *HashTemplateManager) installTemplate(basepath string, t Template) error {
	absTarget := path.Join(basepath, t.Installpath)
	_, err := os.Lstat(absTarget)
	if err != nil && os.IsNotExist(err) == false {
		return err
	}
	if err == nil {
		return fmt.Errorf("File %s already exists", absTarget)
	}

	return os.Symlink(path.Join(m.basepath, t.Datapath), absTarget)

}

func (m *HashTemplateManager) ApplyTemplates(basepath string, v WebotsVersion) error {
	if err := m.tryLock(); err != nil {
		return err
	}
	defer m.unlock()

	//check if blacklisted
	for _, t := range m.byPath {
		//if blacklisted
		shouldRemove := false
		if _, ok := t.Blacklist[v.String()]; ok == true {
			shouldRemove = true
		}
		if len(t.Whitelist) != 0 {
			if _, ok := t.Whitelist[v.String()]; ok == false {
				shouldRemove = true
			}
		}

		if shouldRemove {
			if err := m.uninstallTemplate(basepath, t); err != nil {
				return err
			}
		} else {
			if err := m.installTemplate(basepath, t); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *HashTemplateManager) RemoveTemplate(installpath string) error {
	if err := m.tryLock(); err != nil {
		return err
	}
	defer m.unlock()

	t, ok := m.byPath[installpath]
	if ok == false {
		return fmt.Errorf("Id %s not found", installpath)
	}

	err := os.Remove(path.Join(m.basepath, t.Datapath))
	if err != nil {
		return err
	}
	//we still conserve that we should remove these links
	delete(m.byPath, installpath)
	return m.save()
}
