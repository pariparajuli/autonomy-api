package store

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/bitmark-inc/autonomy-api/schema"
)

// Group - interface for finding group of people
type Group interface {
	NearestCount(int, schema.Location) ([]string, error)
	NearestDistance(int, schema.Location) ([]string, error)
}

// NearestCount - find nearest account number up to some number
// return matches by id
func (m *mongoDB) NearestCount(count int, loc schema.Location) ([]string, error) {
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)

	var record schema.Profile
	ids := make([]string, 0)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := nearQuery(loc)

	cur, err := c.Find(context.Background(), query)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("query nearest count %d with error: %s", count, err)
		return []string{}, nil
	}

	total := 0

	// iterate
	for cur.Next(ctx) && total < count {
		err = cur.Decode(&record)
		if nil != err {
			log.WithField("prefix", mongoLogPrefix).Errorf("nearest count decode record with error: %s", err)
			return []string{}, fmt.Errorf("decode mongo record with error: %s", err)
		}
		ids = append(ids, record.ID)
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("nearest count wants %d, get %d", count, total)

	return ids, nil
}

// NearestDistance - find nearest account number by distance
// return matches by account number
func (m *mongoDB) NearestDistance(distance int, cords schema.Location) ([]string, error) {
	query := distanceQuery(distance, cords)
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cur, err := c.Find(ctx, query)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("query nearest distance with error: %s", err)
		return []string{}, fmt.Errorf("nearest distance query with error: %s", err)
	}

	accountNumbers := make([]string, 0)
	var record schema.Profile

	// iterate
	for cur.Next(ctx) {
		err = cur.Decode(&record)
		if nil != err {
			log.WithField("prefix", mongoLogPrefix).Infof("query nearest distance with error: %s", err)
			return []string{}, fmt.Errorf("nearest distance query decode record with error: %s", err)
		}
		accountNumbers = append(accountNumbers, record.AccountNumber)
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("nearest distance query gets %d account number: %v", len(accountNumbers),
		accountNumbers)

	return accountNumbers, nil
}

func distanceQuery(distance int, cords schema.Location) bson.M {
	return bson.M{
		"location": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{cords.Longitude, cords.Latitude},
				},
				"$maxDistance": distance,
			},
		},
	}
}

// $hearSphere provides documents from nearest to farthest
// reference: https://docs.mongodb.com/manual/reference/operator/query/nearSphere/#op._S_nearSphere
func nearQuery(cords schema.Location) bson.M {
	return bson.M{
		"location": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{cords.Longitude, cords.Latitude},
				},
			},
		},
	}
}
