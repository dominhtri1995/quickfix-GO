package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"time"
	"github.com/rs/xid"
	"golang.org/x/sync/syncmap"
	"sync"
)

func (e TradeClient) OnCreate(sessionID quickfix.SessionID) {
	return
}
func (e TradeClient) OnLogon(sessionID quickfix.SessionID) {
	fmt.Printf(" %s  Session created !! Ready to rock and roll\n", sessionID.TargetCompID)
	return
}
func (e TradeClient) OnLogout(sessionID quickfix.SessionID) {
	fmt.Printf(" %s  logged out!!!!!! It's Obama's fault !!! \n", sessionID.TargetCompID)
	return
}
func (e TradeClient) FromAdmin(msg quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	//msg.Build()
	//fmt.Printf("Incoming %s\n", &msg)
	return
}
func (e TradeClient) ToAdmin(msg quickfix.Message, sessionID quickfix.SessionID) {
	//msg.Build()
	//fmt.Printf("Outgoing %s\n", &msg)

	messageType, _ := msg.Header.GetString(quickfix.Tag(35))

	if messageType == "A" {
		msg.Header.SetString(quickfix.Tag(96), "12345678")
		msg.Header.SetInt(quickfix.Tag(95), 8)
		t34, err := msg.Header.GetInt(quickfix.Tag(34))
		if t34 == 1 && err == nil && sessionID.TargetCompID == "TTDEV18O" {
			msg.Header.SetBool(quickfix.Tag(141), true) //Set reset sequence
		}
	}
	return
}

func (e TradeClient) ToApp(msg quickfix.Message, sessionID quickfix.SessionID) (err error) {
	msg.Build()
	fmt.Printf("Sending %s\n", &msg)
	return
}

