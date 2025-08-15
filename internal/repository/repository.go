package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"post-analyzer/internal/entity"
	"sync"
)

type ChannelStorageRepository interface {
	AddChannelInfo(context.Context, *entity.ChannelInfo) error
	GetAllChannelInfos(context.Context) ([]*entity.ChannelInfo, error)
	GetChannelInfos(context.Context, *entity.ChannelInfo) ([]*entity.ChannelInfo, error)
	UpdateChannelInfo(context.Context, *entity.ChannelInfo) error
	DeleteChannelInfo(context.Context, *entity.ChannelInfo) error
}

type jsonStorage struct {
	filepath     string
	channelInfos []*entity.ChannelInfo
	mutex        sync.Mutex
}

// NewRepo + CRUD methods
func NewChannelStorage(path string) (ChannelStorageRepository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to get absolute path of storage: %w", err)
	}

	storage := &jsonStorage{
		filepath:     absPath,
		channelInfos: make([]*entity.ChannelInfo, 0),
	}

	if err := storage.load(); err != nil {
		return nil, fmt.Errorf("Failed to load data: %w", err)
	}

	return storage, nil
}

func (s *jsonStorage) AddChannelInfo(ctx context.Context, c *entity.ChannelInfo) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, channel := range s.channelInfos {
		if channel.ChannelID == c.ChannelID && channel.ChatID == c.ChatID && channel.ScheduleID == c.ScheduleID {
			return fmt.Errorf("This channel is already monitored at this time: %d", c.ChannelID)
		}
	}

	s.channelInfos = append(s.channelInfos, c)

	if err := s.save(); err != nil {
		return fmt.Errorf("Failed to save storage changes: %w", err)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

func (s *jsonStorage) GetAllChannelInfos(ctx context.Context) ([]*entity.ChannelInfo, error) {
	infos := make([]*entity.ChannelInfo, 0)
	infos = append(infos, s.channelInfos...)

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return infos, nil
}

func (s *jsonStorage) GetChannelInfos(ctx context.Context, chanInfo *entity.ChannelInfo) ([]*entity.ChannelInfo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	infos := make([]*entity.ChannelInfo, 0)
	for _, channel := range s.channelInfos {
		if channel.ChannelID == chanInfo.ChannelID && channel.ChatID == chanInfo.ChatID {
			infos = append(infos, channel)
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return infos, nil
}

func (s *jsonStorage) UpdateChannelInfo(ctx context.Context, c *entity.ChannelInfo) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range s.channelInfos {
		if s.channelInfos[i].ChannelID == c.ChannelID && s.channelInfos[i].ChatID == c.ChatID {
			s.channelInfos[i].LastCheckedPostID = c.LastCheckedPostID

			err := s.save()
			if err != nil {
				return err
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}

			return nil
		}
	}

	return fmt.Errorf("No channel information found to update")
}

func (s *jsonStorage) DeleteChannelInfo(ctx context.Context, c *entity.ChannelInfo) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range s.channelInfos {
		if s.channelInfos[i].ChannelID == c.ChannelID && s.channelInfos[i].ChatID == c.ChatID && s.channelInfos[i].ScheduleID == c.ScheduleID {
			s.channelInfos = append(s.channelInfos[:i], s.channelInfos[i+1:]...)

			err := s.save()
			if err != nil {
				return err
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}

			return nil
		}
	}

	return fmt.Errorf("No channel information found to delete")
}

// some additional methods to work with (de)serialization
func (s *jsonStorage) save() error {

	if err := os.MkdirAll(filepath.Dir(s.filepath), 0755); err != nil {
		return fmt.Errorf("directory creation failed: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(s.filepath), "tmp-*.json")
	if err != nil {
		return fmt.Errorf("Failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(s.channelInfos); err != nil {
		tmpFile.Close()
		return fmt.Errorf("JSON encoding failed: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("Tmp file sync failed: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("Tmp file closing failed: %w", err)
	}

	if err := os.Rename(tmpPath, s.filepath); err != nil {
		return fmt.Errorf("Storage file replacement failed: %w", err)
	}

	return nil
}

func (s *jsonStorage) load() error {
	if _, err := os.Stat(s.filepath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(s.filepath)
	if err != nil {
		return fmt.Errorf("Failed to open storage file: %w", err)
	}
	defer file.Close()

	var channels []*entity.ChannelInfo
	if err := json.NewDecoder(file).Decode(&channels); err != nil {
		return fmt.Errorf("JSON decoding storage file failed: %w", err)
	}

	s.channelInfos = channels
	return nil
}
