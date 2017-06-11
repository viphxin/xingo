package mongo

import (
	"testing"
	"time"
	"gopkg.in/mgo.v2/bson"
	"errors"
	"gopkg.in/mgo.v2"
	"fmt"
	"encoding/gob"
	"bytes"
)

func Test_DailDB(t *testing.T){
	dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "", "")
	//dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "admin", "admin")
	t.Log("url: ", dbcfg.String())
	dbo := NewDbOperate(dbcfg, 5*time.Second)
	dbo.OpenDB(nil)
	dbo.CloseDB()
}
//go test -v github.com\viphxin\xingo\db\mongo -run ^Test_CommonOperate$
func Test_CommonOperate(t *testing.T){
	dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "", "")
	//dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "admin", "admin")
	t.Log("url: ", dbcfg.String())
	dbo := NewDbOperate(dbcfg, 5*time.Second)
	dbo.OpenDB(func(ms *mgo.Session){
		ms.DB("").C("test").EnsureIndex(mgo.Index{
			Key: []string{"username"},
			Unique: true,
		})
	})

	_, err := dbo.DeleteAll("test", bson.M{"pass": "pass1111"})
	if err != nil{
		//not do anything
	}
	//------------------------------------------------------------------------
	err = dbo.Insert("test", bson.M{"username": "xingo", "pass": "pass1111"})
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}

	err = dbo.Insert("test", bson.M{"username": "xingo_0", "pass": "pass1111"})
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}

	err = dbo.DBFindOne("test", bson.M{"username": "xingo"}, func(a bson.M)error{
		if a != nil{
			t.Log(a)
			return nil
		}else{
			dbo.CloseDB()
			t.Fatal("DBFindOne error")
			return errors.New("DBFindOne error")
		}

	})
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}
	_, err = dbo.DeleteAll("test", bson.M{"pass": "pass1111"})
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}
	//-------------------------------------------------------------------------
	//bulk
	docs := make([]bson.M, 0)
	for i := 0; i < 500; i++{
		docs = append(docs, bson.M{"username": fmt.Sprintf("xingo_%d", i), "pass": "pass1111"})
	}

	err = dbo.BulkInsert("test", docs)
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}

	_, err = dbo.DeleteAll("test", bson.M{"pass": "pass1111"})
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}

	dbo.CloseDB()
}

//go test -v github.com\viphxin\xingo\db\mongo -bench ^Benchmark_CommonOperate$
func Benchmark_CommonOperate(b *testing.B){
	dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "", "")
	//dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "admin", "admin")
	b.Log("url: ", dbcfg.String())
	dbo := NewDbOperate(dbcfg, 5*time.Second)
	dbo.OpenDB(func(ms *mgo.Session){
		ms.DB("").C("test").EnsureIndex(mgo.Index{
			Key: []string{"username"},
			Unique: true,
		})
	})

	for i := 0; i < b.N; i++ {
		_, err := dbo.DeleteAll("test", bson.M{"pass": "pass1111"})
		if err != nil {
			//not do anything
		}
		//------------------------------------------------------------------------
		err = dbo.Insert("test", bson.M{"username": "xingo", "pass": "pass1111"})
		if err != nil {
			dbo.CloseDB()
			b.Fatal(err)
			return
		}

		err = dbo.Insert("test", bson.M{"username": "xingo_0", "pass": "pass1111"})
		if err != nil {
			dbo.CloseDB()
			b.Fatal(err)
			return
		}

		err = dbo.DBFindOne("test", bson.M{"username": "xingo"}, func(a bson.M) error {
			if a != nil {
				b.Log(a)
				return nil
			} else {
				dbo.CloseDB()
				b.Fatal("DBFindOne error")
				return errors.New("DBFindOne error")
			}

		})
		if err != nil {
			dbo.CloseDB()
			b.Fatal(err)
			return
		}
		_, err = dbo.DeleteAll("test", bson.M{"pass": "pass1111"})
		if err != nil {
			dbo.CloseDB()
			b.Fatal(err)
			return
		}
		//-------------------------------------------------------------------------
		//bulk
		docs := make([]bson.M, 0)
		for i := 0; i < 500; i++ {
			docs = append(docs, bson.M{"username": fmt.Sprintf("xingo_%d", i), "pass": "pass1111"})
		}

		err = dbo.BulkInsert("test", docs)
		if err != nil {
			dbo.CloseDB()
			b.Fatal(err)
			return
		}

		_, err = dbo.DeleteAll("test", bson.M{"pass": "pass1111"})
		if err != nil {
			dbo.CloseDB()
			b.Fatal(err)
			return
		}
	}
	dbo.CloseDB()
}

