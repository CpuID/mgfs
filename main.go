package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
)

var appName string = "mgfs"
var gridfsPrefix string

func handleSignal(sc chan os.Signal, mount_point string) {
	for {
		<-sc
		fmt.Printf("\n")
		log.Printf("Exit signal received, cleaning up.\n")
		unmount(mount_point)
		os.Exit(0)
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(appName + ": ")

	app := cli.NewApp()
	app.Name = appName
	app.Usage = "A FUSE filesystem which uses MongoDB GridFS as a storage backend"
	app.Version = mgfs_version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr, a",
			Value: "localhost",
			Usage: "MongoDB host or IP to connect to",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 27017,
			Usage: "MongoDB port to connect to",
		},
		cli.StringFlag{
			Name:  "user, u",
			Value: "",
			Usage: "Username to access MongoDB (Optional)",
		},
		cli.StringFlag{
			Name:  "password, P",
			Value: "",
			Usage: "Password to access MongoDB (Optional)",
		},
		// TODO: do we want this? is this the bucket name? or a different prefix?
		cli.StringFlag{
			Name:  "gridfs, g",
			Value: "fs",
			Usage: "GridFS prefix",
		},
		cli.StringFlag{
			Name:  "bucket, b",
			Value: "fs",
			Usage: "Bucket (Database) Name in MongoDB",
		},
		cli.StringFlag{
			Name:  "mount, m",
			Value: "/mnt/mgfs",
			Usage: "Mount point on local OS",
		},
	}

	app.Action = func(c *cli.Context) {
		//dbName := c.Args()[0]
		dbName := c.String("bucket")
		//mountPoint := c.Args()[1]
		mountPoint := c.String("mount")
		dbHost := c.String("addr")
		dbPort := string(c.String("port"))
		dbUser := c.String("user")
		dbPassword := c.String("password")
		gridfsPrefix = c.String("gridfs")
		credentials := dbUser + ":" + dbPassword

		// Connect to the database
		initDb(dbName, dbHost, dbPort, credentials)

		// Handle signals correctly.
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT)
		signal.Notify(sc, syscall.SIGKILL)
		signal.Notify(sc, syscall.SIGTERM)
		go handleSignal(sc, mountPoint)

		// Mount the database
		err := mount(mountPoint, appName)
		if err != nil {
			log.Fatal(err)
		}
	}

	app.Run(os.Args)
}
