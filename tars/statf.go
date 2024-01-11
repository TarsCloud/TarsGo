package tars

import (
	"fmt"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/statf"
	"github.com/TarsCloud/TarsGo/tars/util/sync"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

var (
	// ReportStat set the default stater(default is `ReportStatFromClient`).
	ReportStat reportStatFunc = ReportStatFromClient
	timePoint                 = [...]int32{5, 10, 50, 100, 200, 500, 1000, 2000, 3000}
)

type reportStatFunc func(msg *Message, succ int32, timeout int32, exec int32)

// StatInfo struct contains stat info head and body.
type StatInfo struct {
	Head statf.StatMicMsgHead
	Body statf.StatMicMsgBody
}

// StatFHelper is helper struct for stat reporting.
type StatFHelper struct {
	chStatInfo           chan StatInfo
	mStatInfo            map[statf.StatMicMsgHead]statf.StatMicMsgBody
	mStatCount           map[statf.StatMicMsgHead]int
	app                  *application
	comm                 *Communicator
	sf                   *statf.StatF
	servant              string
	chStatInfoFromServer chan StatInfo
	mStatInfoFromServer  map[statf.StatMicMsgHead]statf.StatMicMsgBody
	mStatCountFromServer map[statf.StatMicMsgHead]int
}

func newStatFHelper(app *application) *StatFHelper {
	return &StatFHelper{
		chStatInfo:           make(chan StatInfo, app.ServerConfig().StatReportChannelBufLen),
		mStatInfo:            make(map[statf.StatMicMsgHead]statf.StatMicMsgBody),
		mStatCount:           make(map[statf.StatMicMsgHead]int),
		app:                  app,
		chStatInfoFromServer: make(chan StatInfo, app.ServerConfig().StatReportChannelBufLen),
		mStatInfoFromServer:  make(map[statf.StatMicMsgHead]statf.StatMicMsgBody),
		mStatCountFromServer: make(map[statf.StatMicMsgHead]int),
	}
}

// Init the StatFHelper
func (s *StatFHelper) Init(comm *Communicator, servant string) {
	s.servant = servant
	s.comm = comm
	s.sf = new(statf.StatF)
	s.comm.StringToProxy(s.servant, s.sf)
}

func (s *StatFHelper) collectMsg(statInfo StatInfo, mStatInfo map[statf.StatMicMsgHead]statf.StatMicMsgBody, mStatCount map[statf.StatMicMsgHead]int) {
	if body, ok := mStatInfo[statInfo.Head]; ok {
		body.Count += statInfo.Body.Count
		body.TimeoutCount += statInfo.Body.TimeoutCount
		body.ExecCount += statInfo.Body.ExecCount
		body.TotalRspTime += statInfo.Body.TotalRspTime
		if body.MaxRspTime < statInfo.Body.MaxRspTime {
			body.MaxRspTime = statInfo.Body.MaxRspTime
		}
		if body.MinRspTime > statInfo.Body.MinRspTime {
			body.MinRspTime = statInfo.Body.MinRspTime
		}
		s.getIntervCount(int32(statInfo.Body.TotalRspTime), body.IntervalCount)
		mStatInfo[statInfo.Head] = body
		mStatCount[statInfo.Head]++
	} else {
		firstBody := statf.StatMicMsgBody{}
		firstBody.Count = statInfo.Body.Count
		firstBody.TimeoutCount = statInfo.Body.TimeoutCount
		firstBody.ExecCount = statInfo.Body.ExecCount
		firstBody.TotalRspTime = statInfo.Body.TotalRspTime
		firstBody.MaxRspTime = statInfo.Body.MaxRspTime
		firstBody.MinRspTime = statInfo.Body.MinRspTime
		firstBody.IntervalCount = map[int32]int32{}
		s.getIntervCount(int32(statInfo.Body.TotalRspTime), firstBody.IntervalCount)
		mStatInfo[statInfo.Head] = firstBody
		mStatCount[statInfo.Head] = 1
	}
}

func (s *StatFHelper) getIntervCount(totalRspTime int32, intervalCount map[int32]int32) {
	var (
		bNeedInit, bGetIntev bool
	)
	// The first time you need to initialize all the plot points to 0
	if len(intervalCount) == 0 {
		bNeedInit = true
	}
	for _, point := range timePoint {
		if !bGetIntev && totalRspTime < point {
			bGetIntev = true
			intervalCount[point]++
			if !bNeedInit {
				break
			} else {
				continue
			}
		}
		if bNeedInit {
			intervalCount[point] = 0
		}
	}
}

func (s *StatFHelper) reportAndClear(mStat string, bFromClient bool) {
	// report mStatInfo
	if mStat == "mStatInfo" {
		_, err := s.sf.ReportMicMsg(s.mStatInfo, bFromClient, s.comm.Client.Context())
		if err != nil {
			TLOG.Debug("mStatInfo report err:", err.Error())
		}
		s.mStatInfo = make(map[statf.StatMicMsgHead]statf.StatMicMsgBody)
		s.mStatCount = make(map[statf.StatMicMsgHead]int)
	}
	// report mStatInfoFromServer
	if mStat == "mStatInfoFromServer" {
		_, err := s.sf.ReportMicMsg(s.mStatInfoFromServer, bFromClient, s.comm.Client.Context())
		if err != nil {
			TLOG.Debug("mStatInfoFromServer report err:", err.Error())
		}
		s.mStatInfoFromServer = make(map[statf.StatMicMsgHead]statf.StatMicMsgBody)
		s.mStatCountFromServer = make(map[statf.StatMicMsgHead]int)
	}
}

// Run stat report loop
func (s *StatFHelper) Run() {
	ticker := time.NewTicker(s.app.ServerConfig().StatReportInterval)
	for {
		select {
		case stStatInfo := <-s.chStatInfo:
			s.collectMsg(stStatInfo, s.mStatInfo, s.mStatCount)
		case stStatInfoFromServer := <-s.chStatInfoFromServer:
			s.collectMsg(stStatInfoFromServer, s.mStatInfoFromServer, s.mStatCountFromServer)
		case <-ticker.C:
			if len(s.mStatInfo) > 0 {
				s.reportAndClear("mStatInfo", true)
			}
			if len(s.mStatInfoFromServer) > 0 {
				s.reportAndClear("mStatInfoFromServer", false)
			}
		}
	}
}

func (s *StatFHelper) pushBackMsg(stStatInfo StatInfo, fromServer bool) {
	if fromServer {
		s.chStatInfoFromServer <- stStatInfo
	} else {
		s.chStatInfo <- stStatInfo
	}
}

// ReportMicMsg report the StatInfo ,from server shows whether it comes from server.
func (s *StatFHelper) ReportMicMsg(stStatInfo StatInfo, fromServer bool) {
	s.pushBackMsg(stStatInfo, fromServer)
}

// StatReport instance pointer of StatFHelper
var (
	StatReport   *StatFHelper
	statInited   = make(chan struct{}, 1)
	statInitOnce sync.Once
)

func initReport(app *application) error {
	cfg := app.ClientConfig()
	if err := cfg.ValidateStat(); err != nil {
		statInited <- struct{}{}
		return err
	}
	comm := app.Communicator()
	StatReport = newStatFHelper(app)
	StatReport.Init(comm, cfg.Stat)
	statInited <- struct{}{}
	go StatReport.Run()
	return nil
}

// ReportStatBase is base method for report statistics.
func ReportStatBase(head *statf.StatMicMsgHead, body *statf.StatMicMsgBody, FromServer bool) {
	if StatReport == nil {
		return
	}
	statInfo := StatInfo{Head: *head, Body: *body}
	if statInfo.Head.TarsVersion == "" {
		statInfo.Head.TarsVersion = Version
	}
	// statInfo.Head.IStatVer = 2
	StatReport.ReportMicMsg(statInfo, FromServer)
}

// ReportStatFromClient report the statics from client.
func ReportStatFromClient(msg *Message, succ int32, timeout int32, exec int32) {
	cCfg := GetClientConfig()
	var head statf.StatMicMsgHead
	var body statf.StatMicMsgBody
	head.MasterName = cCfg.ModuleName
	head.MasterIp = tools.GetLocalIP()
	if sCfg := GetServerConfig(); sCfg != nil && sCfg.Enableset {
		head.MasterIp = sCfg.LocalIP
		setList := strings.Split(sCfg.Setdivision, ".")
		head.MasterName = fmt.Sprintf("%s.%s.%s%s%s@%s", sCfg.App, sCfg.Server, setList[0], setList[1], setList[2], sCfg.Version)
	}

	head.InterfaceName = msg.Req.SFuncName
	sNames := strings.Split(msg.Req.SServantName, ".")
	if len(sNames) < 2 {
		TLOG.Debugf("report err:servant name (%s) format error", msg.Req.SServantName)
		return
	}
	head.SlaveName = fmt.Sprintf("%s.%s", sNames[0], sNames[1])
	if msg.Adp != nil {
		head.SlaveIp = msg.Adp.GetPoint().Host
		head.SlavePort = msg.Adp.GetPoint().Port
		if msg.Adp.GetPoint().SetId != "" {
			setList := strings.Split(msg.Adp.GetPoint().SetId, ".")
			head.SlaveSetName = setList[0]
			head.SlaveSetArea = setList[1]
			head.SlaveSetID = setList[2]
			head.SlaveName = fmt.Sprintf("%s.%s.%s%s%s", sNames[0], sNames[1], setList[0], setList[1], setList[2])
		}
	}
	if msg.Resp != nil {
		head.ReturnValue = msg.Resp.IRet
	} else {
		head.ReturnValue = -1
	}

	body.Count = succ
	body.TimeoutCount = timeout
	body.ExecCount = exec
	body.TotalRspTime = msg.Cost()
	body.MaxRspTime = int32(body.TotalRspTime)
	body.MinRspTime = int32(body.TotalRspTime)
	ReportStatBase(&head, &body, false)
}

// ReportStatFromServer reports statics from server side.
func ReportStatFromServer(InterfaceName, MasterName string, ReturnValue int32, TotalRspTime int64) {
	cfg := GetServerConfig()
	var head statf.StatMicMsgHead
	var body statf.StatMicMsgBody
	head.SlaveName = fmt.Sprintf("%s.%s", cfg.App, cfg.Server)
	head.SlaveIp = cfg.LocalIP
	if cfg.Enableset {
		setList := strings.Split(cfg.Setdivision, ".")
		head.SlaveName = fmt.Sprintf("%s.%s.%s%s%s", cfg.App, cfg.Server, setList[0], setList[1], setList[2])
		head.SlaveSetName = setList[0]
		head.SlaveSetArea = setList[1]
		head.SlaveSetID = setList[2]
	}
	head.InterfaceName = InterfaceName
	head.MasterName = MasterName
	head.ReturnValue = ReturnValue

	if ReturnValue == 0 {
		body.Count = 1
	} else {
		body.ExecCount = 1
	}
	body.TotalRspTime = TotalRspTime
	body.MaxRspTime = int32(body.TotalRspTime)
	body.MinRspTime = int32(body.TotalRspTime)
	ReportStatBase(&head, &body, true)
}
