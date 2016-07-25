package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
)

// DocumentFile implements both Node and Handle for a document from a collection.
type DocumentFile struct {
	coll string
	Id   interface{} `bson:"_id"`

	Dirent fuse.Dirent
	Fattr  fuse.Attr

	CTime time.Time
	ATime time.Time
	MTime time.Time
}

func (d *DocumentFile) idQuery() bson.M {
	return bson.M{"_id": d.Id}
}

func (d *DocumentFile) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Printf("DocumentFile.Attr() for: %+v", d)
	_, size, err := d.readDocument()

	if err != nil {
		return err
	}

	if d.CTime.IsZero() {
		now := time.Now()
		d.CTime = now
		d.ATime = now
		d.MTime = now
	}

	a.Uid = uint32(os.Getuid())
	a.Gid = uint32(os.Getgid())
	a.Mode = 0600
	a.Size = size
	a.Ctime = d.CTime
	a.Atime = d.ATime
	a.Mtime = d.MTime
	return nil
}

func (d *DocumentFile) Lookup(ctx context.Context, fname string) (fs.Node, error) {
	log.Printf("DocumentFile[%s].Lookup(): %s\n", d.coll, fname)

	return nil, fuse.ENOENT
}

func (d *DocumentFile) ReadAll(ctx context.Context) ([]byte, error) {
	log.Printf("DocumentFile[%s].ReadAll(): %s\n", d.coll, d.Id)

	strval, _, err := d.readDocument()
	if err != nil {
		return nil, err
	}

	d.ATime = time.Now() // update last access time

	return []byte(strval), nil
}

// Read a document and return it as a JSON string
func (d *DocumentFile) readDocument() (string, uint64, error) {
	db, s := getDb()
	defer s.Close()

	var f interface{}

	err := db.C(d.coll).Find(d.idQuery()).One(&f)
	if err != nil {
		log.Fatal(err)
		return "", 0, fuse.EIO
	}

	buf, err := json.MarshalIndent(f, "", "    ")
	if err != nil {
		log.Fatal(err)
		return "", 0, fuse.EIO
	}

	strval := string(buf) + "\n"
	return strval, uint64(len(buf)), nil
}

func (d *DocumentFile) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Printf("DocumentFile.Write(%s) \n", d.Id)

	db, s := getDb()
	defer s.Close()

	doc := make(map[string]interface{})
	err := json.Unmarshal(req.Data, &doc)
	if err != nil {
		log.Printf("Could not parse the data as JSON[%s]: %s \n", d.Id, err.Error())
		return fuse.EIO
	}

	delete(doc, "_id") // _id cannot be updated!

	err = db.C(d.coll).Update(d.idQuery(), bson.M{"$set": doc})
	if err != nil {
		log.Printf("Could not update the document[%s]: %s \n", d.Id, err.Error())
		return fuse.EIO
	}
	d.MTime = time.Now() // update last modified time

	return nil
}
