package serializer

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	"github.com/mwinyimoha/commons/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Src struct {
	ID        primitive.ObjectID
	CreatedAt time.Time
}

type Dst struct {
	ID        string
	CreatedAt *timestamppb.Timestamp
}

func TestSerializer(t *testing.T) {

	t.Run("Register Hook", func(t *testing.T) {
		serializer := New()

		serializer.RegisterHook(func(from reflect.Type, to reflect.Type, data any) (any, error) {
			return data, nil
		})

		assert.Len(t, serializer.hooks, 4)
	})
}

func TestHooks(t *testing.T) {

	t.Run("ObjectID Hook", func(t *testing.T) {
		id := primitive.NewObjectID()

		// ObjectID -> string
		out, err := objectIDHook(reflect.TypeOf(id), reflect.TypeOf(""), id)
		require.NoError(t, err)
		assert.Equal(t, id.Hex(), out)

		// string -> ObjectID (valid)
		out, err = objectIDHook(reflect.TypeOf(""), reflect.TypeOf(primitive.ObjectID{}), id.Hex())
		require.NoError(t, err)
		assert.Equal(t, id.Hex(), out.(primitive.ObjectID).Hex())

		// string -> ObjectID (invalid)
		_, err = objectIDHook(reflect.TypeOf(""), reflect.TypeOf(primitive.ObjectID{}), "invalid_hex")
		assert.Error(t, err)
	})

	t.Run("UUID Hook", func(t *testing.T) {
		u := uuid.New()

		// uuid -> string
		out, err := uuidHook(reflect.TypeOf(u), reflect.TypeOf(""), u)
		require.NoError(t, err)
		assert.Equal(t, u.String(), out)

		// string -> uuid (valid)
		out, err = uuidHook(reflect.TypeOf(""), reflect.TypeOf(uuid.UUID{}), u.String())
		require.NoError(t, err)
		assert.Equal(t, u.String(), out.(uuid.UUID).String())

		// string -> uuid (invalid)
		_, err = uuidHook(reflect.TypeOf(""), reflect.TypeOf(uuid.UUID{}), "not-a-uuid")
		assert.Error(t, err)
	})

	t.Run("Timestamp Hook", func(t *testing.T) {
		now := time.Now()

		// time.Time -> *timestamppb.Timestamp
		out, err := timestampHook(reflect.TypeOf(time.Time{}), reflect.TypeOf(&timestamppb.Timestamp{}), now)
		require.NoError(t, err)
		ts, ok := out.(*timestamppb.Timestamp)
		require.True(t, ok)
		assert.Equal(t, now.Unix(), ts.AsTime().Unix())
	})
}

func TestSerialize(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		s := New()

		now := time.Now().UTC()
		src := &Src{
			ID:        primitive.NewObjectID(),
			CreatedAt: now,
		}

		dst, err := Serialize[Src, Dst](s, src)
		assert.NoError(t, err)
		assert.Equal(t, src.ID.Hex(), dst.ID)
		assert.NotNil(t, dst.CreatedAt)
		assert.Equal(t, now.Unix(), dst.CreatedAt.AsTime().Unix(), "time.Time -> timestamppb.Timestamp")
	})

	t.Run("With Serialization Error", func(t *testing.T) {
		s := New()

		type BadSrc struct {
			ID string
		}

		src := &BadSrc{ID: "invalid"}

		_, err := Serialize[BadSrc, Src](s, src)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ObjectID")
	})

	t.Run("No Hook Changes", func(t *testing.T) {
		s := New()

		type PlainSrc struct {
			Name string
		}
		type PlainDst struct {
			Name string
		}

		src := &PlainSrc{Name: "hello"}
		dst, err := Serialize[PlainSrc, PlainDst](s, src)

		assert.NoError(t, err)
		assert.Equal(t, "hello", dst.Name) // Must stay unchanged
	})
}

func TestSerializeSlice(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		s := New()

		src := []*Src{
			{ID: primitive.NewObjectID(), CreatedAt: time.Now()},
		}

		dst, err := SerializeSlice[Src, Dst](s, src)

		assert.NoError(t, err)
		assert.Len(t, dst, 1)
		assert.Equal(t, src[0].ID.Hex(), dst[0].ID)
	})

	t.Run("Serialization Error", func(t *testing.T) {
		s := New()

		type BadSrc struct {
			ID string
		}

		src := []*BadSrc{
			{ID: "invalid"},
		}

		_, err := SerializeSlice[BadSrc, Src](s, src)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ObjectID")
	})
}

func TestDecoderCache(t *testing.T) {
	ser := New()

	_, _ = Serialize[Src, Dst](ser, &Src{})
	_, _ = Serialize[Src, Dst](ser, &Src{})

	count := 0
	ser.cache.Range(func(_, _ any) bool {
		count++
		return true
	})

	assert.Equal(t, 1, count)
}

func TestGetDecoderFailure(t *testing.T) {
	s := New()

	s.decoderFunc = func(cfg *mapstructure.DecoderConfig) (*mapstructure.Decoder, error) {
		return nil, errors.NewErrorf(errors.Internal, "forced failure")
	}

	src := &Src{}
	_, err := Serialize[Src, Dst](s, src)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forced failure")
}
