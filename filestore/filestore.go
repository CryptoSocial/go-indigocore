// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// Package filestore implements a store that saves all the segments to the file system.
//
// The segments are stored as JSON files named after the link hashes.
// It's a convenient store to use during the development of an agent.
// However, because it doesn't use an index, it's very slow, and shouldn't be used for production.
package filestore

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/store"
)

const (
	// Name is the name set in the store's information.
	Name = "file"

	// Description is the description set in the store's information.
	Description = "Stratumn File Store"

	// DefaultPath is the path where segments will be saved by default.
	DefaultPath = "/var/filestore"
)

// FileStore is the type that implements github.com/stratumn/go/store.Adapter.
type FileStore struct {
	config *Config
}

// Config contains configuration options for the store.
type Config struct {
	// A version string that will set in the store's information.
	Version string

	// A git commit hash that will set in the store's information.
	Commit string

	// Path where segments will be saved.
	Path string
}

// New creates an instance of a FileStore.
func New(config *Config) *FileStore {
	return &FileStore{config}
}

// GetInfo implements github.com/stratumn/go/store.Adapter.GetInfo.
func (a *FileStore) GetInfo() (interface{}, error) {
	return map[string]interface{}{
		"name":        Name,
		"description": Description,
		"version":     a.config.Version,
		"commit":      a.config.Commit,
	}, nil
}

// SaveSegment implements github.com/stratumn/go/store.Adapter.SaveSegment.
func (a *FileStore) SaveSegment(segment *cs.Segment) error {
	js, err := json.MarshalIndent(segment, "", "  ")
	if err != nil {
		return err
	}

	if err = os.MkdirAll(a.config.Path, 0755); err != nil {
		return err
	}

	segmentPath := path.Join(a.config.Path, segment.Meta["linkHash"].(string)+".json")
	return ioutil.WriteFile(segmentPath, js, 0644)
}

// GetSegment implements github.com/stratumn/go/store.Adapter.GetSegment.
func (a *FileStore) GetSegment(linkHash string) (*cs.Segment, error) {
	file, err := os.Open(path.Join(a.config.Path, linkHash+".json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var segment cs.Segment
	if err = json.NewDecoder(file).Decode(&segment); err != nil {
		return nil, err
	}

	return &segment, nil
}

// DeleteSegment implements github.com/stratumn/go/store.Adapter.DeleteSegment.
func (a *FileStore) DeleteSegment(linkHash string) (*cs.Segment, error) {
	segment, err := a.GetSegment(linkHash)
	if segment == nil {
		return segment, err
	}

	if err = os.Remove(path.Join(a.config.Path, linkHash+".json")); err != nil {
		return nil, err
	}

	return segment, err
}

// FindSegments implements github.com/stratumn/go/store.Adapter.FindSegments.
func (a *FileStore) FindSegments(filter *store.Filter) (cs.SegmentSlice, error) {
	var segments cs.SegmentSlice

	a.forEach(func(segment *cs.Segment) error {
		if filter.MapID != "" && filter.MapID != segment.Link.Meta["mapId"].(string) {
			return nil
		}

		prevLinkHash, ok := segment.Link.Meta["prevLinkHash"].(string)
		if filter.PrevLinkHash != "" && (!ok || filter.PrevLinkHash != prevLinkHash) {
			return nil
		}

		if len(filter.Tags) > 0 {
			if t, ok := segment.Link.Meta["tags"].([]interface{}); ok {
				var tags []string
				for _, v := range t {
					tags = append(tags, v.(string))
				}

				for _, tag := range filter.Tags {
					if !containsString(tags, tag) {
						return nil
					}
				}
			} else {
				return nil
			}
		}

		segments = append(segments, segment)

		return nil
	})

	sort.Sort(segments)

	return paginateSegments(segments, &filter.Pagination), nil
}

// GetMapIDs implements github.com/stratumn/go/store.Adapter.GetMapIDs.
func (a *FileStore) GetMapIDs(pagination *store.Pagination) ([]string, error) {
	set := map[string]struct{}{}
	a.forEach(func(segment *cs.Segment) error {
		set[segment.Link.Meta["mapId"].(string)] = struct{}{}
		return nil
	})

	var mapIDs []string
	for mapID := range set {
		mapIDs = append(mapIDs, mapID)
	}

	sort.Strings(mapIDs)
	return paginateStrings(mapIDs, pagination), nil
}

var segmentFileRegepx = regexp.MustCompile("(.*)\\.json$")

func (a *FileStore) forEach(fn func(*cs.Segment) error) error {
	files, err := ioutil.ReadDir(a.config.Path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, file := range files {
		name := file.Name()
		if segmentFileRegepx.MatchString(name) {
			linkHash := name[:len(name)-5]
			segment, err := a.GetSegment(linkHash)
			if err != nil {
				return err
			}
			if segment == nil {
				return errors.New("could not find segment")
			}
			if err = fn(segment); err != nil {
				return err
			}
		}
	}

	return nil
}

func containsString(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}

	return false
}

func paginateStrings(a []string, p *store.Pagination) []string {
	l := len(a)

	if p.Offset >= l {
		return []string{}
	}

	if p.Limit > 0 {
		end := min(l, p.Offset+p.Limit)
		return a[p.Offset:end]
	}

	return a[p.Offset:]
}

func paginateSegments(a cs.SegmentSlice, p *store.Pagination) cs.SegmentSlice {
	l := len(a)

	if p.Offset >= l {
		return cs.SegmentSlice{}
	}

	if p.Limit > 0 {
		end := min(l, p.Offset+p.Limit)
		return a[p.Offset:end]
	}

	return a[p.Offset:]
}

// Min of two ints, duh.
func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
