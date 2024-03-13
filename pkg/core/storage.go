package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

type Manager interface {
	Store(file *File) error
}

type Storage struct {
	directory string
}

func NewStorage(dir string) Storage {
	return Storage{
		directory: dir,
	}
}

func (s *Storage) Store(file *File, filename string) error {
	if _, err := os.Stat(s.directory); os.IsNotExist(err) {
		_ = os.Mkdir(s.directory, os.ModePerm)
	}
	err := ioutil.WriteFile(fmt.Sprintf("%s%s", s.directory, filename), file.buffer.Bytes(), os.ModePerm)
	return err
}

type File struct {
	name   string
	buffer *bytes.Buffer
}

func NewFile(name string) *File {
	return &File{
		name:   name,
		buffer: &bytes.Buffer{},
	}
}

func (f *File) Write(chunk []byte) error {
	_, err := f.buffer.Write(chunk)
	return err
}
