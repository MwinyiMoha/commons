package serializer

import (
	"reflect"
	"sync"

	"github.com/go-viper/mapstructure/v2"
	"github.com/mwinyimoha/commons/pkg/errors"
)

type Serializer struct {
	cache       sync.Map
	decoderFunc func(cfg *mapstructure.DecoderConfig) (*mapstructure.Decoder, error)
	hooks       []mapstructure.DecodeHookFunc
}

// New creates an serializer instance with commonly used hooks
func New() *Serializer {
	return &Serializer{
		cache:       sync.Map{},
		decoderFunc: mapstructure.NewDecoder,
		hooks: []mapstructure.DecodeHookFunc{
			objectIDHook,
			uuidHook,
			timestampHook,
		},
	}
}

// RegisterHook appends a custom hook to a serializer instance to extend decode functionality
func (s *Serializer) RegisterHook(h mapstructure.DecodeHookFunc) {
	s.hooks = append(s.hooks, h)
}

func (s *Serializer) getDecoder(dst any) (*mapstructure.Decoder, error) {
	key := reflect.TypeOf(dst).String()

	if cached, ok := s.cache.Load(key); ok {
		return cached.(*mapstructure.Decoder), nil
	}

	decoder, err := s.decoderFunc(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(s.hooks...),
		Result:     dst,
	})
	if err != nil {
		return nil, err
	}

	s.cache.Store(key, decoder)
	return decoder, nil
}

// Serialize uses a decoder to map fields from a source to a destination type.
// All required hooks MUST be registered before calling Serialize
func Serialize[S any, D any](s *Serializer, src *S) (*D, error) {
	var dst D

	decoder, err := s.getDecoder(&dst)
	if err != nil {
		return nil, errors.WrapError(err, errors.Internal, "unable to create data decoder")
	}

	if err := decoder.Decode(src); err != nil {
		return nil, errors.WrapError(err, errors.Internal, "unable to map struct")
	}

	return &dst, nil
}

// SerializeSlice is an extension of Serialize for slices
func SerializeSlice[S any, D any](s *Serializer, src []*S) ([]*D, error) {
	dst := make([]*D, 0)

	for _, item := range src {
		result, err := Serialize[S, D](s, item)
		if err != nil {
			return nil, err
		}

		dst = append(dst, result)
	}

	return dst, nil
}
