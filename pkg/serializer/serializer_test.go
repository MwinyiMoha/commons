// serializer_test.go
package serializer

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Test structs for ObjectID conversion
type ObjectIDSource struct {
	ID primitive.ObjectID
}

type ObjectIDDest struct {
	ID string
}

// Test structs for UUID conversion
type UUIDSource struct {
	UserID uuid.UUID
}

type UUIDDest struct {
	UserID string
}

// Test structs for Timestamp conversion
type TimeSource struct {
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	OptionalAt *time.Time
}

type TimeDest struct {
	CreatedAt  *timestamppb.Timestamp
	UpdatedAt  *timestamppb.Timestamp
	OptionalAt *timestamppb.Timestamp
}

// Test structs for combined conversions
type ComplexSource struct {
	ID        primitive.ObjectID
	UserID    uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time
	Count     int
}

type ComplexDest struct {
	ID        string
	UserID    string
	Name      string
	CreatedAt *timestamppb.Timestamp
	UpdatedAt *timestamppb.Timestamp
	Count     int
}

func TestObjectIDConversion(t *testing.T) {

	t.Run("ObjectID To String", func(t *testing.T) {
		s := New()
		objectID := primitive.NewObjectID()

		src := &ObjectIDSource{
			ID: objectID,
		}

		dst, err := Serialize[ObjectIDSource, ObjectIDDest](s, src)

		require.NoError(t, err)
		require.NotNil(t, dst)
		assert.Equal(t, objectID.Hex(), dst.ID)
	})

	t.Run("String To ObjectID", func(t *testing.T) {
		s := New()
		objectID := primitive.NewObjectID()

		src := &ObjectIDDest{
			ID: objectID.Hex(),
		}

		dst, err := Serialize[ObjectIDDest, ObjectIDSource](s, src)

		require.NoError(t, err)
		require.NotNil(t, dst)
		assert.Equal(t, objectID, dst.ID)
	})

	t.Run("Invalid ObjectID Hex", func(t *testing.T) {
		s := New()

		src := &ObjectIDDest{
			ID: "invalid-hex-string",
		}

		dst, err := Serialize[ObjectIDDest, ObjectIDSource](s, src)

		assert.Error(t, err)
		assert.Nil(t, dst)
		assert.Contains(t, err.Error(), "invalid ObjectID")
	})
}

