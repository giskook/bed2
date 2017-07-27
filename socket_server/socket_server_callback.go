package socket_server

import (
	"github.com/gansidui/gotcp"
	"github.com/giskook/bed2/socket_server/protocol"
	"log"
	"runtime/debug"
)

func (ss *SocketServer) OnConnect(c *gotcp.Conn) bool {
	connection := NewConnection(c, &ConnConf{
		read_limit:  ss.conf.ReadLimit,
		write_limit: ss.conf.WriteLimit,
	})

	c.PutExtraData(connection)
	go connection.Check()
	log.Printf("<CNT> %x \n", c.GetRawConn())

	return true
}

func (ss *SocketServer) OnClose(c *gotcp.Conn) {
	connection := c.GetExtraData().(*Connection)
	ss.cm.Del(connection.ID)
	connection.Close()
	log.Printf("<DIS> %x\n", c.GetRawConn())
	debug.PrintStack()
}

func (ss *SocketServer) OnMessage(c *gotcp.Conn, p gotcp.Packet) bool {
	connection := c.GetExtraData().(*Connection)
	connection.UpdateReadFlag()
	connection.RecvBuffer.Write(p.Serialize())
	for {
		protocol_id, length := protocol.CheckProtocol(connection.RecvBuffer)
		buf := make([]byte, length)
		connection.RecvBuffer.Read(buf)
		switch protocol_id {
		case protocol.PROTOCOL_HALF_PACK:
			return true
		case protocol.PROTOCOL_ILLEGAL:
			return true
		case protocol.PROTOCOL_REQ_LOGIN:
			ss.eh_login(buf, connection)
		case protocol.PROTOCOL_REQ_HEART:
			ss.eh_heart(buf, connection)
		case protocol.PROTOCOL_REP_CONTROL:
			ss.eh_control(buf)
		case protocol.PROTOCOL_REP_ACTIVE_TEST:
			ss.eh_active_test(buf)
		}
	}
}
