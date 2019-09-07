package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jlym/webservice-benchmarks/sqlite"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func monitor(conf *config, db *sqlite.DataStore) error {
	testStarted := false
	startTime := time.Now()
	startupThreshold := startTime.Add(conf.startupWait)

	shuttingDown := false
	var shuttingDownStart time.Time
	for {
		time.Sleep(conf.pollingInterval)

		processInfos, err := getProcessInfo(conf)
		if err != nil {
			return err
		}

		if !testStarted {
			testStarted = len(processInfos) > 0
		}
		if !testStarted {
			if time.Now().After(startupThreshold) {
				return nil
			}
			continue
		}

		if len(processInfos) == 0 {
			now := time.Now()
			if shuttingDown {
				shutdownThreshold := shuttingDownStart.Add(conf.shutdownWait)
				if now.After(shutdownThreshold) {
					return nil
				}
			} else {
				shuttingDown = true
				shuttingDownStart = now
			}
			continue
		}

		err = monitorProcesses(conf, processInfos, db)
		if err != nil {
			return err
		}

	}

	return nil
}

func monitorProcesses(conf *config, processInfos []*processInfo, db *sqlite.DataStore) error {
	for _, processInfo := range processInfos {
		err := monitorProcess(conf, processInfo, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func monitorProcess(conf *config, processInfo *processInfo, db *sqlite.DataStore) error {
	now := time.Now().UTC()

	pid := processInfo.id
	connStats, err := net.ConnectionsPid("all", pid)
	if err != nil {
		return errors.Wrap(err, "fetching connection info failed")
	}

	tcpConnParams := &sqlite.AddTCPConnParams{
		RunID: conf.runID,
		Time:  now,
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
	db.QueueTCPConn(tcpConnParams)

	for _, connStat := range connStats {
		db.QueueConnStatus(&sqlite.AddConnStatusParams{
			Time:        now,
			RunID:       conf.runID,
			Fd:          connStat.Fd,
			LocalIP:     connStat.Laddr.IP,
			LocalPort:   connStat.Laddr.Port,
			RemoteIP:    connStat.Raddr.IP,
			RemotePort:  connStat.Raddr.Port,
			Status:      connStat.Status,
			ProcessID:   uint32(pid),
			ProcessName: processInfo.name,
		})
	}

	return nil
}

type processInfo struct {
	id   int32
	name string
}

func getProcessInfo(conf *config) ([]*processInfo, error) {
	processNameSet := make(map[string]bool)
	for _, processName := range conf.targetProcessNames {
		processNameSet[strings.ToLower(processName)] = true
	}

	processes, err := process.Processes()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	results := make([]*processInfo, 0, len(conf.targetProcessNames))
	for _, process := range processes {
		name, err := process.Name()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		name = strings.ToLower(name)

		includeProcess := conf.includeAllProcesses || processNameSet[name]
		if !includeProcess {
			continue
		}

		results = append(results, &processInfo{
			id:   process.Pid,
			name: name,
		})
	}

	return results, nil
}
