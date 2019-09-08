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
	startupThreshold := time.Now().Add(conf.startupWait)
	log.Println(fmt.Sprintf("monitor will wait till %v for processes to start", startupThreshold))

	var shuttingDown bool
	var shutdownThreshold time.Time

	for {
		time.Sleep(conf.pollingInterval)

		processInfos, err := getProcessInfo(conf)
		if err != nil {
			return err
		}

		if !testStarted {
			if len(processInfos) > 0 {
				log.Println("processes found, test started")
				testStarted = true
			}
		}
		if !testStarted {
			if time.Now().After(startupThreshold) {
				log.Println("monitor will exit because processes were not found")
				return nil
			}
			continue
		}

		if len(processInfos) == 0 {
			now := time.Now()
			if shuttingDown {
				if now.After(shutdownThreshold) {
					return nil
				}
			} else {
				shuttingDown = true
				shutdownThreshold = now.Add(conf.shutdownWait)
				log.Println(fmt.Sprintf("processes not found, starting shutdown. shutting down after %v at %v", conf.shutdownWait, shutdownThreshold))
			}
			continue
		}

		shuttingDown = false

		err = monitorProcesses(conf, processInfos, db)
		if err != nil {
			return err
		}
	}
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
		return nil, errors.Wrap(err, "fetching process info failed")
	}

	results := make([]*processInfo, 0, len(conf.targetProcessNames))
	for _, process := range processes {
		name, err := process.Name()
		if err != nil {
			return nil, errors.Wrap(err, "fetching process name failed")
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
