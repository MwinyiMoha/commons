package serializers

import (
	"reflect"
	"time"

	"github.com/mwinyimoha/commons/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func objectIDHook(from reflect.Type, to reflect.Type, data any) (any, error) {

	// ObjectID -> string
	if from == reflect.TypeOf(primitive.ObjectID{}) && to.Kind() == reflect.String {
		val, ok := data.(primitive.ObjectID)
		if ok {
			return val.Hex(), nil
		}
	}

	// string -> ObjectID
	if from.Kind() == reflect.String && to == reflect.TypeOf(primitive.ObjectID{}) {
		str, ok := data.(string)
		if ok {
			objectID, err := primitive.ObjectIDFromHex(str)
			if err != nil {
				return nil, errors.WrapError(err, errors.InvalidArgument, "invalid ObjectID: %s", str)
			}
			return objectID, nil
		}
	}

	return data, nil
}

func timestampHook(from reflect.Type, _ reflect.Type, data any) (any, error) {

	// time.Time -> *timestamppb.Timestamp
	f := from
	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}

	if f.AssignableTo(reflect.TypeOf(time.Time{})) {
		t, ok := data.(time.Time)
		if !ok {
			if tp, ok := data.(*time.Time); ok && tp != nil {
				t = *tp
			} else {
				return data, nil
			}
		}

		return timestamppb.New(t), nil
	}

	return data, nil
}
