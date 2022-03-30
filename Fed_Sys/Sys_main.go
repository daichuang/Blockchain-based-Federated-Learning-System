package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sbinet/go-python"

	"errors"
	"flag"
)

type NoiseVector struct {
	noise []float64
}

var (
	outlier_index []int
	InitWeight    []float64
	//Input arguments
	datasetName       string
	numberOfNodes     int
	collectingUpdates bool

	numberOfNodeUpdates int
	myIP                string
	myPrivateIP         string
	myPort              string
	peersFileName       string

	client  Communication_object
	nodeNum int

	allUpdatesReceived  chan bool
	networkBootstrapped chan bool
	blockReceived       chan bool

	portsToConnect     []string
	MinerAddressToConn string
	peerPorts          []string
	peerLookup         map[string]int
	peerAddresses      map[int]net.TCPAddr
	stakeMap           map[int]int

	//Locks
	updateLock     sync.Mutex
	boolLock       sync.Mutex
	convergedLock  sync.Mutex
	peerLock       sync.Mutex
	blockChainLock sync.Mutex

	ensureRPC sync.WaitGroup

	// global shared variables
	updateSent     bool
	converged      bool
	miner          bool
	iterationCount = -1

	// these are maps since it optimizes contains()
	roleIDs map[int]int

	//Logging
	errLog *log.Logger = log.New(os.Stderr, "[错误信息] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)
	outLog *log.Logger = log.New(os.Stderr, "[对等节点] ", log.Lshortfile|log.LUTC|log.Lmicroseconds)

	//Errors
	updateError error = errors.New("更新被拒绝")
	rpcError    error = errors.New("RPC超时")

	disable_dp bool    = false
	epsilon    float64 = 999

	timeoutRPC    time.Duration = 120 * time.Second
	timeoutUpdate time.Duration = 90 * time.Second
	timeoutBlock  time.Duration = 300 * time.Second
	timeoutPeer   time.Duration = 5 * time.Second

	byzantine_index    float64 = 0.0
	numComputeNode     int     = 0
	all_train_data_num float64 = 0

	block_momentum float64 = 0
	use_nbm        bool    = true
	block_lr       float64 = 1
	byzantineNum   int     = 0
)

const (
	basePort       int = 3000
	DEFAULT_STAKE  int = 10
	MAX_ITERATIONS int = 199
)

func main() {
	//Register和RegisterName解决的主要问题是：当编解码中有一个字段是interface{}（interface{}的赋值为map、结构体时）
	//的时候需要对interface{}的可能产生的类型进行注册。
	gob.Register(&net.TCPAddr{})
	gob.Register(&Blockchain{})

	// *********************** 0	、 system args init  *********************** //
	numberOfNodesPtr := flag.Int("total_nodes", 0, "网络中的节点总数")
	nodeNumPtr := flag.Int("index", -1, "网络中节点的下标（index）")
	datasetNamePtr := flag.String("dataset", "", "需要使用的数据集")
	peersFileNamePtr := flag.String("Peer_file", "", "包含对等节点' IP:端口'列表的文件")
	myIPPtr := flag.String("ip_address", "", "如果不在本地训练，本节点IP")
	myPrivateIPPtr := flag.String("private_ip", "", "如果不在本地训练，本节点私有IP")
	myPortPtr := flag.String("port", "", "如果不在本地训练，本节点端口")
	PrivatePtr := flag.Bool("off_private_train", false, "差分隐私所使用的参数Epsilon") //差分隐私所使用的参数
	epsilonPtr := flag.Float64("epsilon", 2.0, "差分隐私所使用的参数Epsilon")          //差分隐私所使用的参数
	byzantinePtr := flag.Float64("byzantine_thresh", 0.0, "byzantine阈值")     //byzantine阈值阈值

	flag.Parse()

	nodeNum := *nodeNumPtr
	numberOfNodes = *numberOfNodesPtr
	datasetName = *datasetNamePtr
	peersFileName = *peersFileNamePtr
	myPrivateIP = *myPrivateIPPtr + ":"
	myIP = *myIPPtr + ":"
	myPort = *myPortPtr
	disable_dp = *PrivatePtr
	epsilon = *epsilonPtr
	byzantine_index = *byzantinePtr

	//关键参数打印
	numComputeNode = numberOfNodes - 1                          //参与计算节点数量
	block_momentum = 1 - float64(1.0/float64(numComputeNode-1)) //bm =  1- 1/N
	// block_momentum = 0

	outLog.Printf("参与计算的节点数量为 : %d", numComputeNode)
	outLog.Printf("参与计算的节点block_momentum为 : %f", block_momentum)
	if !disable_dp {
		outLog.Print("节点本地训练中使用差分隐私")
		outLog.Print("差分隐私参数 epsilon 为 : ", epsilon)
	}

	//参数不正确 ， 退出
	if numberOfNodes <= 0 || nodeNum < 0 || datasetName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// *********************** 1、 Peer to Peer  network args init *********************** //
	// P2P网络 ip ：port
	peerLookup = make(map[string]int)

	//初始化权益列表 stack = Miner_reward + untrimmed_grad_reward
	stakeMap = make(map[int]int)

	potentialPeerList := make([]net.TCPAddr, 0, numberOfNodes-1)

	//本地
	if peersFileName == "" {
		myIP = "127.0.0.1:"
		myPort = strconv.Itoa(nodeNum + basePort) // myPort = 8000 、8001 、8002...

		for i := 0; i < numberOfNodes; i++ {

			peerPort := strconv.Itoa(basePort + i) // 8000 、 8001 、8002 ...

			//如果是自己 ， 不需要加入到potentialPeerList
			if peerPort == myPort {
				peerLookup[fmt.Sprintf(myIP+peerPort)] = i //   "127.0.0.1 : 8000 " : 0
				stakeMap[i] = DEFAULT_STAKE
				continue
			}

			// 其他地址 ， 在potentialPeerList列表中加入他们的 tcp socket ,also peerPorts
			peerPorts = append(peerPorts, peerPort)
			peerAddress, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(myIP+peerPort))
			handleErrorFatal("无法解析潜在的对等地址", err)
			potentialPeerList = append(potentialPeerList, *peerAddress)
			peerLookup[fmt.Sprintf(myIP+peerPort)] = i
			stakeMap[i] = DEFAULT_STAKE
		}
		peerAddresses = make(map[int]net.TCPAddr)
	} else if myIP == ":" || myPort == "" || myPrivateIP == ":" {
		flag.PrintDefaults()
		os.Exit(1)
	} else { //分布式
		file, err := os.Open(peersFileName)
		handleErrorFatal(" 打开对等节点文件错误 ", err)
		defer file.Close()

		scanner := bufio.NewScanner(file)
		nodeInList := false
		i := -1

		for scanner.Scan() {
			i++
			peerAddressStr := scanner.Text() //同伴地址 string

			if strings.Contains(peerAddressStr, myIP) &&
				strings.Contains(peerAddressStr, myPort) {
				nodeInList = true
				peerLookup[peerAddressStr] = i
				stakeMap[i] = DEFAULT_STAKE
				continue
			}

			peerAddress, err := net.ResolveTCPAddr("tcp", peerAddressStr)
			handleErrorFatal("无法解析潜在的对等地址", err)
			potentialPeerList = append(potentialPeerList, *peerAddress)
			peerLookup[peerAddressStr] = i
			stakeMap[i] = DEFAULT_STAKE
		}

		if !nodeInList {
			handleErrorFatal("此节点不在对等节点列表中", err)
		}
	}

	peerAddresses = make(map[int]net.TCPAddr)
	// *********************** 2、 client init 、 time controller *********************** //
	client = Communication_object{
		id:           nodeNum,
		blockUpdates: make([]NewDelta, 0, 5),
	}
	multiplier := time.Duration((numberOfNodes / 100))
	if multiplier >= 1 {
		timeoutUpdate = timeoutUpdate * multiplier
		timeoutBlock = timeoutBlock * multiplier
		timeoutPeer = timeoutPeer * multiplier
		timeoutRPC = timeoutRPC * multiplier
	}
	outLog.Printf("更新的时间限制为 %d", timeoutUpdate)

	pyInit()

	// Byzantine exp
	if byzantine_index > 0 {
		byzantine := int(math.Ceil(float64(numberOfNodes) * (1.0 - byzantine_index)))
		byzantineNum = int(byzantine_index * float64(numberOfNodes))
		outLog.Printf("byzantine node 数量为 %d ", byzantineNum)
		outLog.Printf("投毒攻击(byzantine node) index >= %d ", byzantine)
		if nodeNum >= byzantine {
			outLog.Printf("对等节点 " + strconv.Itoa(client.id) + ":我是拜占庭节点 ， 使用不真实数据集进行训练， 上传拜占庭梯度 ")
			CreateClient(datasetName, "mnist_bad", disable_dp, epsilon)
			InitWeight = getInitWeight()
		} else {
			CreateClient(datasetName, datasetName+strconv.Itoa(client.id), disable_dp, epsilon)
			InitWeight = getInitWeight()
		}
		client.initBlockChain(InitWeight, datasetName)
	} else {
		CreateClient(datasetName, datasetName+strconv.Itoa(client.id), disable_dp, epsilon)
		InitWeight = getInitWeight()
		client.initBlockChain(InitWeight, datasetName)
	}

	converged = false

	updateLock = sync.Mutex{}
	boolLock = sync.Mutex{}
	convergedLock = sync.Mutex{}
	peerLock = sync.Mutex{}
	blockChainLock = sync.Mutex{}

	ensureRPC = sync.WaitGroup{}
	allUpdatesReceived = make(chan bool)
	networkBootstrapped = make(chan bool)
	blockReceived = make(chan bool)

	// p2p server
	peer := new(Peer)
	peerServer := rpc.NewServer()
	peerServer.Register(peer)

	state := python.PyEval_SaveThread()

	go messageListener(peerServer, myPort)

	//向超过预期的同伴宣布你自己。网络中的第一个节点不需要这样做。取而代之的是，他在等待一个即将到来的同伴。
	//无论你是哪个节点，你都不能前进，除非你向你的 Peer 宣布你自己
	if myPort != strconv.Itoa(basePort) {
		go announceToNetwork(potentialPeerList)
	}

	<-networkBootstrapped

	nextCommunicationRound()
	ComputeProcess(peerPorts)
	python.PyEval_RestoreThread(state)

}

func handleErrorFatal(s string, err error) {
	if err != nil {
		errLog.Fatalf("%s, err = %s\n", s, err.Error())
	}
}
