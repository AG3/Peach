package main

import (
	"../Logger"
	"../Structs"
	"encoding/json"
	"net"
	"os"
	"strings"
)

var serverList Structs.ServerList
var serverManager *net.TCPListener

func getServerConfig() {
	serverConfig, err := os.Open("../Config/servers.conf")
	defer serverConfig.Close()
	checkError(err)

	buf := make([]byte, 1024)
	length, err := serverConfig.Read(buf)
	checkError(err)

	err = json.Unmarshal(buf[:length], &serverList)
	checkError(err)

	Logger.Info("Get server config complete")
	return
}

func setLogger() {
	Logger.SetConsole(true)
	Logger.SetRollingDaily("../logs", "Manager-logs.txt")
}

func manageServer(conn net.Conn) {
	var serverName, portToListen string
	var id int
	defer conn.Close()
	defer Logger.Debug("DISCONNECT " + conn.RemoteAddr().String())

	for {
		buffer := make([]byte, 512)
		length, err := conn.Read(buffer)

		if err != nil {
			return
		}
		//checkError(err)

		cmd := strings.Split(string(buffer[:length]), "|")

		if cmd[0] == "ONLINE" {
			if cmd[1] == "GATE_SERVER" {
				id, serverName, portToListen = findFreeServer(Structs.GATE_SERVER)
				if id != -1 {
					serverList.Gate[id].Conn = conn
				}
			}
			if cmd[1] == "CONNECTOR_SERVER" {
				id, serverName, portToListen = findFreeServer(Structs.CONNECTOR_SERVER)
				if id != -1 {
					serverList.Connector[id].Conn = conn
				}
			}
			if cmd[1] == "CHANNEL_SERVER" {
				id, serverName, portToListen = findFreeServer(Structs.CHANNEL_SERVER)
				if id != -1 {
					serverList.Channel[id].Conn = conn
				}
			}
			if cmd[1] == "LOGIC_SERVER" {
				id, serverName, portToListen = findFreeServer(Structs.LOGIC_SERVER)
				if id != -1 {
					serverList.Logic[id].Conn = conn
				}
			}
			conn.Write([]byte(serverName + "|" + portToListen))
		}
		Logger.Debug(conn)
	}
}

func setupManager() {
	listenPort, err := net.ResolveTCPAddr("tcp", serverList.Manager[0].Port)
	checkError(err)

	serverManager, err = net.ListenTCP("tcp", listenPort)
	checkError(err)

	Logger.Info("Server manager setup success")
	for {
		conn, err := serverManager.Accept()
		checkError(err)
		go manageServer(conn)
	}
}

func findFreeServer(serverType int) (int, string, string) {
	switch serverType {
	case Structs.GATE_SERVER:
		for i := 0; i < len(serverList.Gate[:]); i++ {
			if serverList.Gate[i].Name != "" && serverList.Gate[i].IsAvailable == false {
				serverList.Gate[i].IsAvailable = true
				return i, serverList.Gate[i].Name, serverList.Gate[i].Port
			}
		}
	case Structs.CONNECTOR_SERVER:
		for i := 0; i < len(serverList.Connector[:]); i++ {
			if serverList.Connector[i].Name != "" && serverList.Connector[i].IsAvailable == false {
				serverList.Connector[i].IsAvailable = true
				return i, serverList.Connector[i].Name, serverList.Connector[i].Port
			}
		}
	case Structs.CHANNEL_SERVER:
		for i := 0; i < len(serverList.Channel[:]); i++ {
			if serverList.Channel[i].Name != "" && serverList.Channel[i].IsAvailable == false {
				serverList.Channel[i].IsAvailable = true
				return i, serverList.Channel[i].Name, serverList.Channel[i].Port
			}
		}
	case Structs.LOGIC_SERVER:
		for i := 0; i < len(serverList.Logic[:]); i++ {
			if serverList.Logic[i].Name != "" && serverList.Logic[i].IsAvailable == false {
				serverList.Logic[i].IsAvailable = true
				return i, serverList.Logic[i].Name, serverList.Logic[i].Port
			}
		}
	}
	return -1, "ERROR", "ERROR"
}

func main() {
	setLogger()
	Logger.Info("Starting Manager Server...")
	getServerConfig()
	setupManager()
}

func checkError(err error) {
	if err != nil {
		Logger.Error(err.Error())
	}
}