//FromApp implemented as part of Application interface. This is the callback for all Application level messages from the counter party.
func (e TradeClient) FromApp(msg quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {
	msg.Build()
	fmt.Printf("Receiving %s\n", &msg)
	messageType, err := msg.Header.GetString(quickfix.Tag(35))
	if ( err != nil) {
		return
	}
	switch {
	/*  EXECUTION REPORT -EXECUTION REPORT -EXECUTION REPORT -EXECUTION REPORT*/
	case messageType == string(enum.MsgType_EXECUTION_REPORT):
		execrefID, _ := msg.Body.GetString(quickfix.Tag(20))
		account, _ := msg.Body.GetString(quickfix.Tag(1))
		ordStatus, _ := msg.Body.GetString(quickfix.Tag(39))

		if (execrefID != "3") { //NOT order status request

			switch {
			case ordStatus == string(enum.OrdStatus_NEW):
				// New Order
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := newOrderMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.status = "ok"
					order.account = account
					extractInfoExcecutionReport(order, msg) //Get all the order-related info :trade,price plapla
					order.channel <- *order                 //Send data back to channel
					newOrderMap.Delete(clOrdID)             // Remove Order status request from list
				}
			case ordStatus == string(enum.OrdStatus_REJECTED):
				//rejected
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := newOrderMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.status = "rejected"
					order.account = account
					order.reason, _ = msg.Body.GetString(quickfix.Tag(58))
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					newOrderMap.Delete(clOrdID)
				}
			case ordStatus == string(enum.OrdStatus_CANCELED):
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.status = "ok"
					order.account = account
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					cancelAndUpdateMap.Delete(clOrdID)
				}
			case ordStatus == string(enum.OrdStatus_REPLACED):
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.status = "ok"
					order.account = account
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					cancelAndUpdateMap.Delete(clOrdID)
				}
			case ordStatus == string(enum.OrdStatus_PARTIALLY_FILLED):
				//Partially filled
				//priceFilled,_ := msg.Body.GetInt(quickfix.Tag(31))
				//qtyFilled,_ := msg.Body.GetInt(quickfix.Tag(32)) // qty just got filled
				// 	totalqty,_ := msg.Body.GetString(quickfix.Tag(38))
				// 	symbol,_ := msg.Body.GetString(quickfix.Tag(55))
				// 	side , _ := msg.Body.GetString(quickfix.Tag(54))
				// avgPx, _ := msg.Body.GetString(quickfix.Tag(6))

			case ordStatus == string(enum.OrdStatus_FILLED):
				//Fully filled
				//priceFilled,_ := msg.Body.GetInt(quickfix.Tag(31))
				//qtyFilled,_ := msg.Body.GetInt(quickfix.Tag(32)) // qty just got filled
				// 	totalqty,_ := msg.Body.GetString(quickfix.Tag(38))
				// 	symbol,_ := msg.Body.GetString(quickfix.Tag(55))
				// 	side , _ := msg.Body.GetString(quickfix.Tag(54))
				// avgPx, _ := msg.Body.GetString(quickfix.Tag(6))

			}
		} else {
			//Order status request
			account, err := msg.Body.GetString(quickfix.Tag(1))
			numPosReports, _ := msg.Body.GetString(quickfix.Tag(16728))
			for osr := range orderStatusRequestList.Iter() {
				orderStatusRequest, _ := osr.Value.(*OrderStatusReq)
				if err == nil && account == orderStatusRequest.account { // GET book order not single order status request
					orderStatusRequest.count, _ = strconv.Atoi(numPosReports)

					var order WorkingOrder
					order.orderID, _ = msg.Body.GetString(quickfix.Tag(37))
					order.price, _ = msg.Body.GetString(quickfix.Tag(44))
					order.quantity, _ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
					order.filledQuantity, _ = msg.Body.GetString(quickfix.Tag(14))
					order.originalQuantity, _ = msg.Body.GetString(quickfix.Tag(38))
					order.ordStatus, _ = msg.Body.GetString(quickfix.Tag(39))
					order.symbol, _ = msg.Body.GetString(quickfix.Tag(55))
					order.exchange, _ = msg.Body.GetString(quickfix.Tag(207))
					order.sideNum, _ = msg.Body.GetString(quickfix.Tag(54))
					order.ordType, _ = msg.Body.GetString(quickfix.Tag(40))
					order.timeInForce, _ = msg.Body.GetString(quickfix.Tag(59))
					order.securityID, _ = msg.Body.GetString(quickfix.Tag(48))
					order.text, _ = msg.Body.GetString(quickfix.Tag(58))
					order.strikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
					putOrCall, _ := msg.Body.GetInt(quickfix.Tag(201))

					if putOrCall == 0 {
						order.putOrCall = "Put"
					} else {
						order.putOrCall = "Call"
					}

					if order.sideNum == "1" {
						order.side = "Buy"
					} else {
						order.side = "Sell"
					}
					order.productType, _ = msg.Body.GetString(quickfix.Tag(167))
					if order.productType == "FUT" || order.productType == "OPT" || order.productType == "NRG" {
						order.productMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
					}
					orderStatusRequest.workingOrders = append(orderStatusRequest.workingOrders, order)
				}

				if orderStatusRequest.count == len(orderStatusRequest.workingOrders) {
					// Receive all working orders
					fmt.Printf("Receve all working orders : %d for account %s \n", len(orderStatusRequest.workingOrders), orderStatusRequest.account)
					orderStatusRequest.status = "ok"
					orderStatusRequest.channel <- *orderStatusRequest
					orderStatusRequestList.remove(osr.Index)
					break
				}
			}
		}
	case messageType == string(enum.MsgType_ORDER_CANCEL_REJECT):
		clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
		if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
			order, _ := order.(*OrderConfirmation)
			order.status = "rejected"
			order.reason, _ = msg.Body.GetString(quickfix.Tag(58))
			extractInfoExcecutionReport(order, msg)
			order.channel <- *order
			cancelAndUpdateMap.Delete(clOrdID)
		}
		/* UAP - UAP - UAP - UAP - UAP - UAP - UAP */
	case messageType == "UAP":
		// fmt.Printf("Receiving UAP \n")
		var uid, _ = msg.Body.GetString(quickfix.Tag(16710))
		if uan, ok := uanMap.Load(uid); ok {
			uan, _ := uan.(*UAN)
			numPosReports, _ := msg.Body.GetString(quickfix.Tag(16727))
			uan.count, _ = strconv.Atoi(numPosReports)

			//Create a new UAP object
			var uap UAPreport
			uap.quantity, _ = msg.Body.GetString(quickfix.Tag(32))
			q, _ := strconv.Atoi(uap.quantity)

			putOrCall, _ := msg.Body.GetInt(quickfix.Tag(201))

			if putOrCall == 0 {
				uap.putOrCall = "Put"
			} else {
				uap.putOrCall = "Call"
			}

			if (q > 0) {
				uap.side = "Buy"
			} else {
				uap.side = "Sell"
				uap.quantity = string(q * (-1))
			}
			uap.accountGroup, _ = msg.Header.GetString(quickfix.Tag(50))
			uap.account, _ = msg.Body.GetString(quickfix.Tag(1))
			// fmt.Println(uap.accountGroup)
			uap.price, _ = msg.Body.GetString(quickfix.Tag(31))
			uap.securityID, _ = msg.Body.GetString(quickfix.Tag(48))
			uap.product, _ = msg.Body.GetString(quickfix.Tag(55))
			productType, _ := msg.Body.GetString(quickfix.Tag(167))
			uap.strikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
			if (productType == "FUT") {
				uap.productMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
			}
			uan.reports = append(uan.reports, uap)

			//UAN Complete
			posReqType, _ := msg.Body.GetString(quickfix.Tag(16724))
			if len(uan.reports) == uan.count {
				for j := 0; j < len(uan.reports); j++ {
					if posReqType == "4" && uan.accountGroup != uan.reports[j].accountGroup {
						uan.reports = append(uan.reports[:j], uan.reports[j+1:]...)
						j--
					} else if posReqType == "0" && uan.account != uan.reports[j].account {
						uan.reports = append(uan.reports[:j], uan.reports[j+1:]...)
						j--
					}
				}
				fmt.Printf("Number of positions :%d for trader %s \n", len(uan.reports), uan.accountGroup)
				uan.channel <- *uan // return the result to channel
				uanMap.Delete(uid)
			}
		}
	case messageType == "W": //Market data request
		uid, _ := msg.Body.GetString(quickfix.Tag(262))
		if md, ok := marketDataRequestMap.Load(uid); ok {
			md, _ := md.(*MarketDataReq)
			//create market data reqest response
			md.price, err = msg.Body.GetString(quickfix.Tag(270))
			if ( err != nil) {
				md.price = "0" //Price not available
				md.status = "rejected"
				md.reason = "price not available"
			} else {
				md.symbol, _ = msg.Body.GetString(quickfix.Tag(55))
				md.exchange, _ = msg.Body.GetString(quickfix.Tag(207))
				productType, _ := msg.Body.GetString(quickfix.Tag(167))
				if (productType == "FUT") {
					md.productMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
				}

				md.priceType, _ = msg.Body.GetString(quickfix.Tag(269))
				md.status = "ok"
			}

			md.channel <- *md
			marketDataRequestMap.Delete(uid)
		}
	case messageType == "Y": //Market data reject
		uid, _ := msg.Body.GetString(quickfix.Tag(262))
		if md, ok := marketDataRequestMap.Load(uid); ok {
			md, _ := md.(*MarketDataReq)
			md.status = "rejected"
			md.reason, _ = msg.Body.GetString(quickfix.Tag(58))
			md.channel <- *md
			marketDataRequestMap.Delete(uid)
		}
	case messageType == "d":
		id, _ := msg.Body.GetString(quickfix.Tag(320))
		count,_ := msg.Body.GetInt(quickfix.Tag(393))
		if sdr, ok := securityDefinitionMap.Load(id); ok {
			sdr, _ := sdr.(*SecurityDefinitionReq)
			sdr.count =count

			var security Security
			security.symbol, _ = msg.Body.GetString(quickfix.Tag(55))
			security.exchange, _ = msg.Body.GetString(quickfix.Tag(207))
			security.productType, _ = msg.Body.GetString(quickfix.Tag(167))
			security.securityID, _ = msg.Body.GetString(quickfix.Tag(48))
			security.securityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
			security.putOrCall, _ = msg.Body.GetString(quickfix.Tag(201))
			security.strikePrice, _ = msg.Body.GetString(quickfix.Tag(202))

			exchTickSize, _ := msg.Body.GetString(quickfix.Tag(16552))
			exchPointValue, _ := msg.Body.GetString(quickfix.Tag(16554))
			security.exchTickSize,_ = strconv.ParseFloat(exchTickSize,64)
			security.exchPointValue,_  = strconv.ParseFloat(exchPointValue,64)
			security.numTicktblEntries,_ = msg.Body.GetInt(quickfix.Tag(16456))

			security.tickValue, security.tickSize = calculateTickValueAndSize(security.exchTickSize,security.exchPointValue,0,0,"0","0")
			//tag16552 int,tag16554 int,tag16456 int, tag16457 int,tag16458 string
			sdr.securityList = append(sdr.securityList, security)
			if sdr.count == len(sdr.securityList){
				fmt.Println("receive all security definition")
				sdr.status = "ok"
				sdr.channel <- *sdr
				securityDefinitionMap.Delete(id)
			}
		}
	}
	return
}

