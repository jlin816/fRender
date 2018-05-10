package client

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	. "common"
)

//const masterAddress = "hello_world"
const blenderPath = "/Applications/Blender/blender.app/Contents/MacOS/blender"

type RenderFramesArgs struct {
	StartFrame int
	EndFrame   int
	Filename   string
}

type Friend struct {
	me            int
    username      string
    port          int
    masterAddr    net.Addr
	requesterConn net.Conn
	server        net.Listener
	available	  bool
	httpClient	  *net.Client
}

func initFriend(username string, port int, masterAddr string) *Friend {
    addr, err := net.ResolveTCPAddr("tcp", masterAddr)
    if err != nil{
        fmt.Printf("Invalid master addr %s", masterAddr)
        panic(err)
    }
    friend := Friend{username: username, port: port, masterAddr: addr}
	friend.registerWithMaster()


    // Friends receive RPCs as well, init as a server
	rpc.Register(&friend)

    // Hacky stuff from https://github.com/golang/go/issues/13395
    oldMux := http.DefaultServeMux
    mux := http.NewServeMux()
    http.DefaultServeMux = mux

	rpc.HandleHTTP()

    http.DefaultServeMux = oldMux

    // Init TCP socket for file transfer
	server, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port)) // TODO: update with actual port
	if err != nil {
		os.Exit(1)
	}

	friend.server = server
	go http.Serve(server, nil)

	go friend.sendHeartbeatsToMaster()
	go friend.listenServer()

	return &friend
}

func (fr *Friend) listenServer() {
	for {
		conn, err := fr.server.Accept()
		if err != nil {
			os.Exit(1)
		}
		fr.requesterConn = conn
		fr.receiveFile(conn)
		// do something
	}
}

func fillString(returnString string, toLength int) string {
	// from http://www.mrwaggel.be/post/golang-transfer-a-file-over-a-tcp-socket/
	for {
		lengthString := len(returnString)
		if lengthString < toLength {
			returnString = returnString + ":"
			continue
		}
		break
	}
	return returnString
}

func (fr *Friend) sendFile(connection net.Conn, filename string) {
	// from http://www.mrwaggel.be/post/golang-transfer-a-file-over-a-tcp-socket/
	defer connection.Close()
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Sending filename and filesize
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	return
}

func (fr *Friend) receiveFile(connection net.Conn) { // maybe want port as argument
	// connection, err := net.Dial("tcp", "localhost:27001") // TODO: Update port
	// if err != nil {
	// 	panic(err)
	// }
	// defer connection.Close()
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)
	fmt.Printf("file received\n")

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")
	newFile, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, connection, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}
}

func (fr *Friend) registerWithMaster() {
    httpClient, err := rpc.DialHTTP("tcp", fr.masterAddr.String())
	if err != nil {
        fmt.Println("Couldn't connect friend to master")
		panic(err)
	}

	myIP, err := externalIP()
	if err != nil {
		fmt.Println("Couldn't get IP addr for friend")
		panic(err)
	}
    args := RegisterFriendArgs{Address: fmt.Sprintf("%s:%d", myIP, fr.port), Username: fr.username}
    reply := RegisterFriendReply{}
    err = httpClient.Call("Master.RegisterFriend", args, &reply)
    if err != nil {
        fmt.Printf("Error registering friend: %v", err)
        panic(err)
    }

	fr.httpClient = httpClient

	fmt.Printf("friend registered w/master\n")
}

func (fr *Friend) receiveJob() {

}

func (fr *Friend) renderFrames(file string, start_frame int, end_frame int) {
	// blender -b bob_lamp_update_export.blend -s 0 -e 100 -o render_files/frame_##### -a

	output_folder, _ := filepath.Abs(fmt.Sprintf("%v_frames/frame_#####", file))
	absoluteFilepath, _ := filepath.Abs(file)

	args := []string{
		"-b",
		absoluteFilepath,
		"-F",
		"PNG",
		"-s",
		fmt.Sprint(start_frame),
		"-e",
		fmt.Sprint(end_frame),
		"-o",
		output_folder,
		"-a",
	}

	blenderCmd := exec.Command(blenderPath, args...)
	_, err := blenderCmd.Output()
	if err != nil {
		panic(err)
	}

}

func (fr *Friend) sendHeartbeatsToMaster() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for _ = range ticker.C {
		args := HeartbeatArgs{Available: fr.available}
		reply := HeartbeatReply{}
		err := fr.httpClient.Call("Master.Heartbeat", args, &reply)
		if err != nil {
			fmt.Printf("Error sending heartbeats to master: %v", err)
			panic(err)
		}
	}
}

func (fr *Friend) PrintHello(args int, reply *int) error {
	fmt.Printf("HIYA\n")
	*reply = 100
	return nil
}

func (fr *Friend) RenderFrames(args RenderFramesArgs, reply *int) error {
	fmt.Printf("rendering frames\n")
	fr.renderFrames(args.Filename, args.StartFrame, args.EndFrame)
	return nil
}

// From https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func externalIP() (string, error) {
    ifaces, err := net.Interfaces()
    if err != nil {
        return "", err
    }
    for _, iface := range ifaces {
        if iface.Flags&net.FlagUp == 0 {
            continue // interface down
        }
        if iface.Flags&net.FlagLoopback != 0 {
            continue // loopback interface
        }
        addrs, err := iface.Addrs()
        if err != nil {
            return "", err
        }
        for _, addr := range addrs {
            var ip net.IP
            switch v := addr.(type) {
            case *net.IPNet:
                ip = v.IP
            case *net.IPAddr:
                ip = v.IP
            }
            if ip == nil || ip.IsLoopback() {
                continue
            }
            ip = ip.To4()
            if ip == nil {
                continue // not an ipv4 address
            }
            return ip.String(), nil
        }
    }
    return "", errors.New("are you connected to the network?")
}
