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
)

func monitorConns(db *sqlite.DataStore, stopReciever *util.StopReciever, run *sqlite.Run) error {
	defer stopReciever.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	pid := int32(os.Getpid())
	for stopReciever.ShouldContinue() {
		connStats, err := net.ConnectionsPid("all", pid)
		if err != nil {
			return errors.Wrap(err, "fetching connection info failed")
		}

		tcpConnParams := &sqlite.AddTCPConnParams{
			Time: time.Now().UTC(),
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

		<-ticker.C
	}
	return nil
}
