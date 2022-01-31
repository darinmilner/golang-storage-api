package handlers

import (
	"fileuploader/pkg/logger"
	"net/rpc"
	"os"
)

func rpcClient(inMaintenanceMode bool) {
	rpcPort := os.Getenv("RPC_PORT")
	c, err := rpc.Dial("tcp", "127.0.0.1:"+rpcPort)
	if err != nil {
		logger.Errorf("error starting rpc server ", err)
		return
	}

	logger.Info("Connected...")
	var result string
	err = c.Call("RPCServer.MaintenanceMode", inMaintenanceMode, &result)
	if err != nil {
		logger.Errorf("error calling maintenance mode ", err)
		return
	}

	logger.Info(result)
}
