package store

import (
	"context"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/bitmark-inc/autonomy-api/schema"
)

// Healthier - interface to get average health score by ids
type Healthier interface {
	Health([]string) (float64, error)
}

// Health - average health score by uuid
func (m *mongoDB) Health(ids []string) (float64, error) {
	c := m.client.Database(m.database).Collection(schema.ProfileCollectionName)
	cur, err := c.Find(context.Background(), healthQuery(ids))
	if nil != err {
		return float64(0), err
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("achieve group health score from ids: %v", ids)

	var sum float64
	var count int
	var p schema.Profile
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	for cur.Next(ctx) {
		err := cur.Decode(&p)
		if nil != err {
			return float64(0), err
		}

		sum += p.HealthScore
		count++
	}

	if count == 0 {
		return 0, nil
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("health score %f, total: %d", sum, count)

	return sum / float64(count), nil
}

func healthQuery(ids []string) bson.D {
	arr := bson.A{}
	for _, id := range ids {
		arr = append(arr, id)
	}

	return bson.D{{
		"id",
		bson.D{{
			"$in",
			arr,
		}},
	}}
}
