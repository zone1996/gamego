package cmd

import (
	"gamego/netya"

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
	registerCmd(1, &ExampleCmd{})
}

func GetCmd(code int32) (cmd Cmd, ok bool) {
	cmd, ok = Cmds[code]
	return
}

// Just an example
type ExampleCmd struct{}

func (this *ExampleCmd) Exec(session *netya.IoSession, msg *netya.PbMsg) {
	userId := msg.GetUserId()
	log.Info("Receive code=? from SessionId=?, UserId=?", msg.GetCode(), session.Id, userId)

	msg1 := netya.NewPbMsg(msg.GetCode())
	msg1.SetUserId(userId)
	msg1.SetPayload([]byte("send back payload data"))

	mdata := msg1.Bytes()
	if mdata == nil {
		log.Info("mdata is nil")
	}

	n, err := session.Write(mdata)
	if err != nil {
		log.Info("WriteBack err:?", err)
	}
	log.Info("===========Write Back=======?, n=?", session.Id, n)
}
