package subsystem

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string // 时间片权重
	CpuSet      string // cpu核心数
}

type Subsystem interface {
	// Name 返回subsystem的名字，比如cpu memory
	Name() string
	// Set 设置某个 cgroup 在这个Subsystem中的资源限制
	Set(cgroupPath string, res *ResourceConfig) error
	// Apply 将进程添加到某个cgroup中
	Apply(cgroupPath string, pid int) error
	// Remove 移除某个group
	Remove(cgroupPath string) error
}

var (
	SubsystemIns = []Subsystem{
		&CpusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
