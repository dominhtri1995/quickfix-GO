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

	if (messageType == "A" ) {
		msg.Header.SetString(quickfix.Tag(96), "12345678")
		msg.Header.SetInt(quickfix.Tag(95), 8)
		t34, err := msg.Header.GetInt(quickfix.Tag(34))
		if (t34 == 1 && err == nil && sessionID.TargetCompID == "TTDEV18O") {
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
				for i, order := range NewOrders {
					if (err == nil && order.id == clOrdID ) {
						order.status = "ok"
						order.account = account
						order.channel <- order                                //Send data back to channel
						NewOrders = append(NewOrders[:i], NewOrders[i+1:]...) // Remove Order status request from list
					}
				}
			case ordStatus == string(enum.OrdStatus_REJECTED):
				//rejected
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				for i, order := range NewOrders {
					if (err == nil && order.id == clOrdID ) {
						order.status = "rejected"
						order.account = account
						order.reason, _ = msg.Body.GetString(quickfix.Tag(58))
						order.channel <- order
						NewOrders = append(NewOrders[:i], NewOrders[i+1:]...)
					}
				}
			case ordStatus == string(enum.OrdStatus_CANCELED):
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				for i, order := range CancelAndUpdateOrders {
					if (err == nil && order.id == clOrdID ) {
						order.status = "ok"
						order.account = account
						order.channel <- order
						CancelAndUpdateOrders = append(CancelAndUpdateOrders[:i], CancelAndUpdateOrders[i+1:]...)
					}
				}
			case ordStatus == string(enum.OrdStatus_REPLACED):
				clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
				for i, order := range CancelAndUpdateOrders {
					if (err == nil && order.id == clOrdID ) {
						order.status = "ok"
						order.account = account
						order.channel <- order
						CancelAndUpdateOrders = append(CancelAndUpdateOrders[:i], CancelAndUpdateOrders[i+1:]...)
					}
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
			for i := range OSRs {
				if (err == nil && account == OSRs[i].account) { // GET book order not single order status request
					OSRs[i].count, _ = strconv.Atoi(numPosReports)

					var order WorkingOrder
					order.orderID, _ = msg.Body.GetString(quickfix.Tag(37))
					order.price, _ = msg.Body.GetString(quickfix.Tag(44))
					order.quantity, _ = msg.Body.GetString(quickfix.Tag(151)) // leaves qty
					order.ordStatus, _ = msg.Body.GetString(quickfix.Tag(39))
					order.symbol, _ = msg.Body.GetString(quickfix.Tag(55))
					order.exchange,_ = msg.Body.GetString(quickfix.Tag(207))
					order.side, _ = msg.Body.GetString(quickfix.Tag(54))
					order.sideNum, _ = msg.Body.GetString(quickfix.Tag(54))
					if order.sideNum == "1" {
						order.side = "buy"
					} else {
						order.side = "sell"
					}
					order.productType, _ = msg.Body.GetString(quickfix.Tag(167))
					if (order.productType == "FUT") {
						order.productMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
					}
					OSRs[i].workingOrders = append(OSRs[i].workingOrders, order)
				}

				if (OSRs[i].count == len(OSRs[i].workingOrders)) {
					// Receive all working orders
					fmt.Printf("Receve all working orders : %d for account %s \n", len(OSRs[i].workingOrders), OSRs[i].account)
					OSRs[i].channel <- OSRs[i]
					OSRs = append(OSRs[:i], OSRs[i+1:]...)
					break
				}
			}
		}
	case messageType == string(enum.MsgType_ORDER_CANCEL_REJECT):
		clOrdID, err := msg.Body.GetString(quickfix.Tag(11))
		for i, order := range CancelAndUpdateOrders {
			if (err == nil && order.id == clOrdID ) {
				order.status = "rejected"
				order.reason, _ = msg.Body.GetString(quickfix.Tag(58))
				order.channel <- order
				CancelAndUpdateOrders = append(CancelAndUpdateOrders[:i], CancelAndUpdateOrders[i+1:]...)
			}
		}
		/* UAP - UAP - UAP - UAP - UAP - UAP - UAP */
	case messageType == "UAP":
		// fmt.Printf("Receiving UAP \n")
		var uid, _ = msg.Body.GetString(quickfix.Tag(16710))
		for i := range UANs {
			if (uid == UANs[i].id) {
				numPosReports, _ := msg.Body.GetString(quickfix.Tag(16727))
				UANs[i].count, _ = strconv.Atoi(numPosReports)

				//Create a new UAP object
				var uap UAPreport
				uap.quantity, _ = msg.Body.GetString(quickfix.Tag(32))
				q, _ := strconv.Atoi(uap.quantity)

				if (q > 0) {
					uap.side = "long"
				} else {
					uap.side = "short"
					uap.quantity = string(q * (-1))
				}
				uap.accountGroup, _ = msg.Header.GetString(quickfix.Tag(50))
				// fmt.Println(uap.accountGroup)
				uap.price, _ = msg.Body.GetString(quickfix.Tag(31))
				uap.product, _ = msg.Body.GetString(quickfix.Tag(55))
				productType, _ := msg.Body.GetString(quickfix.Tag(167))
				if (productType == "FUT") {
					uap.productMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
				}
				UANs[i].reports = append(UANs[i].reports, uap)

				//UAN Complete
				if (len(UANs[i].reports) == UANs[i].count) {
					for j := 0; j < len(UANs[i].reports); j++ {
						if UANs[i].accountGroup != UANs[i].reports[j].accountGroup {
							UANs[i].reports = append(UANs[i].reports[:j], UANs[i].reports[j+1:]...)
							j--
						}
					}
					fmt.Printf("Number of positions :%d for trader %s \n", len(UANs[i].reports), UANs[i].accountGroup)
					UANs[i].channel <- UANs[i] // return the result to channel
					UANs = append(UANs[:i], UANs[i+1:]...)
				}
				break
			}
		}
	case messageType == "W": //Market data request
		uid, _ := msg.Body.GetString(quickfix.Tag(262))
		for i, md := range MarketDataRequests {
			if uid == md.id {

				//create market data reqest response
				md.price, err = msg.Body.GetString(quickfix.Tag(270))
				if ( err != nil) {
					md.price = "0" //Price not available
					md.status= "rejected"
					md.reason = "price not available"
				}else{
					md.symbol, _ = msg.Body.GetString(quickfix.Tag(55))
					md.exchange, _ = msg.Body.GetString(quickfix.Tag(207))
					productType, _ := msg.Body.GetString(quickfix.Tag(167))
					if (productType == "FUT") {
						md.productMaturity, _ = msg.Body.GetString(quickfix.Tag(200))
					}

					md.priceType, _ = msg.Body.GetString(quickfix.Tag(269))
					md.status = "ok"
				}

				md.channel <- md
				MarketDataRequests = append(MarketDataRequests[:i], MarketDataRequests[i+1:]...)
				break
			}
		}
	case messageType == "Y": //Market data reject
		uid, _ := msg.Body.GetString(quickfix.Tag(262))
		for i, md := range MarketDataRequests {
			if uid == md.id {
				md.status = "rejected"
				md.reason, _ = msg.Body.GetString(quickfix.Tag(58))
				md.channel <- md
				MarketDataRequests = append(MarketDataRequests[:i], MarketDataRequests[i+1:]...)
			}
		}
	}
	return
}

