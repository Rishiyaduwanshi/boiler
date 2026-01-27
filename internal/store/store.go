package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Meta struct {
	Stacks   map[string]string `json:"stacks"`
	Snippets map[string]string `json:"snippets"`
}

type SnippetEntry struct {
	Name        string
	Version     string
	Extension   string
	Path        string
	Author      string
	Description string
}

type StackEntry struct {
	Name        string
	Version     string
	Path        string
	Author      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Store struct {
	metaPath string
	meta     *Meta
}

func NewStore(storePath string) *Store {
	return &Store{
		metaPath: filepath.Join(storePath, "boiler.meta.json"),
		meta: &Meta{
			Stacks:   make(map[string]string),
			Snippets: make(map[string]string),
		},
	}
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.Save()
		}
		return fmt.Errorf("failed to read meta file: %w", err)
	}

	if err := json.Unmarshal(data, s.meta); err != nil {
		return fmt.Errorf("failed to parse meta file: %w", err)
	}

	if s.meta.Stacks == nil {
		s.meta.Stacks = make(map[string]string)
	}
	if s.meta.Snippets == nil {
		s.meta.Snippets = make(map[string]string)
	}

	return nil
}

func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.metaPath), 0755); err != nil {
		return fmt.Errorf("failed to create meta directory: %w", err)
	}

	data, err := json.MarshalIndent(s.meta, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	if err := os.WriteFile(s.metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write meta file: %w", err)
	}

	return nil
}

func (s *Store) AddSnippet(name, path string) error {
	s.meta.Snippets[name] = path
	return s.Save()
}

func (s *Store) AddStack(name, path string) error {
	s.meta.Stacks[name] = path
	return s.Save()
}

func (s *Store) GetSnippet(name string) (string, bool) {
	path, ok := s.meta.Snippets[name]
	return path, ok
}

func (s *Store) GetStack(name string) (string, bool) {
	path, ok := s.meta.Stacks[name]
	return path, ok
}

func (s *Store) RemoveSnippet(name string) error {
	delete(s.meta.Snippets, name)
	return s.Save()
}

func (s *Store) RemoveStack(name string) error {
	delete(s.meta.Stacks, name)
	return s.Save()
}

func (s *Store) ListSnippets() []string {
	snippets := make([]string, 0, len(s.meta.Snippets))
	for name := range s.meta.Snippets {
		snippets = append(snippets, name)
	}
	return snippets
}

func (s *Store) ListStacks() []string {
	stacks := make([]string, 0, len(s.meta.Stacks))
	for name := range s.meta.Stacks {
		stacks = append(stacks, name)
	}
	return stacks
}

func ParseResourceName(resource string) (name, version, extension string) {
	parts := strings.SplitN(resource, "@", 2)
	nameWithExt := parts[0]
	
	if len(parts) == 2 {
		versionWithExt := parts[1]
		// Check if version has extension
		ext := filepath.Ext(versionWithExt)
		if ext != "" {
			version = strings.TrimSuffix(versionWithExt, ext)
			extension = ext
		} else {
			version = versionWithExt
		}
	}

	if extension == "" {
		extension = filepath.Ext(nameWithExt)
		name = strings.TrimSuffix(nameWithExt, extension)
	} else {
		name = nameWithExt
	}

	return name, version, extension
}

func IsStack(resource string) bool {
	_, _, ext := ParseResourceName(resource)
	return ext == ""
}

func IsSnippet(resource string) bool {
	return !IsStack(resource)
}