var orderStatusRequestList ConcurrentSlice
var uanMap syncmap.Map
var newOrderMap syncmap.Map
var cancelAndUpdateMap syncmap.Map
var marketDataRequestMap syncmap.Map
var securityDefinitionMap syncmap.Map

func StartQuickFix() {
	flag.Parse()

	cfgFileName := path.Join("config", "tradeclient.cfg")
	if flag.NArg() > 0 {
		cfgFileName = flag.Arg(0)
	}

	cfg, err := os.Open(cfgFileName)
	if err != nil {
		fmt.Printf("Error opening %v, %v\n", cfgFileName, err)
		return
	}

	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		fmt.Println("Error reading cfg,", err)
		return
	}

	app := TradeClient{}
	fileLogFactory, err := quickfix.NewFileLogFactory(appSettings)
	//screenLogFactory := quickfix.NewScreenLogFactory()

	if err != nil {
		fmt.Println("Error creating file log factory,", err)
		return
	}

	initiator, err := quickfix.NewInitiator(app, quickfix.NewMemoryStoreFactory(), appSettings, fileLogFactory)
	if err != nil {
		fmt.Printf("Unable to create Initiator: %s\n", err)
		return
	}
	initiator.Start()

}

// Wrapper fro Query function in console.go, using channel to wait for all message come through.
/*
/* Use these function to avoid using channel in main code
/*	Sepearte TT from mistro code
/*  For easy debug and maintain code with TT
 */
