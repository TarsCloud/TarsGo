package tars

import (
	"github.com/TarsCloud/TarsGo/tars/protocol/res/propertyf"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ReportMethod interface {
	Desc() string
	Set(int)
	Get() string
	clear()
}

type Sum struct {
	data  int
	mlock *sync.Mutex
}

func NewSum() *Sum {
	return &Sum{
		data:  0,
		mlock: new(sync.Mutex)}

}

func (s *Sum) Desc() string {
	return "Sum"
}
func (s *Sum) Get() (out string) {
	s.mlock.Lock()
	defer s.mlock.Unlock()
	out = strconv.Itoa(s.data)
	s.clear()
	return
}

func (s *Sum) clear() {
	s.data = 0
}

func (s *Sum) Set(in int) {
	s.mlock.Lock()
	defer s.mlock.Unlock()
	s.data += in
}

type Avg struct {
	count int
	sum   int
	mlock *sync.Mutex
}

func NewAvg() *Avg {
	return &Avg{
		count: 0,
		sum:   0,
		mlock: new(sync.Mutex),
	}
}
func (a *Avg) Desc() string {
	return "Avg"
}
func (a *Avg) Get() (out string) {
	a.mlock.Lock()
	defer a.mlock.Unlock()
	if a.count == 0 {
		out = "0"
		return
	}
	out = strconv.FormatFloat(float64(a.sum)/float64(a.count), 'f', -1, 64)
	a.clear()
	return
}
func (a *Avg) Set(in int) {
	a.mlock.Lock()
	defer a.mlock.Unlock()
	a.count++
	a.sum += in
}

func (a *Avg) clear() {
	a.count = 0
	a.sum = 0
}

type Max struct {
	data  int
	mlock *sync.Mutex
}

func NewMax() *Max {
	return &Max{
		data:  -9999999,
		mlock: new(sync.Mutex)}
}

func (m *Max) Set(in int) {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	if in > m.data {
		m.data = in
	}
}

func (m *Max) Get() (out string) {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	out = strconv.Itoa(m.data)
	m.clear()
	return
}
func (m *Max) clear() {
	m.data = -9999999
}

func (m *Max) Desc() string {
	return "Max"
}

type Min struct {
	data  int
	mlock *sync.Mutex
}

func NewMin() *Min {
	return &Min{
		data:  0,
		mlock: new(sync.Mutex),
	}

}
func (m *Min) Desc() string {
	return "Min"
}

func (m *Min) Set(in int) {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	if m.data == 0 || (m.data > in && in != 0) {
		m.data = in
	}
}
func (m *Min) Get() (out string) {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	out = strconv.Itoa(m.data)
	m.clear()
	return

}
func (m *Min) clear() {
	m.data = 0
}

type Distr struct {
	dataRange []int
	result    []int
	mlock     *sync.Mutex
}

func (d *Distr) Desc() string {
	return "Distr"
}

func NewDistr(in []int) (d *Distr) {
	d = new(Distr)
	d.mlock = new(sync.Mutex)
	s := tools.UniqueInts(in)
	sort.Ints(s)
	d.dataRange = s
	d.result = make([]int, len(d.dataRange))
	return d
}

func (d *Distr) Set(in int) {
	d.mlock.Lock()
	defer d.mlock.Unlock()
	index := tools.UpperBound(d.dataRange, in)
	d.result[index]++
}

func (d *Distr) Get() string {
	d.mlock.Lock()
	defer d.mlock.Unlock()
	var s string = ""
	for i := range d.dataRange {
		if i != 0 {
			s += ","
		}
		s = s + strconv.Itoa(d.dataRange[i]) + "|" + strconv.Itoa(d.result[i])
	}
	d.clear()
	return s
}

func (d *Distr) clear() {
	for i := range d.result {
		d.result[i] = 0
	}
}

type Count struct {
	mlock *sync.Mutex
	data  int
}

func NewCount() *Count {
	return &Count{
		data:  0,
		mlock: new(sync.Mutex),
	}
}

func (c *Count) Set(in int) {
	c.mlock.Lock()
	defer c.mlock.Unlock()
	c.data++
}

func (c *Count) Get() (out string) {
	c.mlock.Lock()
	defer c.mlock.Unlock()
	out = strconv.Itoa(c.data)
	c.clear()
	return
}

func (c *Count) clear() {
	c.data = 0
}
func (c *Count) Desc() string {
	return "Count"
}

type PropertyReportHelper struct {
	reportPtrs []*PropertyReport
	mlock      *sync.Mutex
	comm       *Communicator
	pf         *propertyf.PropertyF
	node       string
}

var ProHelper *PropertyReportHelper

func (p *PropertyReportHelper) ReportToServer() {
	p.mlock.Lock()
	defer p.mlock.Unlock()
	cfg := GetServerConfig()
	var head propertyf.StatPropMsgHead
	statMsg := make(map[propertyf.StatPropMsgHead]propertyf.StatPropMsgBody)
	head.IPropertyVer = 2
	if cfg != nil {
		if cfg.Enableset {
			setList := strings.Split(cfg.Setdivision, ".")
			head.ModuleName = cfg.App + "." + cfg.Server + "." + setList[0] + setList[1] + setList[2]
			head.SetName = setList[0]
			head.SetArea = setList[1]
			head.SetID = setList[2]

		} else {
			head.ModuleName = cfg.App + "." + cfg.Server

		}

	} else {
		return
	}
	head.Ip = cfg.LocalIP
	for _, v := range p.reportPtrs {
		head.PropertyName = v.key
		var body propertyf.StatPropMsgBody
		body.VInfo = make([]propertyf.StatPropInfo, 0)
		for _, m := range v.reportMethods {
			var info propertyf.StatPropInfo
			bflag := false
			desc := m.Desc()
			result := m.Get()
			//todo: use interface method IsDefault() bool
			switch desc {
			case "Sum":
				if result != "0" {
					bflag = true

				}
			case "Avg":
				if result != "0" {
					bflag = true

				}
			case "Distr":
				if result != "" {
					bflag = true

				}
			case "Max":
				if result != "-9999999" {
					bflag = true

				}
			case "Min":
				if result != "0" {
					bflag = true
				}
			case "Count":
				if result != "0" {
					bflag = true

				}
			default:
				bflag = true
			}

			if !bflag {
				continue
			}
			info.Policy = desc
			info.Value = result
			body.VInfo = append(body.VInfo, info)

		}
		statMsg[head] = body
		_, err := p.pf.ReportPropMsg(statMsg)
		if err != nil {
			TLOG.Error("Send to property server Error", reflect.TypeOf(err), err)
		}
	}

}

func (p *PropertyReportHelper) Init(comm *Communicator, node string) {
	p.node = node
	p.mlock = new(sync.Mutex)
	p.comm = comm
	p.pf = new(propertyf.PropertyF)
	p.reportPtrs = make([]*PropertyReport, 0)
	p.comm.StringToProxy(p.node, p.pf)
}

func initProReport() {
	comm := NewCommunicator()
	comm.SetProperty("netthread", 1)
	ProHelper = new(PropertyReportHelper)
	ProHelper.Init(comm, GetClientConfig().property)
	go ProHelper.Run()

}

func (p *PropertyReportHelper) AddToReport(pr *PropertyReport) {
	p.reportPtrs = append(p.reportPtrs, pr)
}

func (p *PropertyReportHelper) Run() {
	//todo , get report interval from config
	loop := time.NewTicker(PropertyReportInterval)
	for range loop.C {
		p.ReportToServer()
	}

}

type PropertyReport struct {
	key           string
	reportMethods []ReportMethod
}

func CreatePropertyReport(key string, argvs ...ReportMethod) (ptr *PropertyReport) {
	for _, v := range ProHelper.reportPtrs {
		if v.key == key {
			ptr = v
			return
		}
	}
	ptr = new(PropertyReport)
	ptr.reportMethods = make([]ReportMethod, 0)
	for _, v := range argvs {
		ptr.reportMethods = append(ptr.reportMethods, v)
	}
	ptr.key = key
	ProHelper.reportPtrs = append(ProHelper.reportPtrs, ptr)
	return
}

func (p *PropertyReport) Report(in int) {
	for _, v := range p.reportMethods {
		v.Set(in)
	}
}
