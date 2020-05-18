package spectrumservice

type CollectedStorageMetrics struct {
	Metrics            []*StorageMetrics
	Status             int
	CollectionDuration float64
}

type CollectedSwitchMetrics struct {
	Metrics            []*SwitchMetrics
	Status             int
	CollectionDuration float64
}

type StorageMetrics struct {
	Storage              StorageSystem
	StorageSystemMetrics []MetricValue
	VolumeMap            map[string]string
	VolumeMetrics        []MetricValue
}

type SwitchMetrics struct {
	Switch                  Switch
	SwitchAggregatedMetrics []MetricValue
}

type CollectedPoolMetrics struct {
	Metrics            []*PoolsMetrics
	Status             int
	CollectionDuration float64
}

type PoolsMetrics struct {
	Pool Pool
}

// types generated from IBM Spectrum response

//MetricsDetails  for V1
type MetricDetails struct {
	Metrics map[string]MetricDetail `json:"metricDetails"`
}

// MetricDetails struct for V1
type MetricDetail struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Units       string `json:"units"`
}

// MetricValue struct for V1
type MetricValue struct {
	Current []struct {
		X int64    `json:"x"`
		Y *float64 `json:"y"`
	} `json:"current"`
	DeviceID         int     `json:"deviceId"`
	DeviceName       string  `json:"deviceName"`
	EndTime          int64   `json:"endTime"`
	Label            string  `json:"label"`
	MaxValue         float64 `json:"maxValue"`
	MetricID         int     `json:"metricId"`
	MinValue         float64 `json:"minValue"`
	ParentDeviceName string  `json:"parentDeviceName"`
	Precision        int     `json:"precision"`
	ResourceID       uint64  `json:"resourceID"`
	StartTime        int64   `json:"startTime"`
	Units            string  `json:"units"`
}

// StorageSystem struct for V1
type StorageSystem struct {
	AllocatedSpace            string `json:"Allocated Space"`
	AssignedVolumeSpace       string `json:"Assigned Volume Space"`
	AvailablePoolSpace        string `json:"Available Pool Space"`
	Compressed                string `json:"Compressed"`
	CompressionSavings        string `json:"Compression Savings"`
	CustomTag1                string `json:"Custom Tag 1"`
	CustomTag2                string `json:"Custom Tag 2"`
	CustomTag3                string `json:"Custom Tag 3"`
	DataCollection            string `json:"Data Collection"`
	DeduplicationSavings      string `json:"Deduplication Savings"`
	Disks                     string `json:"Disks"`
	Events                    string `json:"Events"`
	Firmware                  string `json:"Firmware"`
	FlashCopy                 string `json:"FlashCopy"`
	IPAddress                 string `json:"IP Address"`
	Location                  string `json:"Location"`
	ManagedDisks              string `json:"Managed Disks"`
	Model                     string `json:"Model"`
	Name                      string `json:"Name"`
	PhysicalAllocation        string `json:"Physical Allocation"`
	PoolCapacity              string `json:"Pool Capacity"`
	Pools                     string `json:"Pools"`
	Ports                     string `json:"Ports"`
	RawDiskCapacity           string `json:"Raw Disk Capacity"`
	ReadCache                 string `json:"Read Cache"`
	RemoteRelationships       string `json:"Remote Relationships"`
	SerialNumber              string `json:"Serial Number"`
	Shortfall                 string `json:"Shortfall"`
	TimeZone                  string `json:"Time Zone"`
	Topology                  string `json:"Topology"`
	TotalDataReductionSavings string `json:"Total Data Reduction Savings"`
	TotalVolumeCapacity       string `json:"Total Volume Capacity"`
	TurboPerformance          string `json:"Turbo Performance"`
	Type                      string `json:"Type"`
	UnassignedVolumeSpace     string `json:"Unassigned Volume Space"`
	UnprotectedVolumes        string `json:"Unprotected Volumes"`
	UsedPoolSpace             string `json:"Used Pool Space"`
	UsedSpace                 string `json:"Used Space"`
	VDiskMirrors              string `json:"VDisk Mirrors"`
	Vendor                    string `json:"Vendor"`
	VirtualAllocation         string `json:"Virtual Allocation"`
	Volumes                   string `json:"Volumes"`
	WriteCache                string `json:"Write Cache"`
	ID                        string `json:"id"`
}

type Volumes []struct {
	VolumeUniqueID string `json:"Volume Unique ID"`
	ID             string `json:"id"`
}

