package secrets

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/util"

	"github.com/grafana/grafana/pkg/bus"

	"github.com/grafana/grafana/pkg/models"

	"github.com/grafana/grafana/pkg/services/sqlstore"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"
)

var logger = log.New("secrets")

type Secrets struct {
	store *sqlstore.SQLStore `inject:""`
	bus   bus.Bus            `inject:""`

	defaultEncryptionKey string
	defaultProvider      string
	providers            map[string]Provider
	dataKeyCache         map[string]dataKeyCacheItem
}

type dataKeyCacheItem struct {
	expiry  time.Time
	dataKey []byte
}

type Provider interface {
	Encrypt(blob []byte) ([]byte, error)
	Decrypt(blob []byte) ([]byte, error)
}

func (s *Secrets) Init() error {
	s.providers = map[string]Provider{
		"": &secretKey{
			key: func() []byte {
				return []byte(setting.SecretKey)
			},
		},
	}

	base_key := "root"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.store.GetDataKey(ctx, base_key)
	if err != nil {
		if errors.Is(err, models.ErrDataKeyNotFound) {
			err = s.newRandomDataKey(ctx, base_key)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	util.Encrypt = s.Encrypt
	util.Decrypt = s.Decrypt

	return nil
}

func (s *Secrets) newRandomDataKey(ctx context.Context, base_key string) error {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}

	encrypted, err := s.Encrypt(b)
	if err != nil {
		return err
	}

	err = s.store.CreateDataKey(ctx, models.DataKey{
		Active:        true,
		Name:          base_key,
		Provider:      s.defaultProvider,
		EncryptedData: encrypted,
	})
	return nil
}

var b64 = base64.RawStdEncoding

func (s *Secrets) Encrypt(payload []byte) ([]byte, error) {
	key := s.defaultEncryptionKey

	dataKey, err := s.dataKey(key)
	if err != nil {
		return nil, err
	}

	encrypted, err := encrypt(payload, dataKey)
	if err != nil {
		return nil, err
	}

	prefix := make([]byte, b64.EncodedLen(len(key))+2)
	b64.Encode(prefix[1:], []byte(key))
	prefix[0] = '#'
	prefix[len(prefix)-1] = '#'

	blob := make([]byte, len(prefix)+len(encrypted))
	copy(blob, prefix)
	copy(blob[len(prefix):], encrypted)

	return blob, nil
}

func (s *Secrets) Decrypt(payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		return []byte{}, nil // TODO: Not sure if it should return error like decrypt does (also see tests)
	}

	var dataKey []byte

	if payload[0] != '#' {
		dataKey = []byte(setting.SecretKey)
	} else {
		payload = payload[1:]
		endOfKey := bytes.Index(payload, []byte{'#'})
		if endOfKey == -1 {
			return nil, fmt.Errorf("could not find valid key in encrypted payload")
		}
		b64Key := payload[:endOfKey]
		payload = payload[endOfKey+1:]
		key := make([]byte, b64.DecodedLen(len(b64Key)))
		_, err := b64.Decode(key, b64Key)
		if err != nil {
			return nil, err
		}

		dataKey, err = s.dataKey(string(key))
		if err != nil {
			return nil, err
		}
	}

	return decrypt(payload, dataKey)
}

func (s *Secrets) dataKey(key string) ([]byte, error) {
	if key == "" {
		return []byte(setting.SecretKey), nil
	}

	if item, exists := s.dataKeyCache[key]; exists {
		if item.expiry.Before(time.Now()) && !item.expiry.IsZero() {
			delete(s.dataKeyCache, key)
		} else {
			return item.dataKey, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 1. get encrypted data key from database
	dataKey, err := s.store.GetDataKey(ctx, key)
	if err != nil {
		return nil, err
	}

	// 2. decrypt data key
	provider, exists := s.providers[dataKey.Provider]
	if !exists {
		return nil, fmt.Errorf("could not find encryption provider '%s'", dataKey.Provider)
	}

	decrypted, err := provider.Decrypt(dataKey.EncryptedData)
	if err != nil {
		return nil, err
	}

	// 3. cache data key
	s.dataKeyCache[key] = dataKeyCacheItem{
		expiry:  time.Now().Add(15 * time.Minute),
		dataKey: decrypted,
	}

	return decrypted, nil
}
