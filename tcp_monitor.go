package webservice_benchmarks

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jlym/webservice-benchmarks/sqlite"
	"github.com/jlym/webservice-benchmarks/util"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func monitorConns(db *sqlite.DataStore, stopReciever *util.StopReciever, run *sqlite.Run) error {
	defer stopReciever.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	pid := int32(os.Getpid())
	for stopReciever.ShouldContinue() {
		pidToName := make(map[int32]string)

		connStats, err := net.ConnectionsPid("all", pid)
		if err != nil {
			return errors.Wrap(err, "fetching connection info failed")
		}

		now := time.Now().UTC()
		tcpConnParams := &sqlite.AddTCPConnParams{
			Time: now,
		}

		for _, connStat := range connStats {
			switch connStat.Status {
			case "ESTABLISHED":
				tcpConnParams.Established++
			case "SYN_SENT":
				tcpConnParams.SynSent++
			case "SYN_RECV":
				tcpConnParams.SynRecv++
			case "FIN_WAIT1":
				tcpConnParams.FinWait1++
			case "FIN_WAIT2":
				tcpConnParams.FinWait2++
			case "TIME_WAIT":
				tcpConnParams.TimeWait++
			case "CLOSE":
				tcpConnParams.Close++
			case "CLOSE_WAIT":
				tcpConnParams.CloseWait++
			case "LAST_ACK":
				tcpConnParams.LastAck++
			case "LISTEN":
				tcpConnParams.Listen++
			case "CLOSING":
				tcpConnParams.Closing++
			default:
				log.Println(fmt.Sprintf("unrecognized status: %s", connStat.Status))
			}
		}
		db.QueueTCPConn(run, tcpConnParams)

		for _, connStat := range connStats {
			pid := connStat.Pid
			processName, stored := pidToName[pid]
			if !stored {
				p, err := process.NewProcess(pid)
				if err != nil {
					return errors.Wrapf(err, "fetching process for pid %d failed", pid)
				}
				processName, err = p.Name()
				if err != nil {
					return errors.Wrapf(err, "fetching process name for pid %d failed", pid)
				}
				pidToName[pid] = processName
			}

			connStatusParams := &sqlite.AddConnStatusParams{
				Time:        now,
				Fd:          connStat.Fd,
				LocalIP:     connStat.Laddr.IP,
				LocalPort:   connStat.Laddr.Port,
				RemoteIP:    connStat.Raddr.IP,
				RemotePort:  connStat.Raddr.Port,
				Status:      connStat.Status,
				ProcessID:   uint32(pid),
				ProcessName: processName,
			}
			db.QueueConnStatus(run, connStatusParams)
		}

		<-ticker.C
	}
	return nil
}
