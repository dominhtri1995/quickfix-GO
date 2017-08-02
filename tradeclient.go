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
	"golang.org/x/sync/syncmap"
	"sync"
	"github.com/rs/xid"
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

		if (execrefID != "3") { //NOT order Status request

			switch {
			case ordStatus == string(enum.OrdStatus_NEW):
				// New Order
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := newOrderMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "ok"
					order.Account = account
					extractInfoExcecutionReport(order, msg) //Get all the order-related info :trade,Price plapla
					order.channel <- *order                 //Send data back to Channel
					newOrderMap.Delete(clOrdID)             // Remove Order Status request from list
				}
			case ordStatus == string(enum.OrdStatus_REJECTED):
				//rejected
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := newOrderMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "rejected"
					order.Account = account
					order.Reason, _ = msg.Body.GetString(quickfix.Tag(58))
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					newOrderMap.Delete(clOrdID)
				}
			case ordStatus == string(enum.OrdStatus_CANCELED):
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "ok"
					order.Account = account
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					cancelAndUpdateMap.Delete(clOrdID)
				}
			case ordStatus == string(enum.OrdStatus_REPLACED):
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				if order, ok := cancelAndUpdateMap.Load(clOrdID); ok && err == nil {
					order, _ := order.(*OrderConfirmation)
					order.Status = "ok"
					order.Account = account
					extractInfoExcecutionReport(order, msg)
					order.channel <- *order
					cancelAndUpdateMap.Delete(clOrdID)
				}
			case ordStatus == string(enum.OrdStatus_PARTIALLY_FILLED):
				//Partially filled
				//priceFilled,_ := msg.Body.GetInt(quickfix.Tag(31))
				//qtyFilled,_ := msg.Body.GetInt(quickfix.Tag(32)) // qty just got filled
				// 	totalqty,_ := msg.Body.GetString(quickfix.Tag(38))
				// 	Symbol,_ := msg.Body.GetString(quickfix.Tag(55))
				// 	Side , _ := msg.Body.GetString(quickfix.Tag(54))
				// avgPx, _ := msg.Body.GetString(quickfix.Tag(6))

			case ordStatus == string(enum.OrdStatus_FILLED):
				//Fully filled
				//priceFilled,_ := msg.Body.GetInt(quickfix.Tag(31))
				//qtyFilled,_ := msg.Body.GetInt(quickfix.Tag(32)) // qty just got filled
				// 	totalqty,_ := msg.Body.GetString(quickfix.Tag(38))
				// 	Symbol,_ := msg.Body.GetString(quickfix.Tag(55))
				// 	Side , _ := msg.Body.GetString(quickfix.Tag(54))
				// avgPx, _ := msg.Body.GetString(quickfix.Tag(6))

			}
		} else {
			//Order Status request
			account, err := msg.Body.GetString(quickfix.Tag(1))
			numPosReports, _ := msg.Body.GetString(quickfix.Tag(16728))
			for osr := range orderStatusRequestList.Iter() {
				orderStatusRequest, _ := osr.Value.(*OrderStatusReq)
				if err == nil && account == orderStatusRequest.Account { // GET book order not single order Status request
					orderStatusRequest.Count, _ = strconv.Atoi(numPosReports)
					if orderStatusRequest.Count ==0 {
						orderStatusRequest.Status = "ok"
						orderStatusRequest.channel <- *orderStatusRequest
						orderStatusRequestList.remove(osr.Index)
						return
					}

					var order WorkingOrder
					order.OrderID, _ = msg.Body.GetString(quickfix.Tag(37))
					order.Price, _ = msg.Body.GetString(quickfix.Tag(44))
					order.Quantity, _ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
					order.FilledQuantity, _ = msg.Body.GetString(quickfix.Tag(14))
					order.OriginalQuantity, _ = msg.Body.GetString(quickfix.Tag(38))
					order.OrdStatus, _ = msg.Body.GetString(quickfix.Tag(39))
					order.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
					order.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
					order.SideNum, _ = msg.Body.GetString(quickfix.Tag(54))
					order.OrdType, _ = msg.Body.GetString(quickfix.Tag(40))
					order.TimeInForce, _ = msg.Body.GetString(quickfix.Tag(59))
					order.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
					order.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
					order.Text, _ = msg.Body.GetString(quickfix.Tag(58))
					order.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
					putOrCall, _ := msg.Body.GetInt(quickfix.Tag(201))

					if putOrCall == 0 {
						order.PutOrCall = "Put"
					} else {
						order.PutOrCall = "Call"
					}

					if order.SideNum == "1" {
						order.Side = "Buy"
					} else {
						order.Side = "Sell"
					}
					order.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
					if order.ProductType == "FUT" || order.ProductType == "OPT" || order.ProductType == "NRG" {
						order.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
					}
					orderStatusRequest.WorkingOrders = append(orderStatusRequest.WorkingOrders, order)
				}

				if orderStatusRequest.Count == len(orderStatusRequest.WorkingOrders) {
					// Receive all working orders
					fmt.Printf("Receve all working orders : %d for Account %s \n", len(orderStatusRequest.WorkingOrders), orderStatusRequest.Account)
					orderStatusRequest.Status = "ok"
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
			order.Status = "rejected"
			order.Reason, _ = msg.Body.GetString(quickfix.Tag(58))
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
			uan.Count, _ = strconv.Atoi(numPosReports)
			if uan.Count == 0 {
				uan.channel <- *uan // return the result to Channel
				uanMap.Delete(uid)
				return
			}

			//Create a new UAP object
			var uap UAPreport
			uap.Quantity, _ = msg.Body.GetString(quickfix.Tag(32))
			q, _ := strconv.Atoi(uap.Quantity)

			putOrCall, _ := msg.Body.GetInt(quickfix.Tag(201))

			if putOrCall == 0 {
				uap.PutOrCall = "Put"
			} else {
				uap.PutOrCall = "Call"
			}

			if (q > 0) {
				uap.Side = "Buy"
			} else {
				uap.Side = "Sell"
				uap.Quantity = string(q * (-1))
			}
			uap.AccountGroup, _ = msg.Header.GetString(quickfix.Tag(50))
			uap.Account, _ = msg.Body.GetString(quickfix.Tag(1))
			// fmt.Println(uap.AccountGroup)
			uap.Price, _ = msg.Body.GetString(quickfix.Tag(31))
			uap.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
			uap.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
			uap.Product, _ = msg.Body.GetString(quickfix.Tag(55))
			uap.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
			uap.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
			if uap.ProductType == "FUT"  || uap.ProductType == "OPT" || uap.ProductType == "NRG"{
				uap.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
			}
			uan.Reports = append(uan.Reports, uap)

			//UAN Complete
			posReqType, _ := msg.Body.GetString(quickfix.Tag(16724))
			fmt.Println(len(uan.Reports))
			fmt.Println(uan.Count)
			if len(uan.Reports) == uan.Count {
				for j := 0; j < len(uan.Reports); j++ {
					if posReqType == "4" && uan.AccountGroup != uan.Reports[j].AccountGroup {
						uan.Reports = append(uan.Reports[:j], uan.Reports[j+1:]...)
						j--
					}
				}
				fmt.Println("Done UAP")
				uan.channel <- *uan // return the result to Channel
				uanMap.Delete(uid)
			}
		}
	case messageType == "W": //Market data request
		uid, _ := msg.Body.GetString(quickfix.Tag(262))
		if md, ok := marketDataRequestMap.Load(uid); ok {
			md, _ := md.(*MarketDataReq)
			//create market data reqest response
			md.Price, err = msg.Body.GetString(quickfix.Tag(270))
			if ( err != nil) {
				md.Price = "0" //Price not available
				md.Status = "rejected"
				md.Reason = "Price not available"
			} else {
				md.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
				md.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
				productType, _ := msg.Body.GetString(quickfix.Tag(167))
				if (productType == "FUT") {
					md.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
				}

				md.PriceType, _ = msg.Body.GetString(quickfix.Tag(269))
				md.Status = "ok"
			}

			md.channel <- *md
			marketDataRequestMap.Delete(uid)
		}
	case messageType == "Y": //Market data reject
		uid, _ := msg.Body.GetString(quickfix.Tag(262))
		if md, ok := marketDataRequestMap.Load(uid); ok {
			md, _ := md.(*MarketDataReq)
			md.Status = "rejected"
			md.Reason, _ = msg.Body.GetString(quickfix.Tag(58))
			md.channel <- *md
			marketDataRequestMap.Delete(uid)
		}
	case messageType == "d":
		id, _ := msg.Body.GetString(quickfix.Tag(320))
		count,_ := msg.Body.GetInt(quickfix.Tag(393))
		if sdr, ok := securityDefinitionMap.Load(id); ok {
			sdr, _ := sdr.(*SecurityDefinitionReq)
			sdr.Count =count

			var security Security
			security.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
			security.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
			security.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
			security.ProductMaturity,_ = msg.Body.GetString(quickfix.Tag(200))
			security.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
			security.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
			security.PutOrCall, _ = msg.Body.GetString(quickfix.Tag(201))
			security.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))

			exchTickSize, _ := msg.Body.GetString(quickfix.Tag(16552))
			exchPointValue, _ := msg.Body.GetString(quickfix.Tag(16554))
			security.ExchTickSize,_ = strconv.ParseFloat(exchTickSize,64)
			security.ExchPointValue,_  = strconv.ParseFloat(exchPointValue,64)
			security.NumTicktblEntries,_ = msg.Body.GetInt(quickfix.Tag(16456))

			security.TickValue, security.TickSize = calculateTickValueAndSize(security.ExchTickSize,security.ExchPointValue,0,0,"0","0")
			//tag16552 int,tag16554 int,tag16456 int, tag16457 int,tag16458 string
			sdr.SecurityList = append(sdr.SecurityList, security)
			fmt.Println(sdr.Count)
			fmt.Println(len(sdr.SecurityList))
			if sdr.Count == len(sdr.SecurityList){
				fmt.Println("receive all security definition")
				sdr.Status = "ok"
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

// Wrapper fro Query function in console.go, using Channel to wait for all message come through.
/*
/* Use these function to avoid using Channel in main code
/*	Sepearte TT from mistro code
/*  For easy debug and maintain code with TT
 */
func TT_PAndLSOD(id string, account string, accountGroup string, sender string) (uan UAN) {
	c := make(chan UAN)
	QueryPAndLSOD(id, accountGroup, sender, c) // Get SOD report for position upto today



	select {
	case uan= <-c:

	case <-getTimeOutChan():
		var uan UAN
		uan.Status = "rejected"
		uan.Reason = "time out"
	}

	var uan1 UAN
	c1 := make(chan UAN)
	QueryPAndLPos(xid.New().String(), account, sender, c1) // Get Position report for positions placed today
	select {
	case uan1 = <-c1:
		if uan.Status != "rejected"{
			uan.Reports = append(uan.Reports, uan1.Reports[0:]...)
			uan.Count = len(uan.Reports)
			uan.Account = uan1.Account
			uan.Status = "ok"
		}
	case <-getTimeOutChan():
		fmt.Println("Time out Position")
		var uan UAN
		uan.Status = "rejected"
		uan.Reason = "time out"
	}

	return uan
}
func TT_NewOrderSingle(id string, account string, side string, ordType string, quantity string, limitPri string, stopPri string, symbol string, exchange string, maturity string, productType string, timeInForce string, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryNewOrderSingle(id, account, side, ordType, quantity, limitPri, stopPri, symbol, exchange, maturity, productType, timeInForce, sender, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
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
		wo.Status = "rejected"
		wo.Reason = "time out"
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
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
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
		ordStatus.Status = "rejected"
		ordStatus.Reason = "time out"
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
		mdr.Status = "rejected"
		mdr.Reason = "time out"
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
		sdr.Status = "rejected"
		sdr.Reason = "time out or no security found"
	}

	return sdr
}

func getTimeOutChan() chan bool {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(60 * time.Second)
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
		//p := float64(Price)
		//maxPirce := float64(tag16458)
		//for i := 0; i < tag16456; i++ {
		//	if p < maxPirce {
		//		TickSize = baseTickSize * tag16457
		//		TickValue = TickSize * tag16554
		//		break
		//	}
		//}
	}

	return tickValue, tickSize
}

func extractInfoExcecutionReport(order *OrderConfirmation, msg quickfix.Message) {
	order.Price, _ = msg.Body.GetString(quickfix.Tag(44))     //not available for market order
	order.Quantity, _ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
	order.Symbol, _ = msg.Body.GetString(quickfix.Tag(55))
	order.Exchange, _ = msg.Body.GetString(quickfix.Tag(207))
	order.ProductType, _ = msg.Body.GetString(quickfix.Tag(167))
	order.OrdType, _ = msg.Body.GetString(quickfix.Tag(40))
	order.SideNum, _ = msg.Body.GetString(quickfix.Tag(54))
	order.TimeInForce, _ = msg.Body.GetString(quickfix.Tag(59))
	order.SecurityID, _ = msg.Body.GetString(quickfix.Tag(48))
	order.SecurityAltID, _ = msg.Body.GetString(quickfix.Tag(10455))
	order.StrikePrice, _ = msg.Body.GetString(quickfix.Tag(202))
	putOrCall, _ := msg.Body.GetInt(quickfix.Tag(201))

	if putOrCall == 0 {
		order.PutOrCall = "Put"
	} else {
		order.PutOrCall = "Call"
	}
	if order.SideNum == "1" {
		order.Side = "Buy"
	} else {
		order.Side = "Sell"
	}
	if order.ProductType == "FUT" || order.ProductType == "OPT" || order.ProductType == "NRG" {
		order.ProductMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
	}
}

type TradeClient struct {
}
type UAN struct {
	Id           string
	Account      string
	AccountGroup string
	Count        int
	channel      chan UAN
	Reports      []UAPreport
	Status       string
	Reason       string
}
type UAPreport struct {
	Id              string
	AccountGroup    string
	Account         string
	Quantity        string
	Price           string
	Side            string
	Product         string
	ProductType     string
	ProductMaturity string
	Exchange        string
	SecurityID      string
	SecurityAltID   string
	PutOrCall       string //for option only
	StrikePrice     string //for option only
}

type OrderStatusReq struct {
	Account       string
	Count         int
	channel       chan OrderStatusReq
	WorkingOrders []WorkingOrder
	Status        string
	Reason        string
}
type WorkingOrder struct {
	OrderID          string // Used to cancel order or request order Status later
	Price            string
	OrdStatus        string
	Quantity         string
	FilledQuantity   string
	OriginalQuantity string
	Symbol           string
	ProductMaturity  string
	Exchange         string
	ProductType      string
	SecurityID       string
	SecurityAltID   string
	Side             string
	SideNum          string
	OrdType          string
	TimeInForce      string
	Text             string
	PutOrCall        string //for option only
	StrikePrice      string //for option only
}

type OrderConfirmation struct {
	Id              string
	Account         string
	Status          string
	Reason          string
	Symbol          string
	ProductMaturity string
	Exchange        string
	ProductType     string
	SecurityID      string
	SecurityAltID   string
	Side            string
	Price           string
	Quantity        string
	TimeInForce     string
	OrdType         string
	SideNum         string
	channel         chan OrderConfirmation
	PutOrCall       string //for option only
	StrikePrice     string //for option only
}

type MarketDataReq struct {
	Id              string
	PriceType       string
	Price           string
	Symbol          string
	ProductMaturity string
	Exchange        string
	MarketDepth     string

	channel chan MarketDataReq
	Status  string
	Reason  string
}

type SecurityDefinitionReq struct {
	Id           string
	Count        int
	channel      chan SecurityDefinitionReq
	Status       string
	Reason       string
	SecurityList []Security
}
type Security struct {
	Symbol          string
	ProductMaturity string
	Exchange        string
	ProductType     string
	SecurityID      string
	SecurityAltID   string
	PutOrCall       string //for option only
	StrikePrice     string //for option only
	Currency        string

	ExchTickSize      float64
	ExchPointValue    float64
	NumTicktblEntries int

	TickSize  float64
	TickValue float64
}
type TTNotification struct {
	Id string //unique for this fill
	Account string
	SubAccount string
	OrderID string
	QtyFilled string
	PriceFilled string
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
// Each item is sent over a Channel, so that
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
