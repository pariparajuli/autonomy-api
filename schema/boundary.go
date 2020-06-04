package schema

const (
	BoundaryCollection = "boundary"
)

type Geometry struct {
	Type        string      `bson:"type"`
	Coordinates interface{} `bson:"coordinates"`
}

type Boundary struct {
	Country  string   `bson:"country"`
	Island   string   `bson:"island"`
	State    string   `bson:"state"`
	County   string   `bson:"county"`
	Geometry Geometry `bson:"geometry"`
}
