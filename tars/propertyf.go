package tars

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/propertyf"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

// ReportPolicy is report policy
type ReportPolicy int

const (
	ReportPolicyUnknown ReportPolicy = iota
	ReportPolicySum                  // 1
	ReportPolicyAvg                  // 2
	ReportPolicyDistr                // 3
	ReportPolicyMax                  // 4
	ReportPolicyMin                  // 5
	ReportPolicyCount                // 6
)

func (p ReportPolicy) String() string {
	switch p {
	case ReportPolicySum:
		return "Sum"
	case ReportPolicyAvg:
		return "Avg"
	case ReportPolicyDistr:
		return "Distr"
	case ReportPolicyMax:
		return "Max"
	case ReportPolicyMin:
		return "Min"
	case ReportPolicyCount:
		return "Count"
	default:
		return "Unknown"
	}
}

// ReportMethod is the interface for all kinds of report methods.
type ReportMethod interface {
	Enum() ReportPolicy
	Set(int)
	Get() string
	clear()
}

// Sum report methods.
type Sum struct {
	data  int
	mlock *sync.Mutex
}

// NewSum new and init the sum report methods.
func NewSum() *Sum {
	return &Sum{
		data:  0,
		mlock: new(sync.Mutex)}

}

// Enum return the report policy
func (s *Sum) Enum() ReportPolicy {
	return ReportPolicySum
}

// Get gets the result of the sum report method.
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

// Set sets a value tho the sum method.
func (s *Sum) Set(in int) {
	s.mlock.Lock()
	defer s.mlock.Unlock()
	s.data += in
}

// Avg for counting average for the report value.
type Avg struct {
	count int
	sum   int
	mlock *sync.Mutex
}

// NewAvg new and init the average struct.
func NewAvg() *Avg {
	return &Avg{
		count: 0,
		sum:   0,
		mlock: new(sync.Mutex),
	}
}

// Enum return the report policy
func (a *Avg) Enum() ReportPolicy {
	return ReportPolicyAvg
}

// Get gets the result of the average counting.
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

// Set sets the value for the average counting.
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

// Max struct is for counting the Max value for the reporting value.
type Max struct {
	data  int
	mlock *sync.Mutex
}

// NewMax new and init the Max struct.
func NewMax() *Max {
	return &Max{
		data:  -9999999,
		mlock: new(sync.Mutex)}
}

// Enum return the report policy
func (m *Max) Enum() ReportPolicy {
	return ReportPolicyMax
}

// Set sets a value for counting max.
func (m *Max) Set(in int) {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	if in > m.data {
		m.data = in
	}
}

// Get gets the max value.
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

// Min is the struct for counting the min value.
type Min struct {
	data  int
	mlock *sync.Mutex
}

// NewMin new and init the min struct.
func NewMin() *Min {
	return &Min{
		data:  0,
		mlock: new(sync.Mutex),
	}

}

// Enum return the report policy
func (m *Min) Enum() ReportPolicy {
	return ReportPolicyMin
}

// Set sets a value for counting min value.
func (m *Min) Set(in int) {
	m.mlock.Lock()
	defer m.mlock.Unlock()
	if m.data == 0 || (m.data > in && in != 0) {
		m.data = in
	}
}

// Get get the min value for the Min struct.
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

// Distr is used for counting the distribution of the reporting values.
type Distr struct {
	dataRange []int
	result    []int
	mlock     *sync.Mutex
}

// NewDistr new and int the Distr
func NewDistr(in []int) (d *Distr) {
	d = new(Distr)
	d.mlock = new(sync.Mutex)
	s := tools.UniqueInts(in)
	sort.Ints(s)
	d.dataRange = s
	d.result = make([]int, len(d.dataRange))
	return d
}

// Enum return the report policy
func (d *Distr) Enum() ReportPolicy {
	return ReportPolicyDistr
}

// Set sets the value for counting distribution.
func (d *Distr) Set(in int) {
	d.mlock.Lock()
	defer d.mlock.Unlock()
	index := tools.UpperBound(d.dataRange, in)
	d.result[index]++
}