var UANs []UAN
var OSRs []OrderStatusReq
var NewOrders []OrderConfirmation
var CancelAndUpdateOrders []OrderConfirmation
var MarketDataRequests []MarketDataReq

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
func TT_PAndLSOD(id string, account string, accountGroup string,sender string) (uan UAN) {
	c := make(chan UAN)
	QueryPAndLSOD(id, accountGroup, c)

	intradayID := xid.New().String()
	c1 := make (chan UAN)
	QueryPAndLPos(intradayID,account,c1)
	select {
	case uan= <-c:
		return uan
	case <-getTimeOutChan():
		var uan UAN
		uan.status = "rejected"
		uan.reason = "time out"
	}
	return uan
}
func TT_NewOrderSingle(id string, account string, side enum.Side, ordType string, quantity string, pri string, symbol string, exchange string, maturity string, productType enum.SecurityType, sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryNewOrderSingle(id, account, side, ordType, quantity, pri, symbol, exchange, maturity, productType, c)
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.status = "rejected"
		ordStatus.reason = "time out"
	}
	return ordStatus
}

func TT_WorkingOrder(account string,sender string) (wo OrderStatusReq) {
	c := make(chan OrderStatusReq)
	QueryWorkingOrder(account, c)
	select {
	case wo = <-c:
		return wo
	case <-getTimeOutChan():
		wo.status = "rejected"
		wo.reason = "time out"
	}
	return wo
}
func TT_OrderCancel(id string, orderID string,sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryOrderCancel(id, orderID, c) // Cancel the first working order
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.status = "rejected"
		ordStatus.reason = "time out"
	}
	return ordStatus
}
func TT_OrderCancelReplace(orderID string, newid string, account string, side enum.Side, ordType enum.OrdType, quantity string, pri string, symbol string, exchange string, maturity string, productType enum.SecurityType,sender string) (ordStatus OrderConfirmation) {
	c := make(chan OrderConfirmation)
	QueryOrderCancelReplace(orderID, newid, account, side, ordType, quantity, pri, symbol, exchange, maturity, productType, c) // Replace the first working order
	select {
	case ordStatus = <-c:
		return ordStatus
	case <-getTimeOutChan():
		ordStatus.status = "rejected"
		ordStatus.reason = "time out"
	}
	return ordStatus
}

func TT_MarketDataRequest(id string, requestType enum.SubscriptionRequestType, marketDepth int, priceType enum.MDEntryType, symbol string, exchange string, maturity string, productType enum.SecurityType,sender string) (mdr MarketDataReq) {
	c := make(chan MarketDataReq)
	QueryMarketDataRequest(id, requestType, marketDepth, priceType, symbol, exchange, maturity, productType, c)
	select {
	case mdr = <-c:
		return mdr
	case <-getTimeOutChan():
		mdr.status = "rejected"
		mdr.reason = "time out"
	}
	return mdr
}
func getTimeOutChan() chan bool {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		timeout <- true
	}()
	return timeout
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
	quantity        string
	price           string
	side            string
	product         string
	productMaturity string
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
	orderID         string // Used to cancel order or request order status later
	price           string
	ordStatus       string
	quantity        string
	symbol          string
	productMaturity string
	exchange 		string
	productType     string
	side            string
	sideNum			string
	ordType 		string
}

type OrderConfirmation struct {
	id      string
	account string
	status  string
	reason  string
	channel chan OrderConfirmation
}

type MarketDataReq struct {
	id              string
	priceType       string
	price           string
	symbol          string
	productMaturity string
	exchange        string
	channel         chan MarketDataReq
	status          string
	reason          string
}
