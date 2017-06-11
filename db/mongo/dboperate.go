package mongo

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"github.com/viphxin/xingo/logger"
)

const (
	Strong = 1
	Monotonic = 2
)

var (
	XINGO_MONGODB_SESSION_NIL_ERR    = errors.New("DbOperate session nil.")
	XINGO_MONGODB_NOTFOUND_ERR       = errors.New("not found!")
	XINGO_MONGODB_DBFINDALL_ERR      = errors.New("DBFindAll failed,q is nil!")
	XINGO_MONGODB_OPENGRIDFILE_ERR   =  errors.New("OpenGridFile failed!")
	XINGO_MONGODB_READGRIDFILE_ERR   =  errors.New("ReadGridFile failed!")
	XINGO_MONGODB_CREATEGRIDFILE_ERR   =  errors.New("CreateGridFile is nil")
)

type DbCfg struct {
	DbHost string
	DbPort int
	DbName string
	DbUser string
	DbPass string
}

func NewDbCfg(host string, port int, name , user, pass string) *DbCfg{
	return &DbCfg{
		DbHost: host,
		DbPort: port,
		DbName: name,
		DbUser: user,
		DbPass: pass,
	}
}

func (this *DbCfg)String() string{
	url := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		this.DbUser, this.DbPass, this.DbHost, this.DbPort, this.DbName)
	if this.DbUser == "" || this.DbPass == "" {
		url = fmt.Sprintf("mongodb://%s:%d/%s", this.DbHost, this.DbPort, this.DbName)
	}
	return url
}

type DbOperate struct {
	session      *mgo.Session
	timeout      time.Duration
	dbcfg        *DbCfg
}

func NewDbOperate(dbcfg *DbCfg, timeout time.Duration) *DbOperate{
	return &DbOperate{nil, timeout, dbcfg}
}

func (db *DbOperate) GetDbSession() *mgo.Session {
	return db.session
}

func (this *DbOperate) SetMode(mode int, refresh bool) {
	status := mgo.Monotonic
	if mode == Strong {
		status = mgo.Strong
	} else {
		status = mgo.Monotonic
	}

	this.session.SetMode(status, refresh)
}

func (this *DbOperate) OpenDB(set_index_func func(ms *mgo.Session)) error {
	logger.Info(fmt.Sprintf("DbOperate mongodb connect url: %s\n", this.dbcfg.String()))

	var err error
	this.session, err = mgo.DialWithTimeout(this.dbcfg.String(), this.timeout)
	if err != nil {
		panic(err.Error())
	}

	this.session.SetMode(mgo.Monotonic, true)
	//set index
	if set_index_func != nil{
		set_index_func(this.session)
	}
	logger.Info(fmt.Sprintf("DbOperate Connect %v mongodb...OK", this.dbcfg.String()))
	return nil
}

func (this *DbOperate) CloseDB() {
	if this.session != nil {
		this.session.DB("").Logout()
		this.session.Close()
		this.session = nil
		logger.Info("Disconnect mongodb url: ", this.dbcfg.String())
	}
}

func (this *DbOperate) RefreshSession() {
	this.session.Refresh()

}

func (this *DbOperate) Insert(collection string, doc interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)

	return c.Insert(doc)
}

func (this *DbOperate) StrongInsert(collection string, doc interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)

	c := local_session.DB("").C(collection)

	return c.Insert(doc)
}

func (this *DbOperate) Cover(collection string, cond interface{}, change interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)

	return c.Update(cond, change)
}

func (this *DbOperate) Update(collection string, cond interface{}, change interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)

	return c.Update(cond, bson.M{"$set": change})
}

func (this *DbOperate) StrongUpdate(collection string, cond interface{}, change interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)
	c := local_session.DB("").C(collection)

	return c.Update(cond, bson.M{"$set": change})
}

func (this *DbOperate) UpdateInsert(collection string, cond interface{}, doc interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)
	_, err := c.Upsert(cond, bson.M{"$set": doc})
	if err != nil {
		logger.Error(fmt.Sprintf("UpdateInsert failed collection is:%s. cond is:%v", collection, cond))
	}

	return err
}

func (this *DbOperate) StrongUpdateInsert(collection string, cond interface{}, doc interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)

	c := local_session.DB("").C(collection)
	_, err := c.Upsert(cond, bson.M{"$set": doc})
	if err != nil {
		logger.Error(fmt.Sprintf("UpdateInsert failed collection is:%s. cond is:%v", collection, cond))
	}

	return err
}

func (this *DbOperate) RemoveOne(collection string, cond_name string, cond_value int64) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)
	err := c.Remove(bson.M{cond_name: cond_value})
	if err != nil && err != mgo.ErrNotFound {
		logger.Error(fmt.Sprintf("remove failed from collection:%s. name:%s-value:%d", collection, cond_name, cond_value))
	}

	return err

}

