package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/urfave/cli.v1"
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

func verifyRequiredFlags(c *cli.Context) error {
	if len(c.String("addr")) == 0 {
		return errors.New("Error: addr (a) flag is required.")
	}
	if len(c.String("mount")) == 0 {
		return errors.New("Error: mount (m) flag is required.")
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "A FUSE filesystem which uses MongoDB GridFS as a storage backend"
	app.Version = mgfs_version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr, a",
			Value: "",
			Usage: "MongoDB host or IP to connect to (Required)",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 27017,
			Usage: "MongoDB port to connect to (Optional)",
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
		cli.StringFlag{
			Name:  "bucket, b",
			Value: "fs",
			Usage: "Bucket (Database) Name in MongoDB (Optional)",
		},
		cli.StringFlag{
			Name:  "mount, m",
			Value: "",
			Usage: "Mount point on local OS (Required)",
		},
	}

	app.Action = func(c *cli.Context) error {
		//dbName := c.Args()[0]
		dbName := c.String("bucket")
		//mountPoint := c.Args()[1]
		mountPoint := c.String("mount")
		dbHost := c.String("addr")
		dbPort := string(c.String("port"))
		dbUser := c.String("user")
		dbPassword := c.String("password")
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
		return nil
	}

	app.Before = func(c *cli.Context) error {
		err := verifyRequiredFlags(c)
		if err != nil {
			log.Printf("%s\n\n", err.Error())
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		return nil
	}

	app.Run(os.Args)
}
