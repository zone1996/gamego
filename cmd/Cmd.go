package cmd

import (
	"gamego/netya"

	proto "github.com/golang/protobuf/proto"
	log "github.com/zone1996/logo"
)

type Cmd interface { // one code, one Cmd
	Exec(session *netya.IoSession, msg *netya.PbMsg)
}

var Cmds map[int32]Cmd

func registerCmd(code int32, c Cmd) {
	if _, ok := Cmds[code]; ok {
		panic("Cmd already exist")
	}
	Cmds[code] = c
}

func InitCmd() {
	Cmds = make(map[int32]Cmd)
	// Register all your cmds here.
	registerCmd(1, &PlayerLoginCmd{})
}

func GetCmd(code int32) (cmd Cmd, ok bool) {
	cmd, ok = Cmds[code]
	return
}

// Just an example
type PlayerLoginCmd struct{}

func (this *PlayerLoginCmd) Exec(session *netya.IoSession, msg *netya.PbMsg) {
	userId := msg.UserId
	log.Info("Receive code=? from SessionId=?, UserId=?", msg.Code, session.Id, userId)

	msg1 := &netya.PbMsg{}
	msg1.Code = 1
	msg1.UserId = 999
	msg1.Payload = []byte("send back payload data")
	msg1.Length = int32(msg.XXX_Size())
	mdata, err := proto.Marshal(msg1)

	if err == nil {
		_, err = session.Write(mdata)
		if err != nil {
			log.Info("SendBack err:?", err)
		}
	} else {
		log.Info("Marshal Err:?", err.Error())
	}
}