func (this *DbOperate) RemoveOneByCond(collection string, cond interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)
	err := c.Remove(cond)

	if err != nil && err != mgo.ErrNotFound {
		logger.Error(fmt.Sprintf("remove failed from collection:%s. cond :%v, err: %v.", collection, cond, err))
	}

	return err

}

func (this *DbOperate) RemoveAll(collection string, cond interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)
	change, err := c.RemoveAll(cond)
	if err != nil && err != mgo.ErrNotFound {
		logger.Error(fmt.Sprintf("DbOperate.RemoveAll failed : %s, %v", collection, cond))
		return err
	}
	logger.Debug(fmt.Sprintf("DbOperate.RemoveAll: %v, %v", change.Updated, change.Removed))
	return nil
}

func (this *DbOperate) DBFindOne(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)
	q := c.Find(cond)

	m := make(bson.M)
	if err := q.One(m); err != nil {
		if mgo.ErrNotFound != err {
			logger.Error(fmt.Sprintf("DBFindOne query falied,return error: %v; name: %v.", err, collection))
		}
		return err
	}

	if nil != resHandler {
		return resHandler(m)
	}

	return nil

}

func (this *DbOperate) StrongDBFindOne(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)

	c := local_session.DB("").C(collection)
	q := c.Find(cond)

	m := make(bson.M)
	if err := q.One(m); err != nil {
		if mgo.ErrNotFound != err {
			logger.Error(fmt.Sprintf("DBFindOne query falied, return error: %v; name: %v.", err, collection))
		}
		return err
	}

	if nil != resHandler {
		return resHandler(m)
	}

	return nil

}

func (this *DbOperate) DBFindAll(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)
	q := c.Find(cond)

	if nil == q {
		return XINGO_MONGODB_DBFINDALL_ERR
	}

	iter := q.Iter()
	m := make(bson.M)
	for iter.Next(m) == true {
		if nil != resHandler {
			err := resHandler(m)
			if err != nil {
				logger.Error(fmt.Sprintf("resHandler error :%v!!!", err))
				return err
			}
		}
	}

	return nil

}

func (this *DbOperate) StrongDBFindAll(collection string, cond interface{}, resHandler func(bson.M) error) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)

	c := local_session.DB("").C(collection)
	q := c.Find(cond)

	logger.Debug(fmt.Sprintf("[DbOperate.DBFindAll] name:%s,query:%v", collection, cond))

	if nil == q {
		return XINGO_MONGODB_DBFINDALL_ERR
	}
	iter := q.Iter()
	m := make(bson.M)
	for iter.Next(m) == true {
		if resHandler != nil {
			err := resHandler(m)
			if err != nil {
				logger.Error(fmt.Sprintf("resHandler error :%v!!!", err))
				return err
			}
		}
	}

	return nil
}

func (this *DbOperate) DBFindAllEx(collection string, cond interface{}, resHandler func(*mgo.Query) error) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)
	q := c.Find(cond)

	logger.Error(fmt.Sprintf("[DbOperate.DBFindAll] name:%s,query:%v", collection, cond))

	if nil == q {
		return XINGO_MONGODB_DBFINDALL_ERR
	}
	if nil != resHandler {
		return resHandler(q)
	}
	return nil
}

func (this *DbOperate) StrongDBFindAllEx(collection string, cond interface{}, resHandler func(*mgo.Query) error) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)

	c := local_session.DB("").C(collection)
	q := c.Find(cond)

	logger.Debug(fmt.Sprintf("[DbOperate.DBFindAll] name:%s,query:%v", collection, cond))

	if nil == q {
		return XINGO_MONGODB_DBFINDALL_ERR
	}
	if nil != resHandler {
		return resHandler(q)
	}
	return nil
}

func (this *DbOperate) FindAndModify(collection string, cond interface{}, change mgo.Change, val interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)
	c := local_session.DB("").C(collection)
	_, err := c.Find(cond).Apply(change, val)
	return err
}

func (this *DbOperate) FindAll(collection string, cond interface{}, all interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)
	c := local_session.DB("").C(collection)
	err := c.Find(cond).All(all)
	return err
}

func (this *DbOperate) StrongBatchInsert(collection string, docs ...interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	local_session.SetMode(mgo.Strong, true)
	c := local_session.DB("").C(collection)
	return c.Insert(docs...)
}

func (this *DbOperate) FindOne(collection string, cond interface{}, value interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()
	local_session.SetMode(mgo.Strong, true)
	c := local_session.DB("").C(collection)
	return c.Find(cond).One(value)
}

//友情提示，如果存在多个document，会报错，请用DeleteAll
func (this *DbOperate) DeleteOne(collection string, cond interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()
	local_session.SetMode(mgo.Strong, true)
	c := local_session.DB("").C(collection)
	return c.Remove(cond)
}