func TestUUIDConversion(t *testing.T) {

	t.Run("UUID To String", func(t *testing.T) {
		s := New()
		uid := uuid.New()

		src := &UUIDSource{
			UserID: uid,
		}

		dst, err := Serialize[UUIDSource, UUIDDest](s, src)

		require.NoError(t, err)
		require.NotNil(t, dst)
		assert.Equal(t, uid.String(), dst.UserID)
	})

	t.Run("String To UUID", func(t *testing.T) {
		s := New()
		uid := uuid.New()

		src := &UUIDDest{
			UserID: uid.String(),
		}

		dst, err := Serialize[UUIDDest, UUIDSource](s, src)

		require.NoError(t, err)
		require.NotNil(t, dst)
		assert.Equal(t, uid, dst.UserID)
	})

	t.Run("Invalid UUID String", func(t *testing.T) {
		s := New()

		src := &UUIDDest{
			UserID: "not-a-valid-uuid",
		}

		dst, err := Serialize[UUIDDest, UUIDSource](s, src)

		assert.Error(t, err)
		assert.Nil(t, dst)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}

func TestTimestampConversion(t *testing.T) {

	t.Run("Time To Timestamp", func(t *testing.T) {
		s := New()
		now := time.Now()
		updatedAt := time.Now().Add(1 * time.Hour)

		src := &TimeSource{
			CreatedAt:  now,
			UpdatedAt:  &updatedAt,
			OptionalAt: nil,
		}

		dst, err := Serialize[TimeSource, TimeDest](s, src)

		require.NoError(t, err)
		require.NotNil(t, dst)
		require.NotNil(t, dst.CreatedAt)
		require.NotNil(t, dst.UpdatedAt)
		assert.Nil(t, dst.OptionalAt)

		// Compare timestamps (allowing for minor precision differences)
		assert.Equal(t, now.Unix(), dst.CreatedAt.Seconds)
		assert.InDelta(t, now.Nanosecond(), dst.CreatedAt.Nanos, 1000)

		assert.Equal(t, updatedAt.Unix(), dst.UpdatedAt.Seconds)
		assert.InDelta(t, updatedAt.Nanosecond(), dst.UpdatedAt.Nanos, 1000)
	})

}

func TestSerialize(t *testing.T) {

	t.Run("Nil Source", func(t *testing.T) {
		s := New()

		dst, err := Serialize[ComplexSource, ComplexDest](s, nil)

		require.NoError(t, err)
		assert.Nil(t, dst)
	})

	t.Run("Complex Object", func(t *testing.T) {
		s := New()

		objectID := primitive.NewObjectID()
		uid := uuid.New()
		now := time.Now()
		updatedAt := time.Now().Add(2 * time.Hour)

		src := &ComplexSource{
			ID:        objectID,
			UserID:    uid,
			Name:      "John Doe",
			CreatedAt: now,
			UpdatedAt: &updatedAt,
			Count:     42,
		}

		dst, err := Serialize[ComplexSource, ComplexDest](s, src)

		require.NoError(t, err)
		require.NotNil(t, dst)

		// Verify all conversions
		assert.Equal(t, objectID.Hex(), dst.ID)
		assert.Equal(t, uid.String(), dst.UserID)
		assert.Equal(t, "John Doe", dst.Name)
		assert.Equal(t, 42, dst.Count)

		require.NotNil(t, dst.CreatedAt)
		assert.Equal(t, now.Unix(), dst.CreatedAt.Seconds)

		require.NotNil(t, dst.UpdatedAt)
		assert.Equal(t, updatedAt.Unix(), dst.UpdatedAt.Seconds)
	})

}

func TestSerializeSlice(t *testing.T) {

	t.Run("Empty Slice", func(t *testing.T) {
		s := New()

		src := []*ObjectIDSource{}

		dst, err := SerializeSlice[ObjectIDSource, ObjectIDDest](s, src)

		require.NoError(t, err)
		assert.Empty(t, dst)
		assert.NotNil(t, dst)
	})

	t.Run("Complex Types", func(t *testing.T) {
		s := New()

		now := time.Now()
		items := []*ComplexSource{
			{
				ID:        primitive.NewObjectID(),
				UserID:    uuid.New(),
				Name:      "Alice",
				CreatedAt: now,
				Count:     1,
			},
			{
				ID:        primitive.NewObjectID(),
				UserID:    uuid.New(),
				Name:      "Bob",
				CreatedAt: now.Add(1 * time.Hour),
				Count:     2,
			},
		}

		dst, err := SerializeSlice[ComplexSource, ComplexDest](s, items)

		require.NoError(t, err)
		require.Len(t, dst, 2)

		assert.Equal(t, items[0].ID.Hex(), dst[0].ID)
		assert.Equal(t, items[0].UserID.String(), dst[0].UserID)
		assert.Equal(t, "Alice", dst[0].Name)
		assert.Equal(t, 1, dst[0].Count)

		assert.Equal(t, items[1].ID.Hex(), dst[1].ID)
		assert.Equal(t, items[1].UserID.String(), dst[1].UserID)
		assert.Equal(t, "Bob", dst[1].Name)
		assert.Equal(t, 2, dst[1].Count)
	})

	t.Run("Simple Types", func(t *testing.T) {
		s := New()

		objectID1 := primitive.NewObjectID()
		objectID2 := primitive.NewObjectID()
		objectID3 := primitive.NewObjectID()

		src := []*ObjectIDSource{
			{ID: objectID1},
			nil,
			{ID: objectID2},
			{ID: objectID3},
		}

		dst, err := SerializeSlice[ObjectIDSource, ObjectIDDest](s, src)

		require.NoError(t, err)
		require.Len(t, dst, 4)

		assert.Equal(t, objectID1.Hex(), dst[0].ID)
		assert.Nil(t, dst[1])
		assert.Equal(t, objectID2.Hex(), dst[2].ID)
		assert.Equal(t, objectID3.Hex(), dst[3].ID)
	})

	t.Run("Serialization Error", func(t *testing.T) {
		s := New()

		src := []*ObjectIDDest{
			{
				ID: "not-valid-oid",
			},
		}

		dst, err := SerializeSlice[ObjectIDDest, ObjectIDSource](s, src)
		require.Error(t, err)
		assert.Len(t, dst, 0)
	})
}

func TestSerializer(t *testing.T) {

	t.Run("Register Custom Converter", func(t *testing.T) {
		type Status int
		const (
			StatusActive Status = iota
			StatusInactive
		)

		type SourceWithEnum struct {
			Name   string
			Status Status
		}

		type DestWithString struct {
			Name   string
			Status string
		}

		s := New()

		// Register custom converter for Status -> string
		s.RegisterConverter(
			Status(0), // source type example
			"",        // destination type example
			func(src any) (any, error) {
				if status, ok := src.(Status); ok {
					switch status {
					case StatusActive:
						return "active", nil
					case StatusInactive:
						return "inactive", nil
					default:
						return "unknown", nil
					}
				}
				return nil, assert.AnError
			},
		)

		src := &SourceWithEnum{
			Name:   "Test",
			Status: StatusActive,
		}

		dst, err := Serialize[SourceWithEnum, DestWithString](s, src)

		require.NoError(t, err)
		require.NotNil(t, dst)
		assert.Equal(t, "Test", dst.Name)
		assert.Equal(t, "active", dst.Status)
	})

	t.Run("Converter Caching", func(t *testing.T) {
		s := New()

		// First serialization - should build option
		src1 := &ObjectIDSource{ID: primitive.NewObjectID()}
		dst1, err := Serialize[ObjectIDSource, ObjectIDDest](s, src1)
		require.NoError(t, err)
		require.NotNil(t, dst1)

		// Second serialization - should use cached option
		src2 := &ObjectIDSource{ID: primitive.NewObjectID()}
		dst2, err := Serialize[ObjectIDSource, ObjectIDDest](s, src2)
		require.NoError(t, err)
		require.NotNil(t, dst2)

		// Register new converter - should invalidate cache
		s.RegisterConverter(
			int(0),
			"",
			func(src any) (any, error) {
				return "converted", nil
			},
		)

		// Third serialization - should rebuild option with new converter
		src3 := &ObjectIDSource{ID: primitive.NewObjectID()}
		dst3, err := Serialize[ObjectIDSource, ObjectIDDest](s, src3)
		require.NoError(t, err)
		require.NotNil(t, dst3)
	})

}

func Benchmark(b *testing.B) {
	b.Run("Object Serialization", func(b *testing.B) {
		s := New()

		src := &ComplexSource{
			ID:        primitive.NewObjectID(),
			UserID:    uuid.New(),
			Name:      "Benchmark User",
			CreatedAt: time.Now(),
			Count:     100,
		}

		for b.Loop() {
			_, _ = Serialize[ComplexSource, ComplexDest](s, src)
		}
	})

	b.Run("Slice Serialization", func(b *testing.B) {
		s := New()

		items := make([]*ComplexSource, 100)
		for i := 0; i < 100; i++ {
			items[i] = &ComplexSource{
				ID:        primitive.NewObjectID(),
				UserID:    uuid.New(),
				Name:      "User",
				CreatedAt: time.Now(),
				Count:     i,
			}
		}

		for b.Loop() {
			_, _ = SerializeSlice[ComplexSource, ComplexDest](s, items)
		}
	})
}
