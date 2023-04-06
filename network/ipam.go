package network

import (
	"encoding/json"
	log "github.com/siruspen/logrus"
	"net"
	"os"
	"path"
	"strings"
)

// 默认的分配信息存储位置
const ipamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

// IPAM用于网络IP地址的分配和释放
type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string //类似位图表示哪些ip被分配了
}

// 初始化一个IPAM对象，使用默认位置存储信息
var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

// 将子网配置读出来
func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipAllocator.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		log.Errorf("error:%v", err)
		return err
	}
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		log.Errorf("error:%v", err)
		return err
	}
	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		log.Errorf("error:%v", err)
		return err
	}
	return nil
}

// 将子网配置放到文件系统中
func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			log.Errorf("error:%v", err)
			return err
		}
	}
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		log.Errorf("error:%v", err)
		return err
	}
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		log.Errorf("error:%v", err)
		return err
	}
	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		log.Errorf("error:%v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	ipam.Subnets = &map[string]string{}
	err = ipam.load()
	if err != nil {
		log.Errorf("Error dump allocate info,%v", err)
	}
	_, subnet, _ = net.ParseCIDR(subnet.String())

	// 返回子网掩码前面1的个数和掩码总长度
	one, size := subnet.Mask.Size()

	// 如果之前没有分配过这个网址，就初始化网段的分配配置
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// 0的个数就是网段中可用ip数
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	// 遍历网段位图数组
	for c := range (*ipam.Subnets)[subnet.String()] {
		// 找到可用的ip
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}

	ipam.dump()
	return
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		log.Errorf("Error dump allocation info, %v", err)
	}

	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	ipam.dump()
	return nil
}
