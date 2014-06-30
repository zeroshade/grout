package grout

import (
	"sort"
)

type TaskManager interface {
	Execute()
	AddTask(t Task) bool
	SuspendTask(t Task)
	RemoveTask(t Task)
	ResumeTask(t Task)
	KillAllTasks()

	GetSettings() *Config
}

type taskList []Task

func (slice taskList) Len() int {
	return len(slice)
}

func (slice taskList) Less(i, j int) bool {
	return slice[i].GetPriority() < slice[j].GetPriority()
}

func (slice taskList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice taskList) Find(t Task) int {
	i := sort.Search(len(slice), func(i int) bool { return slice[i].GetPriority() >= t.GetPriority() })
	if i < len(slice) && slice[i].GetPriority() == t.GetPriority() {
		for k := i; k < len(slice); k++ {
			if slice[k] == t {
				return k
			}
		}
	}
	return -1
}

func (slice *taskList) Remove(i int) {
	s := *slice
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil
	s = s[:len(s)-1]
	*slice = s
}

type taskMgr struct {
	taskList       taskList
	pausedTaskList taskList
	conf           Config
}

func newTaskMgr() *taskMgr {
	t := &taskMgr{taskList: make([]Task, 0), pausedTaskList: make([]Task, 0)}
	if err := loadSettings(&t.conf); err != nil {
		panic("Failed to load settings")
	}
	return t
}

var (
	_This         TaskManager    = newTaskMgr()
	vidUpdate     *videoTask     = &videoTask{NewBasicTask(1000), nil, nil, _This.GetSettings().Video.W, _This.GetSettings().Video.H}
	stateUpdate   *gameStateTask = &gameStateTask{BasicTask: NewBasicTask(500)}
	fpsUpdate     *fpsTask       = &fpsTask{BasicTask: NewBasicTask(1), p: _This.GetSettings().Debug.PrintFPS}
	timerUpdate   *timerTask     = &timerTask{BasicTask: NewBasicTask(2)}
	interUpdate   *listTask      = interpolatorUpdater(3)
	triggerUpdate *listTask      = triggerUpdater(4)
)

func RegisterTrigger(t Trigger) {
	triggerUpdate.list = append(triggerUpdate.list, t)
}

func RegisterInterpolator(i Interpolator) {
	interUpdate.list = append(interUpdate.list, i)
}

func init() {
	_This.AddTask(vidUpdate)
	_This.AddTask(stateUpdate)
	_This.AddTask(fpsUpdate)
	_This.AddTask(timerUpdate)
	_This.AddTask(interUpdate)
}

func InitialGameState(g GameState) {
	stateUpdate.stk[0] = g
	stateUpdate.stk[0].Init(vidUpdate.win)
}

func GetTaskManager() TaskManager {
	return _This
}

func (tm *taskMgr) GetSettings() *Config {
	return &tm.conf
}

func (tm *taskMgr) AddTask(t Task) bool {
	if !t.Start() {
		return false
	}

	tm.taskList = append(tm.taskList, t)
	sort.Sort(tm.taskList)

	return true
}

func (tm *taskMgr) SuspendTask(t Task) {
	if i := tm.taskList.Find(t); i != -1 {
		t.OnSuspend()
		tm.taskList.Remove(i)
		tm.pausedTaskList = append(tm.pausedTaskList, t)
		sort.Sort(tm.pausedTaskList)
	}
}

func (tm *taskMgr) ResumeTask(t Task) {
	if i := tm.pausedTaskList.Find(t); i != -1 {
		t.OnResume()
		tm.pausedTaskList.Remove(i)
		tm.taskList = append(tm.taskList, t)
		sort.Sort(tm.taskList)
	}
}

func (tm *taskMgr) RemoveTask(t Task) {
	if i := tm.taskList.Find(t); i != -1 {
		t.SetCanKill(true)
	}
}

func (tm *taskMgr) KillAllTasks() {
	for _, t := range tm.taskList {
		t.SetCanKill(true)
	}
}

func (tm *taskMgr) Execute() {
	for len(tm.taskList) > 0 {
		for _, t := range tm.taskList {
			if !t.CanKill() {
				t.Update()
			}
		}
		for i := 0; i < len(tm.taskList); i++ {
			if tm.taskList[i].CanKill() {
				tm.taskList[i].Stop()
				tm.taskList.Remove(i)
				i--
			}
		}
	}
}
