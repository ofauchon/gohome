package intg

import (
	"fmt"

	"github.com/markdaws/gohome"
	"github.com/markdaws/gohome/cmd"
	"github.com/markdaws/gohome/comm"
	"github.com/markdaws/gohome/fluxwifi"
)

type fluxwifiCmdBuilder struct {
	System *gohome.System
}

func (b *fluxwifiCmdBuilder) Build(c cmd.Command) (*cmd.Func, error) {
	switch command := c.(type) {
	case *cmd.ZoneTurnOn:
		z := b.System.Zones[command.ZoneID]
		d := b.System.Devices[z.DeviceID]
		return &cmd.Func{
			Func: func() error {
				pool := d.Connections()
				if pool == nil {
					return fmt.Errorf("fluxwifiCmdBuilder - connection pool not ready")
				}

				conn := pool.Get()
				if conn == nil {
					return fmt.Errorf("fluxwifiCmdBuilder - error connecting, no available connections")
				}

				defer func() {
					pool.Release(conn)
				}()
				return fluxwifi.TurnOn(conn)
			},
		}, nil

	case *cmd.ZoneTurnOff:
		z := b.System.Zones[command.ZoneID]
		d := b.System.Devices[z.DeviceID]
		return &cmd.Func{
			Func: func() error {
				pool := d.Connections()
				if pool == nil {
					return fmt.Errorf("fluxwifiCmdBuilder - connection pool not ready")
				}

				conn := pool.Get()
				if conn == nil {
					return fmt.Errorf("fluxwifiCmdBuilder - error connecting, no available connections")
				}

				defer func() {
					pool.Release(conn)
				}()
				return fluxwifi.TurnOff(conn)
			},
		}, nil

	case *cmd.ZoneSetLevel:
		z := b.System.Zones[command.ZoneID]
		d := b.System.Devices[z.DeviceID]
		return &cmd.Func{
			Func: func() error {
				var rV, gV, bV byte
				lvl := command.Level.Value
				if lvl == 0 {
					if (command.Level.R == 0) && (command.Level.G == 0) && (command.Level.B == 0) {
						rV = 0
						gV = 0
						bV = 0
					} else {
						rV = command.Level.R
						gV = command.Level.G
						bV = command.Level.B
					}
				} else {
					rV = byte((lvl / 100) * 255)
					gV = rV
					bV = rV
				}

				pool := d.Connections()
				if pool == nil {
					return fmt.Errorf("fluxwifiCmdBuilder - connection pool not ready")
				}

				conn := pool.Get()
				if conn == nil {
					return fmt.Errorf("fluxwifiCmdBuilder - error connecting, no available connections")
				}

				defer func() {
					pool.Release(conn)
				}()
				return fluxwifi.SetLevel(rV, gV, bV, conn)
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported command type")
	}
}

func (b *fluxwifiCmdBuilder) Connections(name, address string) comm.ConnectionPool {
	createConnection := func() comm.Connection {
		// Add he port number which is 5577 for Flux WIFI bulbs
		conn := comm.NewTelnetConnection(address, nil)

		//TODO: Need to get some ping mechanism for flux bulbs
		/*
			conn.SetPingCallback(func() error {
				if _, err := conn.Write([]byte("#PING\r\n")); err != nil {
					return fmt.Errorf("%s ping failed: %s", d, err)
				}
				return nil
			})*/
		return conn
	}
	return comm.NewConnectionPool(name, 2, createConnection)
}

func (b *fluxwifiCmdBuilder) ID() string {
	return "fluxwifi"
}