package tars

import (
	"fmt"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/statf"
)

var (
	// ReportStat set the default stater(default is `ReportStatFromClient`).
	ReportStat reportStatFunc = ReportStatFromClient
)

type reportStatFunc func(msg *Message, succ int32, timeout int32, exec int32)

// StatInfo struct contains stat info' head and body.
type StatInfo struct {
	Head statf.StatMicMsgHead
	Body statf.StatMicMsgBody
}

// StatFHelper is helper struct for stat reporting.
type StatFHelper struct {
	chStatInfo           chan StatInfo
	mStatInfo            map[statf.StatMicMsgHead]statf.StatMicMsgBody
	mStatCount           map[statf.StatMicMsgHead]int
	comm                 *Communicator
	sf                   *statf.StatF
	node                 string
	chStatInfoFromServer chan StatInfo
	mStatInfoFromServer  map[statf.StatMicMsgHead]statf.StatMicMsgBody
	mStatCountFromServer map[statf.StatMicMsgHead]int
}

// Init init the StatFHelper
func (s *StatFHelper) Init(comm *Communicator, node string) {
	s.node = node
	s.chStatInfo = make(chan StatInfo, GetServerConfig().StatReportChannelBufLen)
	s.chStatInfoFromServer = make(chan StatInfo, GetServerConfig().StatReportChannelBufLen)
	s.mStatInfo = make(map[statf.StatMicMsgHead]statf.StatMicMsgBody)
	s.mStatCount = make(map[statf.StatMicMsgHead]int)
	s.mStatInfoFromServer = make(map[statf.StatMicMsgHead]statf.StatMicMsgBody)
	s.mStatCountFromServer = make(map[statf.StatMicMsgHead]int)
	s.comm = comm
	s.sf = new(statf.StatF)
	s.comm.StringToProxy(s.node, s.sf)
}

func (s *StatFHelper) collectMsg(statInfo StatInfo, mStatInfo map[statf.StatMicMsgHead]statf.StatMicMsgBody, mStatCount map[statf.StatMicMsgHead]int) {
	if body, ok := s.mStatInfo[statInfo.Head]; ok {
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
		s.mStatInfo[statInfo.Head] = body
		s.mStatCount[statInfo.Head]++
	} else {
		firstBody := statf.StatMicMsgBody{}
		firstBody.Count = statInfo.Body.Count
		firstBody.TimeoutCount = statInfo.Body.TimeoutCount
		firstBody.ExecCount = statInfo.Body.ExecCount
		firstBody.TotalRspTime = statInfo.Body.TotalRspTime
		firstBody.MaxRspTime = statInfo.Body.MaxRspTime
		firstBody.MinRspTime = statInfo.Body.MinRspTime
		s.mStatInfo[statInfo.Head] = firstBody
		s.mStatCount[statInfo.Head] = 1
	}
}

func (s *StatFHelper) reportAndClear(mStat string, bFromClient bool) {
	// report mStatInfo
	if mStat == "mStatInfo" {
		_, err := s.sf.ReportMicMsg(s.mStatInfo, bFromClient)
		if err != nil {
			TLOG.Debug("mStatInfo report err:", err.Error())
		}
		s.mStatInfo = make(map[statf.StatMicMsgHead]statf.StatMicMsgBody)
		s.mStatCount = make(map[statf.StatMicMsgHead]int)
	}
	// report mStatInfoFromServer
	if mStat == "mStatInfoFromServer" {
		_, err := s.sf.ReportMicMsg(s.mStatInfoFromServer, bFromClient)
		if err != nil {
			TLOG.Debug("mStatInfoFromServer report err:", err.Error())
		}
		s.mStatInfoFromServer = make(map[statf.StatMicMsgHead]statf.StatMicMsgBody)
		s.mStatCountFromServer = make(map[statf.StatMicMsgHead]int)
	}
}

// Run run stat report loop
func (s *StatFHelper) Run() {
	ticker := time.NewTicker(GetServerConfig().StatReportInterval)
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

// ReportMicMsg report the Statinfo ,from server shows whether it comes from server.
func (s *StatFHelper) ReportMicMsg(stStatInfo StatInfo, fromServer bool) {
	s.pushBackMsg(stStatInfo, fromServer)
}

// StatReport instance pointer of StatFHelper
var StatReport *StatFHelper
var statInited = make(chan struct{}, 1)

func initReport() {
	if GetClientConfig() == nil || GetClientConfig().Stat == "" {
		statInited <- struct{}{}
		return
	}
	comm := NewCommunicator()
	StatReport = new(StatFHelper)
	StatReport.Init(comm, GetClientConfig().Stat)
	statInited <- struct{}{}
	go StatReport.Run()
}

// ReportStatBase is base method for report statitics.
func ReportStatBase(head *statf.StatMicMsgHead, body *statf.StatMicMsgBody, FromServer bool) {
	cfg := GetServerConfig()
	statInfo := StatInfo{Head: *head, Body: *body}
	statInfo.Head.TarsVersion = cfg.Version
	//statInfo.Head.IStatVer = 2
	if StatReport != nil {
		StatReport.ReportMicMsg(statInfo, FromServer)
	}
}

// ReportStatFromClient report the statics from client.
func ReportStatFromClient(msg *Message, succ int32, timeout int32, exec int32) {
	cfg := GetServerConfig()
	if cfg == nil {
		return
	}
	var head statf.StatMicMsgHead
	var body statf.StatMicMsgBody
	head.MasterName = fmt.Sprintf("%s.%s", cfg.App, cfg.Server)
	head.MasterIp = cfg.LocalIP
	//head.SMasterContainer = cfg.Container
	if cfg.Enableset {
		setList := strings.Split(cfg.Setdivision, ".")
		head.MasterName = fmt.Sprintf("%s.%s.%s%s%s@%s", cfg.App, cfg.Server, setList[0], setList[1], setList[2], cfg.Version)
		//head.SMasterSetInfo = cfg.Setdivision
	}

	head.InterfaceName = msg.Req.SFuncName
	sNames := strings.Split(msg.Req.SServantName, ".")
	if len(sNames) < 2 {
		TLOG.Debugf("report err:servant name (%s) format error", msg.Req.SServantName)
		return
	}
	head.SlaveName = fmt.Sprintf("%s.%s", sNames[0], sNames[1])
	if msg.Adp != nil {
		//head.SSlaveContainer = msg.Adp.GetPoint().ContainerName
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
	//head.SSlaveContainer = cfg.Container
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