func (this *DbOperate) DeleteAll(collection string, cond interface{}) (int, error) {
	if this.session == nil {
		return 0, XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()
	local_session.SetMode(mgo.Strong, true)
	c := local_session.DB("").C(collection)
	changeInfo, err := c.RemoveAll(cond)
	if err != nil{
		return 0, err
	}
	return changeInfo.Removed, nil
}

// gridfs
func (this *DbOperate) OpenGridFile(collection string, filename string) ([]byte, error) {
	if this.session == nil {
		return nil, XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	gfs := local_session.DB("").GridFS(collection)
	// open file only for reading
	fsfile, err := gfs.Open(filename)

	if err != nil {
		return nil, err
	}

	if fsfile == nil {
		return nil, XINGO_MONGODB_OPENGRIDFILE_ERR
	}
	defer fsfile.Close()

	data := make([]byte, fsfile.Size())
	_, err = fsfile.Read(data)

	if err != nil {
		return nil, XINGO_MONGODB_READGRIDFILE_ERR
	}
	return data, nil
}

type gfsDocId struct {
	Id interface{} "_id"
}

func (this *DbOperate) CreateGridFile(collection string, filename string, resHandler func(*mgo.GridFile) error) error {

	var doc gfsDocId
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	gfs := local_session.DB("").GridFS(collection)

	file, err := gfs.Create(filename)
	if err != nil {
		return err
	}
	if file == nil {
		return XINGO_MONGODB_CREATEGRIDFILE_ERR
	}
	if resHandler != nil{
		err = resHandler(file)
		if err != nil {
			return err
		}
	}

	newfileId := file.Id()
	file.Close()

	query := gfs.Files.Find(bson.M{"filename": filename, "_id": bson.M{"$ne": newfileId}})
	iter := query.Iter()
	for iter.Next(&doc) {
		if e := gfs.RemoveId(doc.Id); e != nil {
			err = e
			break
		}
	}

	return err
}

func (this *DbOperate) GridFileExists(collection string, filename string) (bool, error) {
	if this.session == nil {
		return false, XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	gfs := local_session.DB("").GridFS(collection)

	query := gfs.Files.Find(bson.M{"filename": filename})
	m := make(bson.M)
	if err := query.One(m); err != nil {
		if mgo.ErrNotFound != err {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func (this *DbOperate) RemoveGridFile(collection string, filename string) (bool, error) {
	if this.session == nil {
		return false, XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	gfs := local_session.DB("").GridFS(collection)

	if err := gfs.Remove(filename); err != nil {
		if mgo.ErrNotFound != err {
			return false, err
		}
	}

	return true, nil
}

func (this *DbOperate) BulkInsertDoc(collection string, docs []interface{}) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)

	bulk := c.Bulk()

	bulk.Insert(docs...)
	_, err := bulk.Run()
	return err
}

func (this *DbOperate) BulkInsert(collection string, pairs []bson.M) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)

	bulk := c.Bulk()

	for i := 0; i < len(pairs); i++ {
		bulk.Insert(pairs[i])
	}
	_, err := bulk.Run()
	return err
}
func (this *DbOperate) BulkUpdate(collection string, pairs []bson.M) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)

	bulk := c.Bulk()

	for i := 0; i < len(pairs); i += 2 {
		selector := pairs[i]
		update := pairs[i+1]
		bulk.Update(selector, update)
	}
	_, err := bulk.Run()
	return err
}

func (this *DbOperate) GetMaxId(collection string, field string) (int64, error) {
	var id int64
	var present bool

	fnc := func(mq *mgo.Query) error {
		q := mq.Sort("-" + field).Limit(1)
		m := make(bson.M)
		if err := q.One(m); err != nil {
			id = 0
		}
		id, present = m[field].(int64)
		if !present {
			id = 0
		}
		return nil
	}
	err := this.DBFindAllEx(collection, nil, fnc)
	if nil != err {
		return 0, nil
	}
	return id, nil
}

func (this *DbOperate) WriteGridFile(collection string, filename string, data []byte) error {
	logger.Debug(fmt.Sprintf("Write grid file: %v, size %v.", filename, len(data)))

	fnc := func(file *mgo.GridFile) error {
		_, err := file.Write(data)
		return err
	}

	err := this.CreateGridFile(collection, filename, fnc)
	if err != nil {
		logger.Error("Write grid file fail: ", err)
	}
	return err
}

func (this *DbOperate) BulkUpsert(collection string, pairs []bson.M) error {
	if this.session == nil {
		return XINGO_MONGODB_SESSION_NIL_ERR
	}

	local_session := this.session.Copy()
	defer local_session.Close()

	c := local_session.DB("").C(collection)

	bulk := c.Bulk()

	for i := 0; i < len(pairs); i += 2 {
		selector := pairs[i]
		update := pairs[i+1]
		bulk.Upsert(selector, update)
	}
	_, err := bulk.Run()
	return err
}