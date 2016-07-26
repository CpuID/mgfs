package main

import (
	"log"
	"os"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"golang.org/x/net/context"
)

func mount(point string, fsname string) error {
	// startup mount
	c, err := fuse.Mount(
		point,
		fuse.FSName(fsname),
		fuse.VolumeName(fsname),
		fuse.LocalVolume(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	log.Println("Mounted: ", point)
	if err = fs.Serve(c, mgoFS{}); err != nil {
		log.Fatal(err)
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func unmount(mount_point string) {
	log.Printf("Unmounting %s\n", mount_point)
	err := fuse.Unmount(mount_point)
	if err != nil {
		log.Fatal(err)
	}
}

///////////////////////////////////////////

// mgoFS implements my mgo fuse filesystem
type mgoFS struct{}

func (mgoFS) Root() (fs.Node, error) {
	log.Println("returning root node")
	return &Dir{"Root"}, nil
}

// Dir implements both Node and Handle for the root directory.
type Dir struct {
	name string
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("Dir.Attr() for ", d.name)
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Dir.Lookup():", name)

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
