package metric

import (
	"context"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

type Disk struct {
	Id        bson.ObjectID `bson:"_id" json:"id"`
	Resource  bson.ObjectID `bson:"r" json:"r"`
	Timestamp time.Time     `bson:"t" json:"t"`

	Mounts []*Mount `bson:"m" json:"m"`
}

type Mount struct {
	Mount  string  `bson:"m" json:"m"`
	Used   float64 `bson:"u" json:"u"`
	Size   uint64  `bson:"-" json:"s"`
	Format string  `bson:"-" json:"f"`
}

type MountStatic struct {
	Mount  string  `bson:"mount" json:"mount"`
	Used   float64 `bson:"used" json:"used"`
	Size   uint64  `bson:"size" json:"size"`
	Format string  `bson:"format" json:"format"`
}

func ParseMount(mn *Mount) *MountStatic {
	return &MountStatic{
		Mount:  mn.Mount,
		Used:   mn.Used,
		Size:   mn.Size,
		Format: mn.Format,
	}
}

type DiskAgg struct {
	Id struct {
		Path      string `bson:"p"`
		Timestamp int64  `bson:"t"`
	} `bson:"_id"`
	Used float64 `bson:"u"`
}

func (d *Disk) GetCollection(db *database.Database) *database.Collection {
	return db.MetricsDisk()
}

func (d *Disk) Format(id bson.ObjectID) time.Time {
	d.Resource = id
	d.Timestamp = d.Timestamp.UTC().Truncate(1 * time.Minute)
	d.Id = GenerateId(id, d.Timestamp)
	return d.Timestamp
}

func (d *Disk) StaticData() bson.M {
	mounts := []*MountStatic{}

	for _, mount := range d.Mounts {
		mounts = append(mounts, ParseMount(mount))
	}

	return bson.M{
		"mounts": mounts,
	}
}

func GetDiskChartSingle(c context.Context, db *database.Database,
	resource bson.ObjectID, start, end time.Time) (
	chartData ChartData, err error) {

	coll := db.MetricsDisk()
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
		doc := &Disk{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		timestamp := doc.Timestamp.UnixMilli()
		for _, mount := range doc.Mounts {
			chart.Add(mount.Mount, timestamp, mount.Used)
		}
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	chartData = chart.Export()

	return
}

func GetDiskChart(c context.Context, db *database.Database,
	resource bson.ObjectID, start, end time.Time,
	interval time.Duration) (chartData ChartData, err error) {

	if interval == 1*time.Minute {
		chartData, err = GetDiskChartSingle(c, db, resource, start, end)
		return
	}

	coll := db.MetricsDisk()
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
			"$unwind": "$m",
		},
		&bson.M{
			"$group": &bson.M{
				"_id": &bson.D{
					{"t", &bson.M{
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
					}},
					{"p", "$m.p"},
				},
				"u": &bson.D{
					{"$avg", "$m.u"},
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
		doc := &DiskAgg{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		chart.Add(doc.Id.Path, doc.Id.Timestamp, doc.Used)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	chartData = chart.Export()

	return
}
