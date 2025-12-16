package serializer

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/mwinyimoha/commons/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ConverterFunc func(src any) (any, error)

type converter struct {
	srcType any
	dstType any
	fn      ConverterFunc
}

type Serializer struct {
	converters []converter
	mu         sync.RWMutex
	opt        copier.Option
	dirty      bool
}

// New creates a serializer instance with commonly used converters
func New() *Serializer {
	s := &Serializer{
		converters: make([]converter, 0),
		dirty:      true,
	}

	s.registerDefaults()
	return s
}

// RegisterConverter adds a custom converter to handle project-specific conversions
// src and dst should be zero values of the types you want to convert
// Example: s.RegisterConverter(MyEnum(0), "", func(src any) (any, error) { ... })
func (s *Serializer) RegisterConverter(src, dst any, fn ConverterFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.addConverter(src, dst, fn)
}

func (s *Serializer) addConverter(srcType, dstType any, fn ConverterFunc) {
	s.converters = append(s.converters, converter{
		srcType: srcType,
		dstType: dstType,
		fn:      fn,
	})

	s.dirty = true
}

func (s *Serializer) registerDefaults() {
	// ObjectID -> string
	s.addConverter(
		primitive.ObjectID{},
		"",
		func(src any) (any, error) {
			if val, ok := src.(primitive.ObjectID); ok {
				return val.Hex(), nil
			}
			return nil, errors.NewErrorf(errors.Internal, "failed to convert ObjectID to string")
		},
	)

	// string -> ObjectID
	s.addConverter(
		"",
		primitive.ObjectID{},
		func(src any) (any, error) {
			if str, ok := src.(string); ok {
				objectID, err := primitive.ObjectIDFromHex(str)
				if err != nil {
					return nil, errors.WrapError(err, errors.InvalidArgument, "invalid ObjectID: %s", str)
				}
				return objectID, nil
			}
			return nil, errors.NewErrorf(errors.Internal, "failed to convert string to ObjectID")
		},
	)

	// UUID -> string
	s.addConverter(
		uuid.UUID{},
		"",
		func(src any) (any, error) {
			if val, ok := src.(uuid.UUID); ok {
				return val.String(), nil
			}
			return nil, errors.NewErrorf(errors.Internal, "failed to convert UUID to string")
		},
	)

	// string -> UUID
	s.addConverter(
		"",
		uuid.UUID{},
		func(src any) (any, error) {
			if str, ok := src.(string); ok {
				uid, err := uuid.Parse(str)
				if err != nil {
					return nil, errors.WrapError(err, errors.InvalidArgument, "invalid UUID string: %s", str)
				}
				return uid, nil
			}
			return nil, errors.NewErrorf(errors.Internal, "failed to convert string to UUID")
		},
	)

	// time.Time -> *timestamppb.Timestamp
	s.addConverter(
		time.Time{},
		&timestamppb.Timestamp{},
		func(src any) (any, error) {
			if t, ok := src.(time.Time); ok {
				return timestamppb.New(t), nil
			}
			return nil, errors.NewErrorf(errors.Internal, "failed to convert time.Time to Timestamp")
		},
	)

	// *time.Time -> *timestamppb.Timestamp
	s.addConverter(
		&time.Time{},
		&timestamppb.Timestamp{},
		func(src any) (any, error) {
			if tp, ok := src.(*time.Time); ok {
				if tp == nil {
					return (*timestamppb.Timestamp)(nil), nil
				}
				return timestamppb.New(*tp), nil
			}
			return nil, errors.NewErrorf(errors.Internal, "failed to convert *time.Time to Timestamp")
		},
	)
}

func (s *Serializer) buildOption() copier.Option {
	converters := make([]copier.TypeConverter, 0, len(s.converters))

	for _, conv := range s.converters {
		converters = append(converters, copier.TypeConverter{
			SrcType: conv.srcType,
			DstType: conv.dstType,
			Fn:      conv.fn,
		})
	}

	return copier.Option{
		IgnoreEmpty: false,
		DeepCopy:    true,
		Converters:  converters,
	}
}

func (s *Serializer) getOption() copier.Option {
	s.mu.RLock()
	if !s.dirty {
		opt := s.opt
		s.mu.RUnlock()
		return opt
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.dirty {
		return s.opt
	}

	s.opt = s.buildOption()
	s.dirty = false
	return s.opt
}

// Serialize copies fields from source to destination with type conversions
func Serialize[S any, D any](s *Serializer, src *S) (*D, error) {
	if src == nil {
		return nil, nil
	}

	var dst D
	opt := s.getOption()
	if err := copier.CopyWithOption(&dst, src, opt); err != nil {
		return nil, errors.WrapError(err, errors.Internal, "unable to serialize object")
	}

	return &dst, nil
}

// SerializeSlice efficiently serializes a slice of objects
func SerializeSlice[S any, D any](s *Serializer, src []*S) ([]*D, error) {
	if len(src) == 0 {
		return []*D{}, nil
	}

	dst := make([]*D, len(src))
	opt := s.getOption()
	for i, item := range src {
		if item == nil {
			dst[i] = nil
			continue
		}

		var result D
		if err := copier.CopyWithOption(&result, item, opt); err != nil {
			return nil, errors.WrapError(err, errors.Internal, "failed to serialize slice item %d", i)
		}

		dst[i] = &result
	}

	return dst, nil
}
