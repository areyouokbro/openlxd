package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Container 容器模型
type Container struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Image             string    `json:"image"`
	Password          string    `json:"password"`
	AccessCode        string    `json:"access_code"`
	CPU               int       `json:"cpu"`
	Memory            int       `json:"memory"`
	Disk              int       `json:"disk"`
	BandwidthIn       int       `json:"bandwidth_in"`
	BandwidthOut      int       `json:"bandwidth_out"`
	TrafficLimit      int64     `json:"traffic_limit"`
	TrafficUsed       int64     `json:"traffic_used"`
	IPv4PoolQuota     int       `json:"ipv4_pool_quota"`
	IPv4MappingQuota  int       `json:"ipv4_mapping_quota"`
	IPv6PoolQuota     int       `json:"ipv6_pool_quota"`
	IPv6MappingQuota  int       `json:"ipv6_mapping_quota"`
	ProxyQuota        int       `json:"proxy_quota"`
	CPULimit          int       `json:"cpu_limit"`
	IORead            int       `json:"io_read"`
	IOWrite           int       `json:"io_write"`
	MaxProcesses      int       `json:"max_processes"`
	Nested            bool      `json:"nested"`
	Swap              bool      `json:"swap"`
	Privileged        bool      `json:"privileged"`
	Remark            string    `json:"remark"`
	CreatedAt         time.Time `json:"created_at"`
}

// IPv4Address IPv4地址
type IPv4Address struct {
	ID            int       `json:"id"`
	IP            string    `json:"ip"`
	ContainerName string    `json:"container_name"`
	Allocated     bool      `json:"allocated"`
	CreatedAt     time.Time `json:"created_at"`
}

// IPv4Mapping IPv4端口映射
type IPv4Mapping struct {
	ID            int       `json:"id"`
	ContainerName string    `json:"container_name"`
	PublicIP      string    `json:"public_ip"`
	PublicPort    int       `json:"public_port"`
	ContainerPort int       `json:"container_port"`
	Protocol      string    `json:"protocol"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// IPv6Address IPv6地址
type IPv6Address struct {
	ID            int       `json:"id"`
	IP            string    `json:"ip"`
	ContainerName string    `json:"container_name"`
	Allocated     bool      `json:"allocated"`
	CreatedAt     time.Time `json:"created_at"`
}

// IPv6Mapping IPv6端口映射
type IPv6Mapping struct {
	ID            int       `json:"id"`
	ContainerName string    `json:"container_name"`
	PublicIP      string    `json:"public_ip"`
	PublicPort    int       `json:"public_port"`
	ContainerPort int       `json:"container_port"`
	Protocol      string    `json:"protocol"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// Proxy 反向代理
type Proxy struct {
	ID            int       `json:"id"`
	ContainerName string    `json:"container_name"`
	Domain        string    `json:"domain"`
	ContainerPort int       `json:"container_port"`
	HTTPS         bool      `json:"https"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// Metric 监控数据
type Metric struct {
	ID            int       `json:"id"`
	ContainerName string    `json:"container_name"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   int64     `json:"memory_usage"`
	DiskUsage     int64     `json:"disk_usage"`
	TrafficRX     int64     `json:"traffic_rx"`
	TrafficTX     int64     `json:"traffic_tx"`
	Timestamp     time.Time `json:"timestamp"`
}

// Init 初始化数据库
func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 创建表
	if err = createTables(); err != nil {
		return err
	}

	log.Println("数据库初始化成功")
	return nil
}

// createTables 创建所有表
func createTables() error {
	schemas := []string{
		// 容器表
		`CREATE TABLE IF NOT EXISTS containers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			image TEXT,
			password TEXT,
			access_code TEXT,
			cpu INTEGER DEFAULT 1,
			memory INTEGER DEFAULT 512,
			disk INTEGER DEFAULT 10,
			bandwidth_in INTEGER DEFAULT 100,
			bandwidth_out INTEGER DEFAULT 100,
			traffic_limit INTEGER DEFAULT 1073741824,
			traffic_used INTEGER DEFAULT 0,
			ipv4_pool_quota INTEGER DEFAULT 0,
			ipv4_mapping_quota INTEGER DEFAULT 10,
			ipv6_pool_quota INTEGER DEFAULT 0,
			ipv6_mapping_quota INTEGER DEFAULT 10,
			proxy_quota INTEGER DEFAULT 5,
			cpu_limit INTEGER DEFAULT 100,
			io_read INTEGER DEFAULT 0,
			io_write INTEGER DEFAULT 0,
			max_processes INTEGER DEFAULT 0,
			nested BOOLEAN DEFAULT 0,
			swap BOOLEAN DEFAULT 1,
			privileged BOOLEAN DEFAULT 0,
			remark TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// IPv4地址池
		`CREATE TABLE IF NOT EXISTS ipv4_pool (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip TEXT UNIQUE NOT NULL,
			container_name TEXT,
			allocated BOOLEAN DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// IPv4端口映射
		`CREATE TABLE IF NOT EXISTS ipv4_mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_name TEXT NOT NULL,
			public_ip TEXT,
			public_port INTEGER,
			container_port INTEGER,
			protocol TEXT DEFAULT 'tcp',
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// IPv6地址池
		`CREATE TABLE IF NOT EXISTS ipv6_pool (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip TEXT UNIQUE NOT NULL,
			container_name TEXT,
			allocated BOOLEAN DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// IPv6端口映射
		`CREATE TABLE IF NOT EXISTS ipv6_mappings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_name TEXT NOT NULL,
			public_ip TEXT,
			public_port INTEGER,
			container_port INTEGER,
			protocol TEXT DEFAULT 'tcp',
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// 反向代理
		`CREATE TABLE IF NOT EXISTS proxies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_name TEXT NOT NULL,
			domain TEXT UNIQUE NOT NULL,
			container_port INTEGER,
			https BOOLEAN DEFAULT 0,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// 监控数据
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_name TEXT NOT NULL,
			cpu_usage REAL,
			memory_usage INTEGER,
			disk_usage INTEGER,
			traffic_rx INTEGER,
			traffic_tx INTEGER,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, schema := range schemas {
		if _, err := DB.Exec(schema); err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}
	}

	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
