# mgfs

*mgfs is a [FUSE](https://bazil.org/fuse/) filesystem which uses [MongoDB GridFS](https://docs.mongodb.com/manual/core/gridfs/) as a storage backend.*

[![Build Status](https://travis-ci.org/CpuID/mgfs.svg?branch=master)](https://travis-ci.org/CpuID/mgfs) [![Coverage Status](https://coveralls.io/repos/github/CpuID/mgfs/badge.svg?branch=master)](https://coveralls.io/github/CpuID/mgfs?branch=master)

# Installation
You need to have [Golang](https://golang.org/doc/install) installed.
Open your terminal, and run `go get github.com/amsa/mgfs`. Now you should be able to run `mgfs` (be sure to add $GOPATH/bin to your $PATH).

# How to use
First mount your MongoDb database: `mgfs test /path/to/mount/dir`. You may now go to the directory specified 
as the mount point, and see the collections (directories), and documents (json files). You may read, update, 
or delete the documents. You may also read and delete GridFs files under the specified prefix (`fs` by default).

Don't forget to unmount the database when you are done (`umount /path/to/mount/dir`).

# Caveats

There is no caching layer implemented in-process or externally (eg. Redis or Memcached). As long as MongoDB is close to your FUSE process latency wise,
you should have no issues. PRs are welcome to implement caching if there is interest :)

# Todo
- [x] Support GridFS read 
- [x] Support GridFS remove 
- [ ] Support GridFS write
- [ ] Show GridFS file names

# Credits

* [bazil.org/fuse](http://bazil.org/fuse)
* [labix.org/mgo](http://labix.org/mgo)
* [mountMgo](https://github.com/cryptix/mountMgo)
* [amsa/mgfs](https://github.com/amsa/mgfs)