func TT_PAndLSOD(id string, account string, accountGroup string, sender string) (uan UAN) {
	c := make(chan UAN)
	QueryPAndLSOD(id, accountGroup, sender, c) // Get SOD report for position upto today

	var uan1 UAN
	c1 := make(chan UAN)
	QueryPAndLPos(xid.New().String(), account, sender, c1) // Get Position report for positions placed today

	select {
	case uan = <-c:

	case <-getTimeOutChan():
		var uan UAN
		uan.status = "rejected"
		uan.reason = "time out"
	}
	select {
	case uan1 = <-c1:
		if (uan.status != "rejected") {
			uan.reports = append(uan.reports, uan1.reports[0:]...)
			uan.count = len(uan.reports)
			uan.account = uan1.account
			uan.status = "ok"
		}
	case <-getTimeOutChan():
		var uan UAN
		uan.status = "rejected"
		uan.reason = "time out"
	}

	return uan
}
func TT_NewOrderSingle(id string, account string, side string, ordType string, quantity string, pri string, symbol string, exchange string, maturity string, productType string, timeInForce string, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryNewOrderSingle(id, account, side, ordType, quantity, pri, symbol, exchange, maturity, productType, timeInForce, sender, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.status = "rejected"
		ordStatus.reason = "time out"
	}
	return ordStatus
}

