package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
)

// GridFsFile implements both Node and Handle for a document from a collection.
type GridFsFile struct {
	Id     bson.ObjectId `bson:"_id"`
	Name   string
	Prefix string

	Dirent fuse.Dirent
	Fattr  fuse.Attr
}

func (g *GridFsFile) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Printf("GridFsFile.Attr() for: %+v", g)

	db, s := getDb()
	defer s.Close()

	file, err := db.GridFS(g.Prefix).OpenId(g.Id)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	a.Mode = 0400
	a.Uid = uint32(os.Getuid())
	a.Gid = uint32(os.Getgid())
	a.Size = uint64(file.Size())
	a.Ctime = file.UploadDate()
	a.Atime = time.Now()
	a.Mtime = file.UploadDate()
	return nil
}

func (g *GridFsFile) Lookup(ctx context.Context, fname string) (fs.Node, error) {
	log.Printf("GridFsFile.Lookup(): %s\n", fname)

	return nil, fuse.ENOENT
}

func (g *GridFsFile) ReadAll(ctx context.Context) ([]byte, error) {
	log.Printf("GridFsFile.ReadAll(): %s\n", g.Id)

	db, s := getDb()
	defer s.Close()

	file, err := db.GridFS(g.Prefix).OpenId(g.Id)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// TODO: return a pointer to the buffer? memory usage?
	var buf []byte
	buf, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
