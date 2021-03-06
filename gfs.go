// Peter Enns' implementation of project 1 in CMSC818E
// memfs implements a simple in-memory file system.
package main

/*
 Two main files are ../fuse.go and ../fs/serve.go
*/

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"dss/myfs"
	"dss/util"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT DATABSE\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage

	debugPtr := flag.Bool("debug", false, "print lots of stuff")
	newfsPtr := flag.Bool("newfs", false, "start with an empty file system")
	mtimePtr := flag.Bool("mtimeArchives", false, "use modify timestamp instead of version timestamp for archives")
	name := flag.String("name", "auto", "replica name")
	configFile := flag.String("config", "config.txt", "path to config file")
	flag.Parse()
	util.SetDebug(*debugPtr)
	myfs.UseMtime = *mtimePtr

	util.P_out("main\n")

	pid := myfs.GetOurPid(*configFile, *name)
	replicas := myfs.ReadReplicaInfo(*configFile)
	thisReplica := replicas[pid]

	if thisReplica == nil {
		util.P_err("No applicable replica")
		os.Exit(1)
	}

	if *newfsPtr {
		os.RemoveAll(thisReplica.DbPath)
	}

	if _, err := os.Stat(thisReplica.MntPoint); os.IsNotExist(err) {
		os.MkdirAll(thisReplica.MntPoint, os.ModeDir)
	}

	db, err := myfs.NewLeveldbFsDatabase(thisReplica.DbPath)
	//db := &myfs.DummyFsDb{}
	//err := error(nil)
	if err != nil {
		util.P_err("Problem loading the database: ", err)
		os.Exit(-1)
	}
	filesystem := myfs.NewFs(db, thisReplica, replicas)
	go filesystem.PeriodicFlush()

	//mountpoint := flag.Arg(0)
	mountpoint := thisReplica.MntPoint

	fuse.Unmount(mountpoint) //!!
	c, err := fuse.Mount(mountpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	err = fs.Serve(c, filesystem)
	if err != nil {
		log.Fatal(err)
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}

// 818E - YOU DON'T NEED THESE
//func (n *Node) Getattr(req *fuse.GetattrRequest, resp *fuse.GetattrResponse, intr fs.Intr) fuse.Error {}
// func (n *Node) Setattr(req *fuse.SetattrRequest, resp *fuse.SetattrResponse, intr fs.Intr) fuse.Error {}
// func (n fs.Node) Open(req *fuse.OpenRequest, resp *fuse.OpenResponse, intr fs.Intr) (fs.Handle, fuse.Error){}
// func (n fs.Node) Release(req *fuse.ReleaseRequest, intr Intr) fuse.Error {}
// func (n fs.Node) Removexattr(req *fuse.RemovexattrRequest, intr Intr) fuse.Error {}
// func (n fs.Node) Setxattr(req *fuse.SetxattrRequest, intr Intr) fuse.Error {}
// func (n fs.Node) Listxattr(req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse, intr Intr) fuse.Error {}
// func (n fs.Node) Getxattr(req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse, intr Intr) fuse.Error {}