func TT_WorkingOrder(account string, sender string) (wo OrderStatusReq) {
	c := make(chan OrderStatusReq)
	QueryWorkingOrder(account, sender, c)
	select {
	case wo = <-c:
		return wo
	case <-getTimeOutChan():
		wo.status = "rejected"
		wo.reason = "time out"
	}
	return wo
}
func TT_OrderCancel(id string, orderID string, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryOrderCancel(id, orderID, sender, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.status = "rejected"
		ordStatus.reason = "time out"
	}
	return ordStatus
}
func TT_OrderCancelReplace(orderID string, newid string, account string, side string, ordType string, quantity string, pri string, symbol string, exchange string, maturity string, productType string, timeInForce string, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryOrderCancelReplace(orderID, newid, account, side, ordType, quantity, pri, symbol, exchange, maturity, productType, timeInForce, sender, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.status = "rejected"
		ordStatus.reason = "time out"
	}
	return ordStatus
}

func TT_MarketDataRequest(id string, requestType enum.SubscriptionRequestType, marketDepth int, priceType enum.MDEntryType, symbol string, exchange string, maturity string, productType string, sender string) (mdr MarketDataReq) {
	c := make(chan MarketDataReq)
	QueryMarketDataRequest(id, requestType, marketDepth, priceType, symbol, exchange, maturity, productType, sender, c)
	select {
	case mdr = <-c:
		return mdr
	case <-getTimeOutChan():
		mdr.status = "rejected"
		mdr.reason = "time out"
	}
	return mdr
}
func TT_QuerySecurityDefinitionRequest(id string, symbol string, exchange string, securityID string, productType string, sender string) (sdr SecurityDefinitionReq) {
	c := make(chan SecurityDefinitionReq)
	QuerySecurityDefinitionRequest(id, symbol, exchange, securityID, productType, sender, c)
	select {
	case sdr = <-c:
		return sdr
	case <-getTimeOutChan():
		sdr.status = "rejected"
		sdr.reason = "time out"
	}

	return sdr
}

func getTimeOutChan() chan bool {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		timeout <- true
	}()
	return timeout
}
func calculateTickValueAndSize(tag16552 float64, tag16554 float64, tag16456 int, tag16457 int, tag16458 string, price string) (tickValue float64, tickSize float64) {
	if tag16456 == 0 {
		tickSize = tag16552
		tickValue = tag16552 * tag16554
	} else {
		if price == "0"{
			return -1, -1
		}
		//baseTickSize := tag16552
		//p := float64(price)
		//maxPirce := float64(tag16458)
		//for i := 0; i < tag16456; i++ {
		//	if p < maxPirce {
		//		tickSize = baseTickSize * tag16457
		//		tickValue = tickSize * tag16554
		//		break
		//	}
		//}
	}

	return tickValue, tickSize
}

