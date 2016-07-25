package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GridFs implements both Node and Handle for a collection.
type GridFs struct {
	Name string // the GridFS prefix

	Dirent fuse.Dirent
	Fattr  fuse.Attr
}

func (g *GridFs) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Printf("GridFs.Attr() for: %+v", g)
	a.Mode = os.ModeDir | 0700
	a.Uid = uint32(os.Getuid())
	a.Gid = uint32(os.Getgid())
	return nil
}

func (g *GridFs) Lookup(ctx context.Context, fname string) (fs.Node, error) {
	log.Printf("GridFs[%s].Lookup(): %s\n", g.Name, fname)

	extIdx := strings.LastIndex(fname, ".")
	if extIdx > 0 {
		fname = fname[0:extIdx]
	}

	if !bson.IsObjectIdHex(fname) {
		log.Printf("Invalid ObjectId: %s\n", fname)
		return nil, fuse.ENOENT
	}

	db, s := getDb()
	defer s.Close()

	id := bson.ObjectIdHex(fname)
	gf := GridFsFile{Id: id}
	file, err := db.GridFS(g.Name).OpenId(id)
	if err != nil {
		log.Printf("Error while looking up %s: %s \n", id, err.Error())
		return nil, fuse.EIO
	}
	defer file.Close()

	gf.Name = file.Name()
	gf.Prefix = g.Name

	return &gf, nil
}

func (g *GridFs) ReadDirAll(ctx context.Context) (ents []fuse.Dirent, ferr error) {
	log.Println("GridFs.ReadDirAll(): ", g.Name)

	db, s := getDb()
	defer s.Close()

	gfs := db.GridFS(g.Name)
	iter := gfs.Find(nil).Iter()

	var f *mgo.GridFile
	for gfs.OpenNext(iter, &f) {
		name := f.Id().(bson.ObjectId).Hex() + filepath.Ext(f.Name())
		ents = append(ents, fuse.Dirent{Name: name, Type: fuse.DT_File})
	}

	if err := iter.Close(); err != nil {
		log.Printf("Could not list GridFs files: %s \n", err.Error())
		return nil, fuse.EIO
	}

	return ents, nil
}

func (g *GridFs) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	log.Printf("GridFs.Remove(): %s/%s \n", g.Name, req.Name)

	id := req.Name
	extIdx := strings.LastIndex(id, ".")
	if extIdx > 0 {
		id = id[0:extIdx]
	}

	if !bson.IsObjectIdHex(id) {
		return fuse.ENOENT
	}

	db, s := getDb()
	defer s.Close()

	if err := db.GridFS(g.Name).RemoveId(bson.ObjectIdHex(id)); err != nil {
		log.Printf("Could not remove GridFs file '%s': %s \n", id, err.Error())
		return fuse.EIO
	}

	return nil
}
