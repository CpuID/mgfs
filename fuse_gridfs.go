package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GridFS implements the MongoDB GridFS filesystem
type GridFS struct{}

//Name string // the GridFS prefix
// TODO: verify if these are used?
//Dirent fuse.Dirent
//Fattr  fuse.Attr

func (g *GridFS) Root() (fs.Node, error) {
	log.Println("GridFS.Root() : Returning root node.")
	return &Dir{
		Inode: 1,
		Name:  "/",
		Mode:  0755,
		// Leave ModTime undefined for the root dir for now.
	}, nil
}

////////////////////////////////////////////////////////

type Dir struct {
	Inode   uint64
	Name    string
	Mode    os.FileMode // 0755
	ModTime time.Time
	Uid     uint32
	Gid     uint32
}

var _ = fs.Node(&Dir{})

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Printf("Dir[%s].Attr()\n", d.Name)
	a.Inode = d.Inode
	a.Mode = d.Mode
	a.Mtime = d.ModTime
	// Use the current user/group at all times, no references in MongoDB.
	a.Uid = uint32(os.Getuid())
	a.Gid = uint32(os.Getgid())
	return nil
}

/*
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Printf("Dir.Lookup(): %s\n", name)

	// TODO: need to perform mongodb lookup against fs.files to build
	// a list of files in the root dir, plus the directories that exist in the root.

	// Check if lookup is on the GridFS
	if name == gridfsPrefix {
		return &GridFs{Name: gridfsPrefix}, nil
	}

	db, s := getDb()
	defer s.Close()

	names, err := db.CollectionNames()
	if err != nil {
		log.Panic(err)
		return nil, fuse.EIO
	}

	for _, collName := range names {
		if collName == name {
			return &CollFile{Name: name}, nil
		}
	}

	return nil, fuse.ENOENT
}
*/

// TODO: do we need to be defining DT_Dir based entries in the alternate ReadDirAll() below?
/*
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("Dir.ReadDirAll():", d.name)

	db, s := getDb()
	defer s.Close()

	names, err := db.CollectionNames()
	if err != nil {
		log.Panic(err)
		return nil, fuse.EIO
	}

	ents := make([]fuse.Dirent, 0, len(names)+1) // one more for GridFS

	// Append GridFS prefix
	ents = append(ents, fuse.Dirent{Name: gridfsPrefix, Type: fuse.DT_Dir})

	// Append the rest of the collections
	for _, name := range names {
		if strings.HasSuffix(name, ".indexes") {
			continue
		}
		ents = append(ents, fuse.Dirent{Name: name, Type: fuse.DT_Dir})
	}
	return ents, nil
}
*/

////////////////////////////////////////////////////////

func (d *Dir) Lookup(ctx context.Context, path string) (fs.Node, error) {
	log.Printf("Dir[%s].Lookup(): %s\n", d.Name, path)

	// TODO: distinguish between root dir and a path? or is it already done for us?
	// TODO: change from object ID listings to filename listings in a directory hierarchy...

	extIdx := strings.LastIndex(path, ".")
	if extIdx > 0 {
		path = path[0:extIdx]
	}
	fmt.Printf("new path: %s\n", path)

	if !bson.IsObjectIdHex(path) {
		log.Printf("Invalid ObjectId: %s\n", path)
		return nil, fuse.ENOENT
	}

	db, s := getDb()
	defer s.Close()

	id := bson.ObjectIdHex(path)
	gf := GridFsFile{Id: id}
	file, err := db.GridFS(d.Name).OpenId(id)
	if err != nil {
		log.Printf("Error while looking up %s: %s \n", id, err.Error())
		return nil, fuse.EIO
	}
	defer file.Close()

	gf.Name = file.Name()
	gf.Prefix = d.Name

	return &gf, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) (ents []fuse.Dirent, ferr error) {
	log.Printf("Dir[%s].ReadDirAll()", d.Name)

	// TODO: change from object ID listings to filename listings in a directory hierarchy...

	db, s := getDb()
	defer s.Close()

	gfs := db.GridFS(d.Name)
	iter := gfs.Find(nil).Iter()

	var f *mgo.GridFile
	for gfs.OpenNext(iter, &f) {
		name := f.Id().(bson.ObjectId).Hex() + filepath.Ext(f.Name())
		ents = append(ents, fuse.Dirent{Name: name, Type: fuse.DT_File})
	}

	if err := iter.Close(); err != nil {
		log.Printf("Could not list GridFS files: %s \n", err.Error())
		return nil, fuse.EIO
	}

	return ents, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	log.Printf("Dir[%s].Remove(): %s\n", d.Name, req.Name)

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

	if err := db.GridFS(d.Name).RemoveId(bson.ObjectIdHex(id)); err != nil {
		log.Printf("Could not remove GridFS file '%s': %s \n", id, err.Error())
		return fuse.EIO
	}

	return nil
}

////////////////////////////////////////////////////////

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

func (g *GridFsFile) Lookup(ctx context.Context, path string) (fs.Node, error) {
	log.Printf("GridFsFile.Lookup(): %s\n", path)

	return nil, fuse.ENOENT
}

// TODO: do chunked reads instead using Read(), far nicer than ReadAll()
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
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