func extractInfoExcecutionReport(order *OrderConfirmation, msg quickfix.Message) {
	order.price, _ = msg.Body.GetString(quickfix.Tag(44))     //not available for market order
	order.quantity, _ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
	order.symbol, _ = msg.Body.GetString(quickfix.Tag(55))
	order.exchange, _ = msg.Body.GetString(quickfix.Tag(207))
	order.productType, _ = msg.Body.GetString(quickfix.Tag(167))
	order.ordType, _ = msg.Body.GetString(quickfix.Tag(40))
	order.sideNum, _ = msg.Body.GetString(quickfix.Tag(54))
	order.timeInForce, _ = msg.Body.GetString(quickfix.Tag(59))
	order.securityID, _ = msg.Body.GetString(quickfix.Tag(48))
	order.strikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
	putOrCall, _ := msg.Body.GetInt(quickfix.Tag(201))

	if putOrCall == 0 {
		order.putOrCall = "Put"
	} else {
		order.putOrCall = "Call"
	}
	if order.sideNum == "1" {
		order.side = "Buy"
	} else {
		order.side = "Sell"
	}
	if order.productType == "FUT" || order.productType == "OPT" || order.productType == "NRG" {
		order.productMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
	}
}

type TradeClient struct {
}
type UAN struct {
	id           string
	account      string
	accountGroup string
	count        int
	channel      chan UAN
	reports      []UAPreport
	status       string
	reason       string
}
type UAPreport struct {
	id              string
	accountGroup    string
	account         string
	quantity        string
	price           string
	side            string
	product         string
	productMaturity string
	exchange        string
	securityID      string
	putOrCall       string //for option only
	strikePrice     string //for option only
}

type OrderStatusReq struct {
	account       string
	count         int
	channel       chan OrderStatusReq
	workingOrders []WorkingOrder
	status        string
	reason        string
}
type WorkingOrder struct {
	orderID          string // Used to cancel order or request order status later
	price            string
	ordStatus        string
	quantity         string
	filledQuantity   string
	originalQuantity string
	symbol           string
	productMaturity  string
	exchange         string
	productType      string
	securityID       string
	side             string
	sideNum          string
	ordType          string
	timeInForce      string
	text             string
	putOrCall        string //for option only
	strikePrice      string //for option only
}

type OrderConfirmation struct {
	id              string
	account         string
	status          string
	reason          string
	symbol          string
	productMaturity string
	exchange        string
	productType     string
	securityID      string
	side            string
	price           string
	quantity        string
	timeInForce     string
	ordType         string
	sideNum         string
	channel         chan OrderConfirmation
	putOrCall       string //for option only
	strikePrice     string //for option only
}

type MarketDataReq struct {
	id              string
	priceType       string
	price           string
	symbol          string
	productMaturity string
	exchange        string
	marketDepth     string

	channel chan MarketDataReq
	status  string
	reason  string
}

type SecurityDefinitionReq struct {
	id           string
	count        int
	channel      chan SecurityDefinitionReq
	status       string
	reason       string
	securityList []Security
}
type Security struct {
	symbol          string
	productMaturity string
	exchange        string
	productType     string
	securityID      string
	securityAltID   string
	putOrCall       string //for option only
	strikePrice     string //for option only
	currency        string

	exchTickSize   float64
	exchPointValue float64
	numTicktblEntries int

	tickSize       float64
	tickValue      float64
}

//********* Concurrent SLice ************************** //

type ConcurrentSlice struct {
	sync.RWMutex
	items []interface{}
}

// Concurrent slice item
type ConcurrentSliceItem struct {
	Index int
	Value interface{}
}

// Appends an item to the concurrent slice
func (cs *ConcurrentSlice) append(item interface{}) {
	cs.Lock()
	defer cs.Unlock()
	cs.items = append(cs.items, item)
}
func (cs *ConcurrentSlice) remove(index int) {
	cs.Lock()
	defer cs.Unlock()
	cs.items = append(cs.items[:index], cs.items[index+1:]...)
}

// Iterates over the items in the concurrent slice
// Each item is sent over a channel, so that
// we can iterate over the slice using the builin range keyword
func (cs *ConcurrentSlice) Iter() <-chan ConcurrentSliceItem {
	c := make(chan ConcurrentSliceItem)

	f := func() {
		cs.Lock()
		defer cs.Unlock()
		for index, value := range cs.items {
			c <- ConcurrentSliceItem{index, value}
		}
		close(c)
	}
	go f()

	return c
}

//********* End Concurrent Slice ************************** //
