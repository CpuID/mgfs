package main

import (
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func mount(mount_point string, fs_name string) error {
	// startup mount
	mount, err := fuse.Mount(
		mount_point,
		fuse.FSName(fs_name),
		fuse.Subtype(fs_name),
		fuse.VolumeName(fs_name),
		fuse.LocalVolume(),
	)
	defer mount.Close()
	if err != nil {
		return err
	}

	// Adding this incase we do kernel cache invalidations later...
	if p := mount.Protocol(); !p.HasInvalidate() {
		log.Fatalf("Kernel FUSE support is too old to have invalidations: version %v\n", p)
	}

	log.Printf("Mounted: %s\n", mount_point)
	if err = fs.Serve(mount, &GridFS{}); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-mount.Ready
	if err := mount.MountError; err != nil {
		return err
	}
	return nil
}

func unmount(mount_point string) {
	log.Printf("Unmounting: %s\n", mount_point)
	err := fuse.Unmount(mount_point)
	if err != nil {
		log.Fatal(err)
	}
}
