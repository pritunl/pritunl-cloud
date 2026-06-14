package metric

import (
	"context"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

type System struct {
	Id        bson.ObjectID `bson:"_id" json:"id"`
	Resource  bson.ObjectID `bson:"r" json:"r"`
	Timestamp time.Time     `bson:"t" json:"t"`

	Processes uint64  `bson:"pc" json:"pc"`
	CpuCores  int     `bson:"-" json:"cc"`
	CpuUsage  float64 `bson:"cu" json:"cu"`
	MemTotal  int     `bson:"-" json:"mt"`
	MemUsage  float64 `bson:"mu" json:"mu"`
	HugeTotal int     `bson:"-" json:"ht"`
	HugeUsage float64 `bson:"hu" json:"hu"`
	SwapTotal int     `bson:"-" json:"st"`
	SwapUsage float64 `bson:"su" json:"su"`
}

type SystemAgg struct {
	Id        int64   `bson:"_id"`
	CpuUsage  float64 `bson:"cu"`
	MemUsage  float64 `bson:"mu"`
	SwapUsage float64 `bson:"su"`
	HugeUsage float64 `bson:"hu"`
}

func (d *System) GetCollection(db *database.Database) *database.Collection {
	return db.MetricsSystem()
}

func (d *System) Format(id bson.ObjectID) time.Time {
	d.Resource = id
	d.Timestamp = d.Timestamp.UTC().Truncate(1 * time.Minute)
	d.Id = GenerateId(id, d.Timestamp)
	return d.Timestamp
}

func (d *System) StaticData() bson.M {
	return bson.M{
		"memory":    d.MemUsage,
		"swap":      d.SwapUsage,
		"hugepages": d.HugeUsage,
	}
}

func GetSystemChartSingle(c context.Context, db *database.Database,
	resource bson.ObjectID, start, end time.Time) (
	chartData ChartData, err error) {

	coll := db.MetricsSystem()
	chart := NewChart(start, end, time.Minute)

	timeQuery := bson.D{
		{"$gte", start},
	}
	if !end.IsZero() {
		timeQuery = append(timeQuery, bson.E{"$lte", end})
	}

	cursor, err := coll.Find(
		c,
		bson.M{
			"r": resource,
			"t": timeQuery,
		},
		options.Find().
			SetSort(bson.D{{"t", 1}}),
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(c)

	for cursor.Next(c) {
		doc := &System{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		timestamp := doc.Timestamp.UnixMilli()

		chart.Add("cpu_usage", timestamp, doc.CpuUsage)
		chart.Add("mem_usage", timestamp, doc.MemUsage)
		chart.Add("swap_usage", timestamp, doc.SwapUsage)
		chart.Add("huge_usage", timestamp, doc.HugeUsage)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	chartData = chart.Export()

	return
}

func GetSystemChart(c context.Context, db *database.Database,
	resource bson.ObjectID, start, end time.Time,
	interval time.Duration) (chartData ChartData, err error) {

	if interval == 1*time.Minute {
		chartData, err = GetSystemChartSingle(c, db, resource, start, end)
		return
	}

	coll := db.MetricsSystem()
	chart := NewChart(start, end, interval)

	timeQuery := bson.D{
		{"$gte", start},
	}
	if !end.IsZero() {
		timeQuery = append(timeQuery, bson.E{"$lte", end})
	}

	cursor, err := coll.Aggregate(c, []*bson.M{
		&bson.M{
			"$match": &bson.M{
				"r": resource,
				"t": timeQuery,
			},
		},
		&bson.M{
			"$group": &bson.M{
				"_id": &bson.M{
					"$let": &bson.M{
						"vars": &bson.M{
							"t": &bson.D{{"$toLong", "$t"}},
						},
						"in": &bson.M{
							"$subtract": &bson.A{
								"$$t",
								&bson.M{
									"$mod": &bson.A{
										"$$t",
										interval.Milliseconds(),
									},
								},
							},
						},
					},
				},
				"cu": &bson.D{
					{"$avg", "$cu"},
				},
				"mu": &bson.D{
					{"$avg", "$mu"},
				},
				"su": &bson.D{
					{"$avg", "$su"},
				},
				"hu": &bson.D{
					{"$avg", "$hu"},
				},
			},
		},
		&bson.M{
			"$sort": &bson.M{
				"_id": 1,
			},
		},
	})
	if err != nil {
		err = database.ParseError(err)
		return
	}
	defer cursor.Close(c)

	for cursor.Next(c) {
		doc := &SystemAgg{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		chart.Add("cpu_usage", doc.Id, doc.CpuUsage)
		chart.Add("mem_usage", doc.Id, doc.MemUsage)
		chart.Add("swap_usage", doc.Id, doc.SwapUsage)
		chart.Add("huge_usage", doc.Id, doc.HugeUsage)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	chartData = chart.Export()

	return
}