type Switch struct {
	Acknowledged                  string `json:"Acknowledged"`
	ConnectedFabrics              string `json:"Connected Fabrics"`
	ConnectedPorts                string `json:"Connected Ports"`
	CustomTag1                    string `json:"Custom Tag 1"`
	CustomTag2                    string `json:"Custom Tag 2"`
	CustomTag3                    string `json:"Custom Tag 3"`
	DataSourceCount               string `json:"Data Source Count"`
	DomainID                      string `json:"Domain ID"`
	Fabric                        string `json:"Fabric"`
	Firmware                      string `json:"Firmware"`
	IPAddress                     string `json:"IP Address"`
	LastSuccessfulMonitor         string `json:"Last Successful Monitor"`
	LastSuccessfulProbe           string `json:"Last Successful Probe"`
	Links                         string `json:"Links"`
	Location                      string `json:"Location"`
	Mode                          string `json:"Mode"`
	Model                         string `json:"Model"`
	Name                          string `json:"Name"`
	ParentSwitch                  string `json:"Parent Switch"`
	PerformanceMonitorIntervalMin string `json:"Performance Monitor Interval (min)"`
	PerformanceMonitorStatus      string `json:"Performance Monitor Status"`
	Ports                         string `json:"Ports"`
	PrincipalSwitchOfFabric       string `json:"Principal Switch of Fabric"`
	ProbeSchedule                 string `json:"Probe Schedule"`
	ProbeStatus                   string `json:"Probe Status"`
	SerialNumber                  string `json:"Serial Number"`
	Status                        string `json:"Status"`
	Vendor                        string `json:"Vendor"`
	Virtual                       string `json:"Virtual"`
	WWN                           string `json:"WWN"`
	ID                            string `json:"id"`
}

type Pool struct {
	Acknowledged                string `json:"Acknowledged"`
	Activity                    string `json:"Activity"`
	AllocatedSpace              string `json:"Allocated Space"`
	AssignedVolumeSpace         string `json:"Assigned Volume Space"`
	AvailablePoolSpace          string `json:"Available Pool Space"`
	AvailableRepositorySpace    string `json:"Available Repository Space"`
	AvailableSoftSpace          string `json:"Available Soft Space"`
	BackEndStorageDiskType      string `json:"Back-end Storage Disk Type"`
	BackEndStorageDisks         string `json:"Back-end Storage Disks"`
	BackEndStorageRAIDLevel     string `json:"Back-end Storage RAID Level"`
	BackEndStorageSystemType    string `json:"Back-end Storage System Type"`
	Capacity                    string `json:"Capacity"`
	CapacityPool                string `json:"Capacity Pool"`
	CompressionSavings          string `json:"Compression Savings"`
	CustomTag1                  string `json:"Custom Tag 1"`
	CustomTag2                  string `json:"Custom Tag 2"`
	CustomTag3                  string `json:"Custom Tag 3"`
	DeduplicationSavings        string `json:"Deduplication Savings"`
	EasyTier                    string `json:"Easy Tier"`
	Encryption                  string `json:"Encryption"`
	EncryptionGroup             string `json:"Encryption Group"`
	EnterpriseHDDAvailableSpace string `json:"Enterprise HDD Available Space"`
	EnterpriseHDDCapacity       string `json:"Enterprise HDD Capacity"`
	ExtentSize                  string `json:"Extent Size"`
	Format                      string `json:"Format"`
	LSSOrLCU                    string `json:"LSS or LCU"`
	LastDataCollection          string `json:"Last Data Collection"`
	ManagedDisks                string `json:"Managed Disks"`
	Name                        string `json:"Name"`
	NearlineHDDAvailableSpace   string `json:"Nearline HDD Available Space"`
	NearlineHDDCapacity         string `json:"Nearline HDD Capacity"`
	OwnerName                   string `json:"Owner Name"`
	ParentName                  string `json:"Parent Name"`
	PhysicalAllocation          string `json:"Physical Allocation"`
	PoolAttributes              string `json:"Pool Attributes"`
	RAIDLevel                   string `json:"RAID Level"`
	RankGroup                   string `json:"Rank Group"`
	RepositoryCapacity          string `json:"Repository Capacity"`
	ReservedPoolSpace           string `json:"Reserved Pool Space"`
	Shortfall                   string `json:"Shortfall"`
	SoftSpace                   string `json:"Soft Space"`
	SolidState                  string `json:"Solid State"`
	Status                      string `json:"Status"`
	StorageSystem               string `json:"Storage System"`
	Tier                        string `json:"Tier"`
	Tier0FlashAvailableSpace    string `json:"Tier 0 Flash Available Space"`
	Tier0FlashCapacity          string `json:"Tier 0 Flash Capacity"`
	Tier1FlashAvailableSpace    string `json:"Tier 1 Flash Available Space"`
	Tier1FlashCapacity          string `json:"Tier 1 Flash Capacity"`
	TotalDataReductionSavings   string `json:"Total Data Reduction Savings"`
	TotalVolumeCapacity         string `json:"Total Volume Capacity"`
	UnallocatableVolumeSpace    string `json:"Unallocatable Volume Space"`
	UnallocatedVolumeSpace      string `json:"Unallocated Volume Space"`
	UnassignedVolumeSpace       string `json:"Unassigned Volume Space"`
	UnreservedPoolSpace         string `json:"Unreserved Pool Space"`
	UnusedSpace                 string `json:"Unused Space"`
	UsedSpace                   string `json:"Used Space"`
	VirtualAllocation           string `json:"Virtual Allocation"`
	Volumes                     string `json:"Volumes"`
	ZeroCapacity                string `json:"Zero Capacity"`
	ID                          string `json:"id"`
}