//go test -v github.com\viphxin\xingo\db\mongo -bench ^Benchmark_CommonOperatePP$
func Benchmark_CommonOperatePP(b *testing.B){
	dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "", "")
	//dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "admin", "admin")
	b.Log("url: ", dbcfg.String())
	dbo := NewDbOperate(dbcfg, 5*time.Second)
	dbo.OpenDB(func(ms *mgo.Session){
		ms.DB("").C("test").DropIndex("username")
	})
	_, err := dbo.DeleteAll("test", bson.M{"pass": "pass1111"})
	if err != nil {
		dbo.CloseDB()
		b.Fatal(err)
		return
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			//------------------------------------------------------------------------
			err := dbo.Insert("test", bson.M{"username": "xingo", "pass": "pass1111"})
			if err != nil {
				dbo.CloseDB()
				b.Fatal(err)
				return
			}

			err = dbo.DBFindOne("test", bson.M{"username": "xingo"}, func(a bson.M) error {
				if a != nil {
					b.Log(a)
					return nil
				} else {
					dbo.CloseDB()
					b.Fatal("DBFindOne error")
					return errors.New("DBFindOne error")
				}

			})
			if err != nil {
				dbo.CloseDB()
				b.Fatal(err)
				return
			}
			//-------------------------------------------------------------------------
			//bulk
			docs := make([]bson.M, 0)
			for i := 0; i < 500; i++ {
				docs = append(docs, bson.M{"username": fmt.Sprintf("xingo_%d", i), "pass": "pass1111"})
			}

			err = dbo.BulkInsert("test", docs)
			if err != nil {
				dbo.CloseDB()
				b.Fatal(err)
				return
			}
		}
	})
	dbo.CloseDB()
}

func Test_GrdiFS(t *testing.T){
	dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "", "")
	//dbcfg := NewDbCfg("127.0.0.1", 27017, "xingodb", "admin", "admin")
	t.Log("url: ", dbcfg.String())
	dbo := NewDbOperate(dbcfg, 5*time.Second)
	dbo.OpenDB(func(ms *mgo.Session){
		ms.DB("").C("test.gfs").EnsureIndexKey("filename")
	})
	_, err := dbo.RemoveGridFile("test.gfs", "mmomap.db")
	if err != nil {
		dbo.CloseDB()
		t.Fatal(err)
		return
	}

	type player struct {
		UserId int64
		Daomond int64
	}
	type playerQueue struct {
		Q []player
	}
	q := &playerQueue{Q: make([]player, 0)}
	for i := int64(0); i < 1000; i++{
		q.Q = append(q.Q, player{i, 1000})
	}
	//create gridfs
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	enc.Encode(q)
	err = dbo.WriteGridFile("test.gfs", "mmomap.db", buff.Bytes())
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}
	//read gridfs
	data, err := dbo.OpenGridFile("test.gfs", "mmomap.db")
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}
	buff.Reset()
	buff.Write(data)
	dec := gob.NewDecoder(buff)
	qq := &playerQueue{Q: make([]player, 0)}
	err = dec.Decode(qq)
	if err != nil{
		dbo.CloseDB()
		t.Fatal(err)
		return
	}else{
		t.Log("success!!!!!!!!!!!!, Q len: ", len(qq.Q))
	}
	dbo.CloseDB()
}