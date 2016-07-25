package main

import (
	"log"
	"os"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
)

// CollFile implements both Node and Handle for a collection.
type CollFile struct {
	Name string

	Dirent fuse.Dirent
	Fattr  fuse.Attr
}

func idQuery(id string) bson.M {
	if bson.IsObjectIdHex(id) {
		return bson.M{"_id": bson.ObjectIdHex(id)}
	} else {
		return bson.M{"_id": id}
	}
}

func (f *CollFile) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Printf("CollFile.Attr() for: %+v", f)
	a.Mode = os.ModeDir | 0700
	a.Uid = uint32(os.Getuid())
	a.Gid = uint32(os.Getgid())
	return nil
}

func (c *CollFile) Lookup(ctx context.Context, fname string) (fs.Node, error) {
	log.Printf("CollFile[%s].Lookup(): %s\n", c.Name, fname)

	if !strings.HasSuffix(fname, ".json") {
		return nil, fuse.ENOENT
	}
	fname = fname[0 : len(fname)-5]

	db, s := getDb()
	defer s.Close()

	var f DocumentFile
	err := db.C(c.Name).Find(idQuery(fname)).One(&f)
	if err != nil {
		log.Printf("Error while looking up %s: %s \n", fname, err.Error())
		return nil, fuse.EIO
	}

	f.coll = c.Name

	return &f, nil
}

func (c *CollFile) ReadDirAll(ctx context.Context) (ents []fuse.Dirent, ferr error) {
	log.Println("CollFile.ReadDirAll(): ", c.Name)

	db, s := getDb()
	defer s.Close()

	iter := db.C(c.Name).Find(nil).Iter()

	var f DocumentFile
	for iter.Next(&f) {
		if strId, ok := f.Id.(string); ok {
			ents = append(ents, fuse.Dirent{Name: strId + ".json", Type: fuse.DT_File})
		} else {
			ents = append(ents, fuse.Dirent{Name: f.Id.(bson.ObjectId).Hex() + ".json", Type: fuse.DT_File})
		}
	}

	if err := iter.Err(); err != nil {
		log.Panic(err)
		return nil, fuse.EIO
	}

	return ents, nil
}

func (c *CollFile) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	log.Printf("DocumentFile.Remove(): %s/%s \n", c.Name, req.Name)

	if !strings.HasSuffix(req.Name, ".json") {
		return fuse.ENOENT
	}
	id := req.Name[0 : len(req.Name)-5]

	db, s := getDb()
	defer s.Close()

	if err := db.C(c.Name).Remove(idQuery(id)); err != nil {
		log.Printf("Could not remove document '%s': %s \n", id, err.Error())
		return fuse.EIO
	}

	return nil
}
