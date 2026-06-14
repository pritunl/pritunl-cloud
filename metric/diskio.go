package metric

import (
	"context"
	"time"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/mongo-go-driver/v2/mongo/options"
	"github.com/pritunl/pritunl-cloud/database"
)

type DiskIo struct {
	Id        bson.ObjectID `bson:"_id" json:"id"`
	Resource  bson.ObjectID `bson:"r" json:"r"`
	Timestamp time.Time     `bson:"t" json:"t"`

	Disks []*DiskIoDisk `bson:"d" json:"d"`
}

type DiskIoDisk struct {
	Node       string `bson:"n" json:"n"`
	BytesRead  uint64 `bson:"br" json:"br"`
	BytesWrite uint64 `bson:"bw" json:"bw"`
	CountRead  uint64 `bson:"cr" json:"cr"`
	CountWrite uint64 `bson:"cw" json:"cw"`
	TimeRead   uint64 `bson:"tr" json:"tr"`
	TimeWrite  uint64 `bson:"tw" json:"tw"`
	TimeIo     uint64 `bson:"ti" json:"ti"`
}

type DiskStatic struct {
	Node string `bson:"node" json:"node"`
}

func ParseDisk(i *DiskIoDisk) *DiskStatic {
	return &DiskStatic{
		Node: i.Node,
	}
}

type DiskIoAgg struct {
	Id struct {
		Node      string `bson:"n"`
		Timestamp int64  `bson:"t"`
	} `bson:"_id"`
	BytesRead  uint64 `bson:"br"`
	BytesWrite uint64 `bson:"bw"`
	CountRead  uint64 `bson:"cr"`
	CountWrite uint64 `bson:"cw"`
	TimeRead   uint64 `bson:"tr"`
	TimeWrite  uint64 `bson:"tw"`
	TimeIo     uint64 `bson:"ti"`
}

func (d *DiskIo) GetCollection(db *database.Database) *database.Collection {
	return db.MetricsDiskIo()
}

func (d *DiskIo) Format(id bson.ObjectID) time.Time {
	d.Resource = id
	d.Timestamp = d.Timestamp.UTC().Truncate(1 * time.Minute)
	d.Id = GenerateId(id, d.Timestamp)
	return d.Timestamp
}

func (d *DiskIo) StaticData() bson.M {
	disks := []*DiskStatic{}

	for _, dsk := range d.Disks {
		disks = append(disks, ParseDisk(dsk))
	}

	return bson.M{
		"disks": disks,
	}
}

func GetDiskIoChartSingle(c context.Context, db *database.Database,
	resource bson.ObjectID, start, end time.Time) (
	chartData ChartData, err error) {

	coll := db.MetricsDiskIo()
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
		doc := &DiskIo{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		for _, dsk := range doc.Disks {
			timestamp := doc.Timestamp.UnixMilli()

			chart.Add(dsk.Node+"-br", timestamp, dsk.BytesRead)
			chart.Add(dsk.Node+"-bw", timestamp, dsk.BytesWrite)
			//chart.Add(dsk.Node+"-cr", timestamp, dsk.CountRead)
			//chart.Add(dsk.Node+"-cw", timestamp, dsk.CountWrite)
			chart.Add(dsk.Node+"-tr", timestamp, dsk.TimeRead)
			chart.Add(dsk.Node+"-tw", timestamp, dsk.TimeWrite)
			chart.Add(dsk.Node+"-ti", timestamp, dsk.TimeIo)
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

func GetDiskIoChart(c context.Context, db *database.Database,
	resource bson.ObjectID, start, end time.Time,
	interval time.Duration) (chartData ChartData, err error) {

	if interval == 1*time.Minute {
		chartData, err = GetDiskIoChartSingle(c, db, resource, start, end)
		return
	}

	coll := db.MetricsDiskIo()
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
			"$unwind": "$d",
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
					{"n", "$d.n"},
				},
				"br": &bson.D{
					{"$sum", "$d.br"},
				},
				"bw": &bson.D{
					{"$sum", "$d.bw"},
				},
				//"cr": &bson.D{
				//	{"$sum", "$d.cr"},
				//},
				//"cw": &bson.D{
				//	{"$sum", "$d.cw"},
				//},
				"tr": &bson.D{
					{"$sum", "$d.tr"},
				},
				"tw": &bson.D{
					{"$sum", "$d.tw"},
				},
				"ti": &bson.D{
					{"$sum", "$d.ti"},
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
		doc := &DiskIoAgg{}
		err = cursor.Decode(doc)
		if err != nil {
			err = database.ParseError(err)
			return
		}

		chart.Add(doc.Id.Node+"-br", doc.Id.Timestamp, doc.BytesRead)
		chart.Add(doc.Id.Node+"-bw", doc.Id.Timestamp, doc.BytesWrite)
		//chart.Add(doc.Id.Node+"-cr", doc.Id.Timestamp, doc.CountRead)
		//chart.Add(doc.Id.Node+"-cw", doc.Id.Timestamp, doc.CountWrite)
		chart.Add(doc.Id.Node+"-tr", doc.Id.Timestamp, doc.TimeRead)
		chart.Add(doc.Id.Node+"-tw", doc.Id.Timestamp, doc.TimeWrite)
		chart.Add(doc.Id.Node+"-ti", doc.Id.Timestamp, doc.TimeIo)
	}

	err = cursor.Err()
	if err != nil {
		err = database.ParseError(err)
		return
	}

	chartData = chart.Export()

	return
}
