package schema

import (
	"context"
	"testing"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/vmihailenco/msgpack/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v2"
)

var profile *schema.Profile

func LoadProfile() bool {
	ctx := context.Background()
	opts := options.Client().ApplyURI("mongodb://127.0.0.1:27017/?compressors=disabled")
	mongoClient, err := mongo.NewClient(opts)
	if nil != err {
		panic(err)
	}

	if err = mongoClient.Connect(ctx); nil != err {
		panic(err)
	}

	if profile == nil {
		var p schema.Profile
		if err := mongoClient.Database("test-db").Collection("profile").FindOne(ctx, bson.M{}).Decode(&p); err != nil {
			panic(err)
		}
		profile = &p
		return true
	}
	return false
}
func BenchmarkDecodeYAML(b *testing.B) {
	if LoadProfile() {
		b.ResetTimer()
	}

	data, err := yaml.Marshal(profile)
	if err != nil {
		b.Fatal(err)
	}

	var p schema.Profile
	for n := 0; n < b.N; n++ {
		if err := yaml.Unmarshal(data, &p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeYAML(b *testing.B) {
	if LoadProfile() {
		b.ResetTimer()
	}

	for n := 0; n < b.N; n++ {
		if _, err := yaml.Marshal(profile); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeMsgPack(b *testing.B) {
	if LoadProfile() {
		b.ResetTimer()
	}

	data, err := msgpack.Marshal(profile)
	if err != nil {
		b.Fatal(err)
	}

	var p schema.Profile
	for n := 0; n < b.N; n++ {
		if err := msgpack.Unmarshal(data, &p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeMsgPack(b *testing.B) {
	if LoadProfile() {
		b.ResetTimer()
	}

	for n := 0; n < b.N; n++ {
		if _, err := msgpack.Marshal(profile); err != nil {
			b.Fatal(err)
		}
	}
}