// Get get the distribution of the reporting values.
func (d *Distr) Get() string {
	d.mlock.Lock()
	defer d.mlock.Unlock()
	var s string
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

// Count is for counting the total of reporting
type Count struct {
	mlock *sync.Mutex
	data  int
}

// NewCount new and init the counting struct.
func NewCount() *Count {
	return &Count{
		data:  0,
		mlock: new(sync.Mutex),
	}
}

// Enum return the report policy
func (c *Count) Enum() ReportPolicy {
	return ReportPolicyCount
}

// Set sets the value for counting.
func (c *Count) Set(in int) {
	c.mlock.Lock()
	defer c.mlock.Unlock()
	c.data++
}

// Get gets the total times of the reporting values.
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

// PropertyReportHelper is helper struct for property report.
type PropertyReportHelper struct {
	reportPtrs *sync.Map //string -> *PropertyReport
	comm       *Communicator
	pf         *propertyf.PropertyF
	node       string
}

// ProHelper is global PropertyReportHelper instance
var ProHelper *PropertyReportHelper
var proOnce sync.Once

// ReportToServer report to the remote propertyreport server.
func (p *PropertyReportHelper) ReportToServer() {
	cfg := GetServerConfig()
	statMsg := make(map[propertyf.StatPropMsgHead]propertyf.StatPropMsgBody)

	var head propertyf.StatPropMsgHead
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
	//head.SContainer = cfg.Container

	p.reportPtrs.Range(func(key, val interface{}) bool {
		v := val.(*PropertyReport)
		head.PropertyName = v.key

		var body propertyf.StatPropMsgBody
		body.VInfo = make([]propertyf.StatPropInfo, 0)
		for _, m := range v.reportMethods {
			if nil == m {
				continue
			}

			var info propertyf.StatPropInfo
			bflag := false
			desc := m.Enum().String()
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

		return true
	})

	var cnt int
	var tmpStatMsg = make(map[propertyf.StatPropMsgHead]propertyf.StatPropMsgBody)
	for k, v := range statMsg {
		cnt++
		if cnt >= 20 {
			_, err := p.pf.ReportPropMsg(tmpStatMsg)
			if err != nil {
				TLOG.Error("Send to property server Error", reflect.TypeOf(err), err)
			}
			tmpStatMsg = make(map[propertyf.StatPropMsgHead]propertyf.StatPropMsgBody)
		}
		tmpStatMsg[k] = v
	}
	if len(tmpStatMsg) > 0 {
		_, err := p.pf.ReportPropMsg(tmpStatMsg)
		if err != nil {
			TLOG.Error("Send to property server Error", reflect.TypeOf(err), err)
		}
	}
}

// Init inits the PropertyReportHelper
func (p *PropertyReportHelper) Init(comm *Communicator, node string) {
	p.node = node
	p.comm = comm
	p.pf = new(propertyf.PropertyF)
	p.reportPtrs = new(sync.Map)
	p.comm.StringToProxy(p.node, p.pf)
}

func initProReport() {
	if GetClientConfig() == nil || GetClientConfig().Property == "" {
		return
	}
	comm := NewCommunicator()
	ProHelper = new(PropertyReportHelper)
	ProHelper.Init(comm, GetClientConfig().Property)
	go ProHelper.Run()
}

// AddToReport adds the user's PropertyReport to the PropertyReportHelper
func (p *PropertyReportHelper) AddToReport(pr *PropertyReport) {
	p.reportPtrs.Store(pr.key, pr)
}

// Run start the properting report goroutine.
func (p *PropertyReportHelper) Run() {
	//todo , get report interval from config
	loop := time.NewTicker(GetServerConfig().PropertyReportInterval)
	for range loop.C {
		p.ReportToServer()
	}
}

// PropertyReport property report struct
type PropertyReport struct {
	key           string
	reportMethods []ReportMethod
}

// Report reports a value.
func (p *PropertyReport) Report(in int) {
	for _, v := range p.reportMethods {
		if v != nil {
			v.Set(in)
		}
	}
}

// CreatePropertyReport creats the property report instance with the key.
func CreatePropertyReport(key string, argvs ...ReportMethod) *PropertyReport {
	ptr := GetPropertyReport(key)
	for _, v := range argvs {
		ptr.reportMethods[v.Enum()] = v
	}

	return ptr
}

// GetPropertyReport gets the property report instance with the key.
func GetPropertyReport(key string) *PropertyReport {
	proOnce.Do(initProReport)
	if val, ok := ProHelper.reportPtrs.Load(key); ok {
		if pr, ok := val.(*PropertyReport); ok {
			return pr
		}
	}

	ptr := new(PropertyReport)
	ptr.key = key
	ptr.reportMethods = make([]ReportMethod, 7)
	ProHelper.AddToReport(ptr)

	return ptr
}

// ReportSum sum report
func ReportSum(key string, i int) {
	ptr := GetPropertyReport(key)
	policy := ReportPolicySum
	if nil == ptr.reportMethods[policy] {
		ptr.reportMethods[policy] = NewSum()
	}
	ptr.reportMethods[policy].Set(i)
}

// ReportAvg avg report
func ReportAvg(key string, i int) {
	ptr := GetPropertyReport(key)
	policy := ReportPolicyAvg
	if nil == ptr.reportMethods[policy] {
		ptr.reportMethods[policy] = NewAvg()
	}
	ptr.reportMethods[policy].Set(i)
}

// ReportMax max report
func ReportMax(key string, i int) {
	ptr := GetPropertyReport(key)
	policy := ReportPolicyMax
	if nil == ptr.reportMethods[policy] {
		ptr.reportMethods[policy] = NewMax()
	}
	ptr.reportMethods[policy].Set(i)
}

// ReportMin min report
func ReportMin(key string, i int) {
	ptr := GetPropertyReport(key)
	policy := ReportPolicyMin
	if nil == ptr.reportMethods[policy] {
		ptr.reportMethods[policy] = NewMin()
	}
	ptr.reportMethods[policy].Set(i)
}

// ReportDistr distr report
func ReportDistr(key string, in []int, i int) {
	ptr := GetPropertyReport(key)
	policy := ReportPolicyDistr
	if nil == ptr.reportMethods[policy] {
		ptr.reportMethods[policy] = NewDistr(in)
	}
	ptr.reportMethods[policy].Set(i)
}

// ReportCount count report
func ReportCount(key string, i int) {
	ptr := GetPropertyReport(key)
	policy := ReportPolicyCount
	if nil == ptr.reportMethods[policy] {
		ptr.reportMethods[policy] = NewCount()
	}
	ptr.reportMethods[policy].Set(i)
}
