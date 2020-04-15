package schema

type Confirm struct {
	County        string `bson:"county"`
	Count         int    `bson:"count"`
	Country       string `bson:"country"`
	UpdateTime    int64  `bson:"update_time"`
	DiffYesterday int    `bson:"diff_yesterday"`
}
